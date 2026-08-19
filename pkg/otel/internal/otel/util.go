package otel

import (
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/octohelm/x/slices"
)

func normalizeKeyValues(keysAndValues []any) []attribute.KeyValue {
	keyValues := make([]attribute.KeyValue, 0, len(keysAndValues))

	for i := 0; i < len(keysAndValues); i++ {
		switch x := keysAndValues[i].(type) {
		case []slog.Attr:
			keyValues = append(keyValues, slices.Map(x, func(e slog.Attr) attribute.KeyValue {
				return attribute.KeyValue{
					Key:   attribute.Key(e.Key),
					Value: LogAnyValue(e.Value.Any()),
				}
			})...)
		case slog.Attr:
			keyValues = append(keyValues, attribute.KeyValue{
				Key:   attribute.Key(x.Key),
				Value: LogAnyValue(x.Value.Any()),
			})
		case []attribute.KeyValue:
			keyValues = append(keyValues, x...)
		case attribute.KeyValue:
			keyValues = append(keyValues, x)
		case string:
			// "key", value
			if i+1 < len(keysAndValues) {
				i++
				keyValues = append(keyValues, attribute.KeyValue{
					Key:   attribute.Key(x),
					Value: LogAnyValue(keysAndValues[i]),
				})
			}
		default:
			panic(fmt.Errorf("unsupported log attr values %T", x))
		}
	}

	return keyValues
}

// LogValue 将 OpenTelemetry attribute.Value 转换为普通 Go 值。
func LogValue(v attribute.Value) any {
	switch v.Type() {
	case attribute.BOOL:
		return v.AsBool()
	case attribute.BOOLSLICE:
		return v.AsBoolSlice()
	case attribute.FLOAT64:
		return v.AsFloat64()
	case attribute.FLOAT64SLICE:
		return v.AsFloat64Slice()
	case attribute.INT64:
		return v.AsInt64()
	case attribute.INT64SLICE:
		return v.AsInt64Slice()
	case attribute.STRING:
		return v.AsString()
	case attribute.STRINGSLICE:
		return v.AsStringSlice()
	case attribute.BYTESLICE:
		return v.AsByteSlice()
	case attribute.SLICE:
		list := v.AsSlice()
		values := make([]any, len(list))
		for i := range list {
			values[i] = LogValue(list[i])
		}
		return values
	case attribute.MAP:
		values := map[string]any{}
		for _, k := range v.AsMap() {
			values[string(k.Key)] = LogValue(k.Value)
		}
		return values
	default:
		return nil
	}
}

// LogAnyValue 将任意 Go 值转换为 OpenTelemetry attribute.Value。
func LogAnyValue(value any) attribute.Value {
	switch v := value.(type) {
	case time.Time:
		return attribute.StringValue(slog.TimeValue(v).String())
	case time.Duration:
		return attribute.StringValue(slog.DurationValue(v).String())
	case fmt.Stringer:
		return attribute.StringValue(v.String())
	case []byte:
		return attribute.ByteSliceValue(v)
	case []string:
		return attribute.StringSliceValue(v)
	case []bool:
		return attribute.BoolSliceValue(v)
	case []int:
		return attribute.IntSliceValue(v)
	case []int8:
		return attribute.Int64SliceValue(slices.Map(v, func(e int8) int64 { return int64(e) }))
	case []int16:
		return attribute.Int64SliceValue(slices.Map(v, func(e int16) int64 { return int64(e) }))
	case []int32:
		return attribute.Int64SliceValue(slices.Map(v, func(e int32) int64 { return int64(e) }))
	case []int64:
		return attribute.Int64SliceValue(v)
	case []uint:
		return attribute.Int64SliceValue(slices.Map(v, func(e uint) int64 { return int64(e) }))
	case []uint16:
		return attribute.Int64SliceValue(slices.Map(v, func(e uint16) int64 { return int64(e) }))
	case []uint32:
		return attribute.Int64SliceValue(slices.Map(v, func(e uint32) int64 { return int64(e) }))
	case []uint64:
		return attribute.Int64SliceValue(slices.Map(v, func(e uint64) int64 { return int64(e) }))
	case []float32:
		return attribute.Float64SliceValue(slices.Map(v, func(e float32) float64 { return float64(e) }))
	case []float64:
		return attribute.Float64SliceValue(v)
	case string:
		return attribute.StringValue(v)
	case uint:
		return attribute.Int64Value(int64(v))
	case uint8:
		return attribute.Int64Value(int64(v))
	case uint16:
		return attribute.Int64Value(int64(v))
	case uint32:
		return attribute.Int64Value(int64(v))
	case int:
		return attribute.Int64Value(int64(v))
	case int8:
		return attribute.Int64Value(int64(v))
	case int16:
		return attribute.Int64Value(int64(v))
	case int32:
		return attribute.Int64Value(int64(v))
	case int64:
		return attribute.Int64Value(v)
	case float32:
		return attribute.Float64Value(float64(v))
	case float64:
		return attribute.Float64Value(v)
	case bool:
		return attribute.BoolValue(v)
	case []any:
		list := make([]attribute.Value, len(v))
		for i, item := range v {
			list[i] = LogAnyValue(item)
		}
		return attribute.SliceValue(list...)
	case map[string]any:
		keyValues := make([]attribute.KeyValue, 0, len(v))
		for k, item := range v {
			keyValues = append(keyValues, attribute.KeyValue{
				Key:   attribute.Key(k),
				Value: LogAnyValue(item),
			})
		}
		return attribute.MapValue(keyValues...)
	default:
		if x, ok := v.(interface{ Unwrap() any }); ok {
			return LogAnyValue(x.Unwrap())
		}
		return attribute.StringValue(slog.AnyValue(v).String())
	}
}
