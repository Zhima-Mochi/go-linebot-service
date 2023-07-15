package linebotservice

import (
	"context"
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
	Process(ctx context.Context, event *linebot.Event) (linebot.SendingMessage, error)
}

type LineBotService struct {
	LineBotClient  *linebot.Client
	MessageService MessageService
	maxGoRoutines  int32
	goroutinePool  chan struct{}
}

func NewLineBotService(lineBotClient *linebot.Client, messageService MessageService, options ...WithOption) *LineBotService {
	l := &LineBotService{
		LineBotClient:  lineBotClient,
		MessageService: messageService,
		maxGoRoutines:  10,
	}
	for _, option := range options {
		option(l)
	}
	l.goroutinePool = make(chan struct{}, l.maxGoRoutines)
	return l
}

func (l *LineBotService) Do(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
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

	l.handleEvents(ctx, events)
}

func (l *LineBotService) handleEvents(ctx context.Context, events []*linebot.Event) {
	for _, event := range events {
		l.goroutinePool <- struct{}{}
		l.handleEvent(ctx, event)
	}
}

func (l *LineBotService) handleEvent(ctx context.Context, event *linebot.Event) {
	defer func() {
		<-l.goroutinePool
	}()

	if event.Type == linebot.EventTypeMessage {
		message, err := l.MessageService.Process(ctx, event)
		if err != nil {
			log.Print(err)
			return
		}
		if _, err := l.LineBotClient.ReplyMessage(event.ReplyToken, message).Do(); err != nil {
			log.Print(err)
			return
		}
	}
}
