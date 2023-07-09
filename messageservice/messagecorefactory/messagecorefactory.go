package messagecorefactory

import (
	"context"
	"errors"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

var (
	ErrorMessageTypeNotSupported = errors.New("MessageType not supported")

	ErrorAudioDownloadFailed = errors.New("audio download failed")
)

type MessageCore interface {
	Process(ctx context.Context, event *linebot.Event) (linebot.SendingMessage, error)
}
