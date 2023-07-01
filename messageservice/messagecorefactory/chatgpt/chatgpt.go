package chatgpt

import (
	"github.com/Zhima-Mochi/go-linebot-service/messageservice/messagecorefactory"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/sashabaranov/go-openai"
)

type Memory interface {
	Remember(message *openai.ChatCompletionMessage)
	// Recall the last n messages
	Recall(n int) []*openai.ChatCompletionMessage
	// Revoke the last n messages
	Revoke(n int) []*openai.ChatCompletionMessage
}

type localMemory struct {
	messages []*openai.ChatCompletionMessage
}

func (l *localMemory) Remember(message *openai.ChatCompletionMessage) {
	l.messages = append(l.messages, message)
}

func (l *localMemory) Recall(n int) []*openai.ChatCompletionMessage {
	if n > len(l.messages) {
		n = len(l.messages)
	}
	return l.messages[len(l.messages)-n:]
}

func (l *localMemory) Revoke(n int) []*openai.ChatCompletionMessage {
	if n > len(l.messages) {
		n = len(l.messages)
	}
	messages := l.messages[len(l.messages)-n:]
	l.messages = l.messages[:len(l.messages)-n]
	return messages
}

type MessageCore struct {
	client  *openai.Client
	memory  Memory
	memoryN int
}

func NewMessageCore(client *openai.Client, options ...WithOptoin) *MessageCore {
	core := &MessageCore{
		client:  client,
		memory:  &localMemory{},
		memoryN: 3,
	}
	for _, option := range options {
		option(core)
	}
	return core
}

func (m *MessageCore) Process(message linebot.Message) (linebot.SendingMessage, error) {
	switch message := message.(type) {
	case *linebot.TextMessage:
		replyText := message.Text
		return linebot.NewTextMessage(replyText), nil
	case *linebot.ImageMessage:
		return linebot.NewImageMessage(message.OriginalContentURL, message.PreviewImageURL), nil
	case *linebot.AudioMessage:
		return linebot.NewAudioMessage(message.OriginalContentURL, message.Duration), nil
	case *linebot.TemplateMessage:
		return linebot.NewTemplateMessage(message.AltText, message.Template), nil
	case *linebot.FlexMessage:
		return linebot.NewFlexMessage(message.AltText, message.Contents), nil
	}
	return nil, messagecorefactory.ErrorMessageTypeNotSupported
}
