package echo

import (
	"context"
	"fmt"

	"github.com/Zhima-Mochi/go-linebot-service/messageservice/messagecorefactory"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type MessageCore struct {
}

func (m *MessageCore) Process(ctx context.Context, event *linebot.Event) (linebot.SendingMessage, error) {
	switch message := event.Message.(type) {
	case *linebot.FileMessage:
		return linebot.NewTextMessage(fmt.Sprintf("FileMessage: %s (%d byte)", message.FileName, message.FileSize)), nil
	case *linebot.TextMessage:
		replyText := message.Text
		return linebot.NewTextMessage(replyText), nil
	case *linebot.ImageMessage:
		return linebot.NewImageMessage(message.OriginalContentURL, message.PreviewImageURL), nil
	case *linebot.VideoMessage:
		return linebot.NewVideoMessage(message.OriginalContentURL, message.PreviewImageURL), nil
	case *linebot.AudioMessage:
		return linebot.NewAudioMessage(message.OriginalContentURL, message.Duration), nil
	case *linebot.LocationMessage:
		return linebot.NewLocationMessage(message.Title, message.Address, message.Latitude, message.Longitude), nil
	case *linebot.StickerMessage:
		return linebot.NewStickerMessage(message.PackageID, message.StickerID), nil
	case *linebot.ImagemapMessage:
		return linebot.NewImagemapMessage(message.BaseURL, message.AltText, message.BaseSize, message.Actions...), nil
	case *linebot.TemplateMessage:
		return linebot.NewTemplateMessage(message.AltText, message.Template), nil
	case *linebot.FlexMessage:
		return linebot.NewFlexMessage(message.AltText, message.Contents), nil
	}
	return nil, messagecorefactory.ErrorMessageTypeNotSupported
}

func NewMessageCore() *MessageCore {
	return &MessageCore{}
}
