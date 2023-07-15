package linebotservice

type WithOption func(*LineBotService)

func WithMaxGoRoutines(maxGoRoutines int32) WithOption {
	return func(l *LineBotService) {
		l.maxGoRoutines = maxGoRoutines
	}
}
