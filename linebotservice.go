package linebotservice

import (
	"github.com/Zhima-Mochi/go-linebot-service/messageservice"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type LineBotService struct {
	LineBotClient  *linebot.Client
	messageService *messageservice.MessageService
}

func NewLineBotService(lineBotClient *linebot.Client) *LineBotService {
	return &LineBotService{LineBotClient: lineBotClient}
}
