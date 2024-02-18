package logger

import (
	"context"

	"go.uber.org/zap"
)

type ZapLog struct {
	logger *zap.Logger
	ctx    context.Context
}

func NewZapLog(ctx context.Context) ILogger {
	logger, _ := zap.NewProduction()
	return &ZapLog{logger: logger, ctx: ctx}
}

func (l *ZapLog) Info(msg string, fields map[string]any) {
	l.logger.Info(msg, zap.Any("args", fields))
}

func (l *ZapLog) WithField(key string, value any) ILogger {
	l.logger.With(zap.Any(key, value))
	return l
}

func (l *ZapLog) Warn(msg string, fields map[string]any) {
	l.logger.Warn(msg, zap.Any("args", fields))
}

func (l *ZapLog) Error(msg string, fields map[string]any) {
	l.logger.Error(msg, zap.Any("args", fields))
}

func (l *ZapLog) Fatal(msg string, fields map[string]any) {
	l.logger.Fatal(msg, zap.Any("args", fields))
}

func (l *ZapLog) Debug(msg string, fields map[string]any) {
	l.logger.Debug(msg, zap.Any("args", fields))
}
