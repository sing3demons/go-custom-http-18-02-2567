package logger

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

type LLogger struct {
	logger *logrus.Logger
	ctx    context.Context
}

func NewLLogger(ctx context.Context) ILogger {
	logLevel, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = logrus.InfoLevel
	}

	logger := logrus.New()
	logger.Out = os.Stdout
	logger.Formatter = &logrus.JSONFormatter{}
	logrus.SetLevel(logLevel)
	log.SetOutput(logger.Writer())
	logger.SetOutput(io.MultiWriter(os.Stdout))
	return &LLogger{logger: logger, ctx: ctx}
}

func (l *LLogger) Info(msg string, fields map[string]any) {
	l.logger.WithFields(fields).Info(msg)
}

func (l *LLogger) WithField(key string, value any) ILogger {
	newLogger := logrus.NewEntry(l.logger)
	newLogger.Logger.WithFields(logrus.Fields{key: value})
	return &LLogger{logger: newLogger.Logger, ctx: l.ctx}
}

func (l *LLogger) Warn(msg string, fields map[string]any) {
	l.logger.WithFields(fields).Warn(msg)
}

func (l *LLogger) Error(msg string, fields map[string]any) {
	l.logger.WithFields(fields).Error(msg)
}

func (l *LLogger) Fatal(msg string, fields map[string]any) {
	l.logger.WithFields(fields).Fatal(msg)
}
func (l *LLogger) Debug(msg string, fields map[string]any) {
	l.logger.WithFields(fields).Debug(msg)
}
