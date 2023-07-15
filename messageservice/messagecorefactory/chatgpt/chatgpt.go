package chatgpt

import (
	"context"

	"github.com/Zhima-Mochi/go-linebot-service/messageservice/messagecorefactory"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/sashabaranov/go-openai"
)

var (
	defaultMessageCore = MessageCore{
		memory: &localMemory{
			messages: make(map[string][]openai.ChatCompletionMessage),
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

type localMemory struct {
	messages map[string][]openai.ChatCompletionMessage
}

func (l *localMemory) Remember(ctx context.Context, userID string, message openai.ChatCompletionMessage) error {
	if l.messages == nil {
		l.messages = make(map[string][]openai.ChatCompletionMessage)
	}
	l.messages[userID] = append(l.messages[userID], message)
	return nil
}

func (l *localMemory) Recall(ctx context.Context, userID string, n int) ([]openai.ChatCompletionMessage, error) {
	if l.messages == nil {
		l.messages = make(map[string][]openai.ChatCompletionMessage)
	}
	if n > len(l.messages[userID]) {
		n = len(l.messages[userID])
	}
	return l.messages[userID][len(l.messages[userID])-n:], nil
}

func (l *localMemory) Revoke(ctx context.Context, userID string, n int) ([]openai.ChatCompletionMessage, error) {
	if l.messages == nil {
		l.messages = make(map[string][]openai.ChatCompletionMessage)
	}
	if n > len(l.messages[userID]) {
		n = len(l.messages[userID])
	}
	revokeMessages := l.messages[userID][len(l.messages[userID])-n:]
	l.messages[userID] = l.messages[userID][:len(l.messages[userID])-n]
	return revokeMessages, nil
}

func (l *localMemory) Forget(ctx context.Context, userID string, n int) error {
	if l.messages == nil {
		l.messages = make(map[string][]openai.ChatCompletionMessage)
	}
	if n > len(l.messages[userID]) {
		n = len(l.messages)
	}
	l.messages[userID] = l.messages[userID][n:]
	return nil
}

func (l *localMemory) GetSize(ctx context.Context, userID string) (int, error) {
	if l.messages == nil {
		l.messages = make(map[string][]openai.ChatCompletionMessage)
	}
	return len(l.messages[userID]), nil
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
		text, err := m.convertAudioToText(ctx, message.OriginalContentURL)
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
	newMessage := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	}
	err := m.memory.Remember(ctx, userID, newMessage)
	if err != nil {
		return "", err
	}
	messages, err := m.memory.Recall(ctx, userID, m.memoryN)
	if err != nil {
		return "", err
	}
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
		return "", err
	}

	req := openai.AudioRequest{
		Model:  openai.Whisper1,
		Reader: callResp.Content,
	}

	transResp, err := m.openaiClient.CreateTranscription(ctx, req)
	if err != nil {
		return "", err
	}
	return transResp.Text, nil
}
