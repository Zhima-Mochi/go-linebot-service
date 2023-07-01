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
