package linebotservice

import (
	"log"
	"net/http"

	"github.com/Zhima-Mochi/go-linebot-service/messageservice/messagecorefactory"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type MessageService interface {
	SetDefaultMessageCore(messageCore messagecorefactory.MessageCore)
	GetDefaultMessageCore() messagecorefactory.MessageCore
	SetCustomMessageTypeCore(messageType linebot.MessageType, messageCore messagecorefactory.MessageCore)
	GetCustomMessageTypeCore(messageType linebot.MessageType) messagecorefactory.MessageCore
	ClearCustomMessageTypeCore(messageType linebot.MessageType)
	ClearAllCustomMessageTypeCore()
	Process(event *linebot.Event) (linebot.SendingMessage, error)
}

type LineBotService struct {
	LineBotClient  *linebot.Client
	MessageService MessageService
}

func NewLineBotService(lineBotClient *linebot.Client, messageService MessageService) *LineBotService {
	return &LineBotService{
		LineBotClient:  lineBotClient,
		MessageService: messageService,
	}
}

func (l *LineBotService) Do(w http.ResponseWriter, req *http.Request) {
	events, err := l.LineBotClient.ParseRequest(req)
	if err != nil {
		log.Print(err)
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			message, err := l.MessageService.Process(event)
			if err != nil {
				log.Print(err)
			}
			if _, err := l.LineBotClient.ReplyMessage(event.ReplyToken, message).Do(); err != nil {
				log.Print(err)
			}
		}
	}
}
