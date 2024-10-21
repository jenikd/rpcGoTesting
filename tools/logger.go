package tools

import "log"

type Logger struct {
	*log.Logger
	logLevel int
}

const (
	TraceLevel = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
)

func NewLogger(logLevel int) *Logger {

	return &Logger{
		Logger:   log.New(log.Writer(), "", log.Lshortfile),
		logLevel: logLevel,
	}
}
func (l *Logger) Debugf(format string, v ...interface{}) {

	if l.logLevel <= DebugLevel {
		l.Printf(format, v...)
	}
}

func (l *Logger) Debugln(v ...any) {

	if l.logLevel <= DebugLevel {
		l.Println(v...)
	}
}
