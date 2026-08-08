package zapx

import (
	"context"

	"github.com/maa3x/zapx/zapcore"
)

func RegisterContextEncoder(enc func(context.Context) Field) {
	zapcore.RegisterContextEncoder(enc)
}

type contextKeyType struct{}

var contextKey contextKeyType

func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(contextKey).(*Logger); ok {
		return logger
	}
	return nil
}

func (log *Logger) Context(ctx context.Context) (context.Context, *Logger) {
	ctx = context.WithValue(ctx, contextKey, log)
	log.ctx = ctx
	return ctx, log
}

func (log *Logger) NamedContext(ctx context.Context, name string) (context.Context, *Logger) {
	return log.Named(name).Context(ctx)
}
