package linebotservice

import (
	"log"
	"net/http"

	"github.com/Zhima-Mochi/go-linebot-service/messageservice"
	messageserviceFactory "github.com/Zhima-Mochi/go-linebot-service/messageservice/factory"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type MessageService interface {
	SetDefaultMessageCore(messageCore messageserviceFactory.MessageCore)
	GetDefaultMessageCore() messageserviceFactory.MessageCore
	SetCustomMessageTypeCore(messageType linebot.MessageType, messageCore messageserviceFactory.MessageCore)
	GetCustomMessageTypeCore(messageType linebot.MessageType) messageserviceFactory.MessageCore
	Process(message linebot.Message) (linebot.SendingMessage, error)
}

type LineBotService struct {
	LineBotClient  *linebot.Client
	MessageService MessageService
}

func NewLineBotService(lineBotClient *linebot.Client) *LineBotService {
	return &LineBotService{
		LineBotClient:  lineBotClient,
		MessageService: messageservice.NewMessageService(),
	}
}

func (l *LineBotService) Do(w http.ResponseWriter, req *http.Request) {
	events, err := l.LineBotClient.ParseRequest(req)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			message, err := l.MessageService.Process(event.Message)
			if err != nil {
				log.Print(err)
			}
			if _, err := l.LineBotClient.ReplyMessage(event.ReplyToken, message).Do(); err != nil {
				log.Print(err)
			}
		}
	}
}
