package factory

import (
	"errors"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

var (
	ErrorMessageTypeNotSupported = errors.New("MessageType not supported")
)

type MessageCore interface {
	Process(message linebot.Message) (linebot.SendingMessage, error)
}
