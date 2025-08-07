package zapx

import (
	"context"

	"github.com/maa3x/zapx/zapcore"
)

func RegisterContextEncoder(enc func(context.Context) Field) {
	zapcore.RegisterContextEncoder(enc)
}
