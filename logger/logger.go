package logger

import (
	"context"
)

type ILogger interface {
	Debug(msg string, field map[string]any)
	Info(msg string, field map[string]any)
	Warn(msg string, field map[string]any)
	Error(msg string, field map[string]any)
	Fatal(msg string, field map[string]any)
}

type Logger struct {
	ILogger
}

func NewLoggerWrapper(loggerType string, ctx context.Context) *Logger {
	var logger ILogger
	switch loggerType {
	case "logrus":
		logger = NewLLogger(ctx)
	case "zap":
		logger = NewZapLog(ctx)
	default:
		logger = NewLLogger(ctx)
	}
	return &Logger{logger}
}
