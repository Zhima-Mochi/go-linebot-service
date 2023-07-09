package chatgpt

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/Zhima-Mochi/go-linebot-service/messageservice/messagecorefactory"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/sashabaranov/go-openai"
)

var (
	defaultMessageCore = MessageCore{
		memory:          &localMemory{},
		memoryN:         3,
		chatModel:       openai.GPT3Dot5Turbo0613,
		chatToken:       150,
		chatTemperature: 0.9,
		audioModel:      openai.Whisper1,
	}
)

type Memory interface {
	Remember(message openai.ChatCompletionMessage)
	// Recall the last n messages
	Recall(n int) []openai.ChatCompletionMessage
	// Revoke the last n messages
	Revoke(n int) []openai.ChatCompletionMessage
}

type localMemory struct {
	messages []openai.ChatCompletionMessage
}

func (l *localMemory) Remember(message openai.ChatCompletionMessage) {
	l.messages = append(l.messages, message)
}

func (l *localMemory) Recall(n int) []openai.ChatCompletionMessage {
	if n > len(l.messages) {
		n = len(l.messages)
	}
	return l.messages[len(l.messages)-n:]
}

func (l *localMemory) Revoke(n int) []openai.ChatCompletionMessage {
	if n > len(l.messages) {
		n = len(l.messages)
	}
	messages := l.messages[len(l.messages)-n:]
	l.messages = l.messages[:len(l.messages)-n]
	return messages
}

type MessageCore struct {
	client          *openai.Client
	memory          Memory
	memoryN         int
	chatModel       string
	chatToken       int
	chatTemperature float32
	audioModel      string
}

func NewMessageCore(client *openai.Client, options ...WithOptoin) *MessageCore {
	core := defaultMessageCore
	core.client = client
	for _, option := range options {
		option(&core)
	}
	return &core
}

func (m *MessageCore) Process(event *linebot.Event) (linebot.SendingMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	userMessage := ""
	replyText := ""
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		userMessage = message.Text
	case *linebot.AudioMessage:
		text, err := m.ConvertAudioToText(ctx, message.OriginalContentURL)
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

	botResponse, err := m.Chat(ctx, userMessage)
	if err != nil {
		return nil, err
	}
	replyText += botResponse

	return linebot.NewTextMessage(replyText), nil
}

func (m *MessageCore) Chat(ctx context.Context, message string) (string, error) {
	newMessage := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	}
	m.memory.Remember(newMessage)
	messages := m.memory.Recall(m.memoryN)
	resp, err := m.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       m.chatModel,
		Messages:    messages,
		MaxTokens:   m.chatToken,
		Temperature: m.chatTemperature,
		Stop:        []string{"\n"},
	})
	if err != nil {
		return "", err
	}
	replyMessage := resp.Choices[0].Message
	m.memory.Remember(replyMessage)
	return replyMessage.Content, nil
}

func (m *MessageCore) ConvertAudioToText(ctx context.Context, audioURL string) (string, error) {
	audioReader, err := downloadAudio(audioURL)
	if err != nil {
		return "", err
	}

	req := openai.AudioRequest{
		Model:  openai.Whisper1,
		Reader: audioReader,
	}

	resp, err := m.client.CreateTranscription(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func downloadAudio(audioURL string) (io.Reader, error) {
	response, err := http.Get(audioURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, messagecorefactory.ErrorAudioDownloadFailed
	}

	return response.Body, nil
}
