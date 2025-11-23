package logger

import (
	"context"
	"log"

	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey struct{}

var once sync.Once

var logger *zap.Logger

func Get() *zap.Logger {
	once.Do(func() {
		conf := zap.NewProductionConfig()
		conf.EncoderConfig.TimeKey = "timestamp"
		conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		var err error
		logger, err = conf.Build()
		if err != nil {
			log.Fatalf("failed to initialize logger: %v", err)
		}
	})

	return logger
}

func FromCtx(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		return l
	} else if globalLogger := logger; l != nil {
		return globalLogger
	}

	return zap.NewNop()
}

func WithCtx(ctx context.Context, l *zap.Logger) context.Context {
	if lp, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		if lp == l {
			return ctx
		}
	}

	return context.WithValue(ctx, ctxKey{}, l)
}
