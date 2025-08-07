package zapx

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"

	"github.com/maa3x/zapx/zapcore"
	"go.opentelemetry.io/otel/codes"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"
)

func (log *Logger) logOpenTelemetry(ce *zapcore.CheckedEntry, fields ...Field) {
	var ctx context.Context
	for _, field := range fields {
		if field.Type == zapcore.ContextType {
			ctx = field.Interface.(context.Context)
		}
	}

	if ce.Level >= zapcore.ErrorLevel && ctx != nil {
		if span := trace.SpanFromContext(ctx); span.IsRecording() {
			span.SetStatus(codes.Error, ce.Message)
		}
	}

	record := otellog.Record{}
	record.SetBody(otellog.StringValue(ce.Message))
	record.SetSeverity(convertLevel(ce.Level))

	kvs := convertFields(fields)
	if log.addStack.Enabled(ce.Level) && ce.Stack != "" {
		kvs = append(kvs, otellog.String("exception.stacktrace", ce.Stack))

	}
	if log.addCaller {
		if ce.Caller.Function != "" {
			kvs = append(kvs, otellog.String("code.function", ce.Caller.Function))
		}
		if ce.Caller.File != "" {
			kvs = append(kvs, otellog.String("code.filepath", ce.Caller.File))
			kvs = append(kvs, otellog.Int("code.lineno", ce.Caller.Line))
		}
	}
	if len(kvs) > 0 {
		record.AddAttributes(kvs...)
	}

	if ctx == nil {
		ctx = context.Background()
	}
	log.otelLogger.Emit(ctx, record)
}

func convertLevel(level zapcore.Level) otellog.Severity {
	switch level {
	case zapcore.DebugLevel:
		return otellog.SeverityDebug
	case zapcore.InfoLevel:
		return otellog.SeverityInfo
	case zapcore.WarnLevel:
		return otellog.SeverityWarn
	case zapcore.ErrorLevel:
		return otellog.SeverityError
	case zapcore.DPanicLevel:
		return otellog.SeverityFatal1
	case zapcore.PanicLevel:
		return otellog.SeverityFatal2
	case zapcore.FatalLevel:
		return otellog.SeverityFatal3
	default:
		return otellog.SeverityUndefined
	}
}

func convertFields(fields []zapcore.Field) []otellog.KeyValue {
	kvs := make([]otellog.KeyValue, 0, len(fields))
	for _, field := range fields {
		kvs = appendField(kvs, field)
	}
	return kvs
}

func appendField(kvs []otellog.KeyValue, f zapcore.Field) []otellog.KeyValue {
	switch f.Type {
	case zapcore.BoolType:
		return append(kvs, otellog.Bool(f.Key, f.Integer == 1))

	case zapcore.Int8Type, zapcore.Int16Type, zapcore.Int32Type, zapcore.Int64Type,
		zapcore.Uint32Type, zapcore.Uint8Type, zapcore.Uint16Type, zapcore.Uint64Type,
		zapcore.UintptrType:
		return append(kvs, otellog.Int64(f.Key, f.Integer))

	case zapcore.Float64Type:
		num := math.Float64frombits(uint64(f.Integer))
		return append(kvs, otellog.Float64(f.Key, num))
	case zapcore.Float32Type:
		num := math.Float32frombits(uint32(f.Integer))
		return append(kvs, otellog.Float64(f.Key, float64(num)))

	case zapcore.Complex64Type:
		str := strconv.FormatComplex(complex128(f.Interface.(complex64)), 'E', -1, 64)
		return append(kvs, otellog.String(f.Key, str))
	case zapcore.Complex128Type:
		str := strconv.FormatComplex(f.Interface.(complex128), 'E', -1, 128)
		return append(kvs, otellog.String(f.Key, str))

	case zapcore.StringType:
		return append(kvs, otellog.String(f.Key, f.String))
	case zapcore.BinaryType, zapcore.ByteStringType:
		bs := f.Interface.([]byte)
		return append(kvs, otellog.Bytes(f.Key, bs))
	case zapcore.StringerType:
		str := f.Interface.(fmt.Stringer).String()
		return append(kvs, otellog.String(f.Key, str))

	case zapcore.DurationType, zapcore.TimeType:
		return append(kvs, otellog.Int64(f.Key, f.Integer))
	case zapcore.TimeFullType:
		str := f.Interface.(time.Time).Format(time.RFC3339Nano)
		return append(kvs, otellog.String(f.Key, str))
	case zapcore.ErrorType:
		err := f.Interface.(error)
		typ := reflect.TypeOf(err).String()
		kvs = append(kvs, otellog.String("exception.type", typ))
		kvs = append(kvs, otellog.String("exception.message", err.Error()))
		return kvs
	case zapcore.ReflectType:
		str := fmt.Sprint(f.Interface)
		return append(kvs, otellog.String(f.Key, str))
	case zapcore.SkipType:
		return kvs

	case zapcore.ArrayMarshalerType:
		kv := otellog.String(f.Key+"_error", "otelzap: zapcore.ArrayMarshalerType is not implemented")
		return append(kvs, kv)
	case zapcore.ObjectMarshalerType:
		kv := otellog.String(f.Key+"_error", "otelzap: zapcore.ObjectMarshalerType is not implemented")
		return append(kvs, kv)

	default:
		kv := otellog.String(f.Key+"_error", fmt.Sprintf("otelzap: unknown field type: %v", f))
		return append(kvs, kv)
	}
}
