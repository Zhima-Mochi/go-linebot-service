package messageservice

import (
	"github.com/Zhima-Mochi/go-linebot-service/messageservice/messagecorefactory"
	"github.com/Zhima-Mochi/go-linebot-service/messageservice/messagecorefactory/echo"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type MessageService struct {
	defaultMessageCore       messagecorefactory.MessageCore
	customMessageTypeCoreMap map[linebot.MessageType]messagecorefactory.MessageCore
}

func NewMessageService() *MessageService {
	return &MessageService{
		defaultMessageCore:       echo.NewMessageCore(),
		customMessageTypeCoreMap: make(map[linebot.MessageType]messagecorefactory.MessageCore),
	}
}

func (m *MessageService) SetDefaultMessageCore(messageCore messagecorefactory.MessageCore) {
	m.defaultMessageCore = messageCore
}

func (m *MessageService) GetDefaultMessageCore() messagecorefactory.MessageCore {
	return m.defaultMessageCore
}

func (m *MessageService) SetCustomMessageTypeCore(messageType linebot.MessageType, messageCore messagecorefactory.MessageCore) {
	m.customMessageTypeCoreMap[messageType] = messageCore
}

func (m *MessageService) GetCustomMessageTypeCore(messageType linebot.MessageType) messagecorefactory.MessageCore {
	if core, ok := m.customMessageTypeCoreMap[messageType]; ok {
		return core
	}
	return m.defaultMessageCore
}

func (m *MessageService) ClearCustomMessageTypeCore(messageType linebot.MessageType) {
	delete(m.customMessageTypeCoreMap, messageType)
}

func (m *MessageService) ClearAllCustomMessageTypeCore() {
	m.customMessageTypeCoreMap = make(map[linebot.MessageType]messagecorefactory.MessageCore)
}

func (m *MessageService) Process(message linebot.Message) (linebot.SendingMessage, error) {
	return m.GetCustomMessageTypeCore(message.Type()).Process(message)
}
