package chatgpt

type WithOptoin func(*MessageCore)

func WithMemory(memory Memory) WithOptoin {
	return func(messageCore *MessageCore) {
		messageCore.memory = memory
	}
}

func WithMemoryN(memoryN int) WithOptoin {
	return func(messageCore *MessageCore) {
		messageCore.memoryN = memoryN
	}
}

func WithChatModel(chatModel string) WithOptoin {
	return func(messageCore *MessageCore) {
		messageCore.chatModel = chatModel
	}
}

func WithChatToken(chatToken int) WithOptoin {
	return func(messageCore *MessageCore) {
		messageCore.chatToken = chatToken
	}
}

func WithChatTemperature(chatTemperature float32) WithOptoin {
	return func(messageCore *MessageCore) {
		messageCore.chatTemperature = chatTemperature
	}
}
