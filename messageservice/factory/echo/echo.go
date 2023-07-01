package echo

import (
	"github.com/Zhima-Mochi/go-linebot-service/messageservice/factory"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type MessageCore struct {
}

func (m *MessageCore) Process(message linebot.Message) (linebot.Message, error) {
	switch message := message.(type) {
	case *linebot.TextMessage:
		replyText := message.Text
		return linebot.NewTextMessage(replyText), nil
	}
	return nil, factory.ErrorMessageTypeNotSupported
}
