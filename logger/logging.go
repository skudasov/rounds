package logger

import (
	"context"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type correlationIdType int

const (
	requestIdKey correlationIdType = iota
	sessionIdKey
)

type Logger struct {
	*zap.SugaredLogger
}

var logger *Logger

// WithRqId returns a context which knows its request ID
func WithRqId(ctx context.Context, rqId string) context.Context {
	return context.WithValue(ctx, requestIdKey, rqId)
}

// WithSessionId returns a context which knows its session ID
func WithSessionId(ctx context.Context, sessionId string) context.Context {
	return context.WithValue(ctx, sessionIdKey, sessionId)
}

// Logger returns a zap logger with as much context as possible
func (m *Logger) FromCtx(ctx context.Context) *Logger {
	newLogger := logger
	if ctx != nil {
		if ctxRqId, ok := ctx.Value(requestIdKey).(string); ok {
			newLogger = &Logger{newLogger.With(zap.String("rqId", ctxRqId))}
		}
		if ctxSessionId, ok := ctx.Value(sessionIdKey).(string); ok {
			newLogger = &Logger{newLogger.With(zap.String("sessionId", ctxSessionId))}
		}
	}
	return newLogger
}

func NewLogger() *Logger {
	var l *zap.Logger
	switch viper.GetString("logging.level") {
	case "debug":
		l, _ = zap.NewDevelopment()
	case "production":
		l, _ = zap.NewProduction()
	default:
		l, _ = zap.NewDevelopment()
	}
	return &Logger{l.Sugar()}
}
