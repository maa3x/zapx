package zapcore

import (
	"context"
)

var contextEncoders []func(context.Context) Field

func RegisterContextEncoder(encoder func(context.Context) Field) {
	contextEncoders = append(contextEncoders, encoder)
}

func FieldsFromContext(ctx context.Context) []Field {
	if ctx == nil {
		return nil
	}

	fields := make([]Field, 0, len(contextEncoders))
	for _, encoder := range contextEncoders {
		if field := encoder(ctx); field.Type != SkipType {
			fields = append(fields, field)
		}
	}
	return fields
}
