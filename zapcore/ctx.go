package zapcore

import (
	"context"
)

var contextEncoders []func(context.Context) Field

func RegisterContextEncoder(encoder func(context.Context) Field) {
	contextEncoders = append(contextEncoders, encoder)
}
