package logger

type Log interface {
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
	Panic(msg string)
}

type Handler func(severity int, msg string)

type Logger struct {
	handler Handler
}

func New() *Logger {
	return &Logger{}
}

func (l *Logger) HandleFunc(handler Handler) {
	l.handler = handler
}

func (l *Logger) Info(msg string) {
	if l.handler == nil {
		return
	}
	l.handler(Info, msg)
}

func (l *Logger) Warn(msg string) {
	if l.handler == nil {
		return
	}
	l.handler(Warn, msg)
}

func (l *Logger) Error(msg string) {
	if l.handler == nil {
		return
	}
	l.handler(Error, msg)
}

func (l *Logger) Fatal(msg string) {
	if l.handler == nil {
		return
	}
	l.handler(Fatal, msg)
}

func (l *Logger) Panic(msg string) {
	if l.handler == nil {
		return
	}
	l.handler(Panic, msg)
}
