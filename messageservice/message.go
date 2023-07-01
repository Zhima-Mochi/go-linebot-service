package messageservice

import (
	"github.com/Zhima-Mochi/go-linebot-service/messageservice/factory"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type MessageService struct {
	defaultMessageCore       factory.MessageCore
	customMessageTypeCoreMap map[linebot.MessageType]factory.MessageCore
}

func NewMessageService() *MessageService {
	return &MessageService{
		defaultMessageCore:       nil,
		customMessageTypeCoreMap: make(map[linebot.MessageType]factory.MessageCore),
	}
}

func (m *MessageService) SetDefaultMessageCore(messageCore factory.MessageCore) {
	m.defaultMessageCore = messageCore
}

func (m *MessageService) GetDefaultMessageCore() factory.MessageCore {
	return m.defaultMessageCore
}

func (m *MessageService) SetCustomMessageTypeCore(messageType linebot.MessageType, messageCore factory.MessageCore) {
	m.customMessageTypeCoreMap[messageType] = messageCore
}

func (m *MessageService) GetCustomMessageTypeCore(messageType linebot.MessageType) factory.MessageCore {
	if core, ok := m.customMessageTypeCoreMap[messageType]; ok {
		return core
	}
	return m.defaultMessageCore
}

func (m *MessageService) Process(message linebot.Message) (linebot.Message, error) {
	return m.GetCustomMessageTypeCore(message.Type()).Process(message)
}
