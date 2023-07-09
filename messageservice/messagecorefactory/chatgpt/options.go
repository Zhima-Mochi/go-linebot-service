package chatgpt

type WithOption func(*MessageCore)

func WithMemory(memory Memory) WithOption {
	return func(messageCore *MessageCore) {
		messageCore.memory = memory
	}
}

func WithMemoryN(memoryN int) WithOption {
	return func(messageCore *MessageCore) {
		messageCore.memoryN = memoryN
	}
}

func WithChatModel(chatModel string) WithOption {
	return func(messageCore *MessageCore) {
		messageCore.chatModel = chatModel
	}
}

func WithChatToken(chatToken int) WithOption {
	return func(messageCore *MessageCore) {
		messageCore.chatToken = chatToken
	}
}

func WithChatTemperature(chatTemperature float32) WithOption {
	return func(messageCore *MessageCore) {
		messageCore.chatTemperature = chatTemperature
	}
}
