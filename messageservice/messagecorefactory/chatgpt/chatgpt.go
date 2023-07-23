package chatgpt

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Zhima-Mochi/go-linebot-service/messageservice/messagecorefactory"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/sashabaranov/go-openai"
)

var (
	defaultMessageCore = MessageCore{
		memory: &LocalMemory{
			Messages: make(map[string][]openai.ChatCompletionMessage),
		},
		memoryN:         10,
		chatModel:       openai.GPT3Dot5Turbo0613,
		chatToken:       150,
		chatTemperature: 0.9,
		audioModel:      openai.Whisper1,
	}
)

type Memory interface {
	// Remember a message
	Remember(ctx context.Context, userID string, message openai.ChatCompletionMessage) error
	// Recall the last n messages
	Recall(ctx context.Context, userID string, n int) ([]openai.ChatCompletionMessage, error)
	// Revoke the last n messages
	Revoke(ctx context.Context, userID string, n int) ([]openai.ChatCompletionMessage, error)
	// Forget the first n messages
	Forget(ctx context.Context, userID string, n int) error
	// GetSize returns the number of messages stored for a user
	GetSize(ctx context.Context, userID string) (int, error)
}

type LocalMemory struct {
	Messages map[string][]openai.ChatCompletionMessage
}

func (l *LocalMemory) Remember(ctx context.Context, userID string, message openai.ChatCompletionMessage) error {
	if l.Messages == nil {
		l.Messages = make(map[string][]openai.ChatCompletionMessage)
	}
	l.Messages[userID] = append(l.Messages[userID], message)
	return nil
}

func (l *LocalMemory) Recall(ctx context.Context, userID string, n int) ([]openai.ChatCompletionMessage, error) {
	if l.Messages == nil {
		l.Messages = make(map[string][]openai.ChatCompletionMessage)
	}
	if n > len(l.Messages[userID]) {
		n = len(l.Messages[userID])
	}
	return l.Messages[userID][len(l.Messages[userID])-n:], nil
}

func (l *LocalMemory) Revoke(ctx context.Context, userID string, n int) ([]openai.ChatCompletionMessage, error) {
	if l.Messages == nil {
		l.Messages = make(map[string][]openai.ChatCompletionMessage)
	}
	if n > len(l.Messages[userID]) {
		n = len(l.Messages[userID])
	}
	revokeMessages := l.Messages[userID][len(l.Messages[userID])-n:]
	l.Messages[userID] = l.Messages[userID][:len(l.Messages[userID])-n]
	return revokeMessages, nil
}

func (l *LocalMemory) Forget(ctx context.Context, userID string, n int) error {
	if l.Messages == nil {
		l.Messages = make(map[string][]openai.ChatCompletionMessage)
	}
	if n > len(l.Messages[userID]) {
		n = len(l.Messages)
	}
	l.Messages[userID] = l.Messages[userID][n:]
	return nil
}

func (l *LocalMemory) GetSize(ctx context.Context, userID string) (int, error) {
	if l.Messages == nil {
		l.Messages = make(map[string][]openai.ChatCompletionMessage)
	}
	return len(l.Messages[userID]), nil
}

type MessageCore struct {
	openaiClient    *openai.Client
	linebotClient   *linebot.Client
	memory          Memory
	memoryN         int
	chatModel       string
	chatToken       int
	chatTemperature float32
	audioModel      string
	systemMessage   string
}

func NewMessageCore(openaiClient *openai.Client, linebotClient *linebot.Client, options ...WithOption) *MessageCore {
	core := defaultMessageCore
	core.openaiClient = openaiClient
	core.linebotClient = linebotClient

	for _, option := range options {
		option(&core)
	}
	return &core
}

func (m *MessageCore) Process(ctx context.Context, event *linebot.Event) (linebot.SendingMessage, error) {
	userMessage := ""
	replyText := ""
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		userMessage = message.Text
	case *linebot.AudioMessage:
		text, err := m.convertAudioToText(ctx, message.ID)
		if err != nil {
			return nil, err
		}
		userMessage = text
		replyText += "🎤: " + text + "\n"
	default:
		return nil, messagecorefactory.ErrorMessageTypeNotSupported
	}
	if userMessage == "" {
		return linebot.NewTextMessage(""), nil
	}

	botResponse, err := m.chat(ctx, event.Source.UserID, userMessage)
	if err != nil {
		return nil, err
	}
	replyText += botResponse

	return linebot.NewTextMessage(replyText), nil
}

func (m *MessageCore) chat(ctx context.Context, userID, message string) (string, error) {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: m.systemMessage,
		},
	}

	newMessage := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	}
	err := m.memory.Remember(ctx, userID, newMessage)
	if err != nil {
		return "", err
	}

	history, err := m.memory.Recall(ctx, userID, m.memoryN)
	if err != nil {
		return "", err
	}
	messages = append(messages, history...)

	resp, err := m.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       m.chatModel,
		Messages:    messages,
		MaxTokens:   m.chatToken,
		Temperature: m.chatTemperature,
	})
	if err != nil {
		return "", err
	}
	replyMessage := resp.Choices[0].Message
	err = m.memory.Remember(ctx, userID, replyMessage)
	if err != nil {
		return "", err
	}

	if len(messages)+1 > m.memoryN {
		err = m.memory.Forget(ctx, userID, len(messages)-m.memoryN)
		if err != nil {
			return "", err
		}
	}
	return replyMessage.Content, nil
}

func (m *MessageCore) convertAudioToText(ctx context.Context, messageID string) (string, error) {
	call := m.linebotClient.GetMessageContent(messageID)
	callResp, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("failed to get message content: %w", err)
	}

	// Read the content
	content, err := ioutil.ReadAll(callResp.Content)
	if err != nil {
		return "", fmt.Errorf("failed to read message content: %w", err)
	}

	// Create a file
	file, err := os.Create(messageID + ".m4a")
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write the content to the file
	_, err = file.Write(content)
	if err != nil {
		return "", fmt.Errorf("failed to write content to file: %w", err)
	}

	req := openai.AudioRequest{
		Model:  openai.Whisper1,
		Reader: file,
	}
	transResp, err := m.openaiClient.CreateTranscription(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create transcription: %w", err)
	}

	// Delete the file
	err = os.Remove(messageID + ".m4a")
	if err != nil {
		return "", fmt.Errorf("failed to delete file: %w", err)
	}

	return transResp.Text, nil
}
