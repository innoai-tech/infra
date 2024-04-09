package otel

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/octohelm/x/slices"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"
)

func normalizeKeyValues(keysAndValues []any) []log.KeyValue {
	keyValues := make([]log.KeyValue, 0, len(keysAndValues))

	for i := 0; i < len(keysAndValues); i++ {
		switch x := keysAndValues[i].(type) {
		case []slog.Attr:
			slog.GroupValue()

			keyValues = append(keyValues, slices.Map(x, func(e slog.Attr) log.KeyValue {
				return log.KeyValue{
					Key:   e.Key,
					Value: LogAnyValue(e.Value.Any()),
				}
			})...)
		case slog.Attr:
			keyValues = append(keyValues, log.KeyValue{
				Key:   x.Key,
				Value: LogAnyValue(x.Value.Any()),
			})
		case []attribute.KeyValue:
			keyValues = append(keyValues, slices.Map(x, func(e attribute.KeyValue) log.KeyValue {
				return log.KeyValue{
					Key:   string(e.Key),
					Value: LogAnyValue(e.Value.AsInterface()),
				}
			})...)
		case attribute.KeyValue:
			keyValues = append(keyValues, log.KeyValue{
				Key:   string(x.Key),
				Value: LogAnyValue(x.Value.AsInterface()),
			})
		case string:
			// "key", value
			if i+1 < len(keysAndValues) {
				i++
				keyValues = append(keyValues, log.KeyValue{
					Key:   x,
					Value: LogAnyValue(keysAndValues[i]),
				})
			}
		default:
			panic(fmt.Errorf("unsupported log attr values %T", x))
		}
	}

	return keyValues
}

func LogValue(v log.Value) any {
	switch v.Kind() {
	case log.KindBool:
		return v.AsBool()
	case log.KindFloat64:
		return v.AsFloat64()
	case log.KindInt64:
		return v.AsInt64()
	case log.KindString:
		return v.AsString()
	case log.KindBytes:
		return v.AsBytes()
	case log.KindSlice:
		list := v.AsSlice()
		values := make([]any, len(list))
		for i := range list {
			values[i] = LogValue(list[i])
		}
		return values
	case log.KindMap:
		values := map[string]any{}
		for _, k := range v.AsMap() {
			values[k.Key] = LogValue(k.Value)
		}
		return values
	default:
		return nil
	}
}

func LogAnyValue(value any) log.Value {
	switch v := value.(type) {
	case time.Time:
		return log.StringValue(slog.TimeValue(v).String())
	case time.Duration:
		return log.StringValue(slog.DurationValue(v).String())
	case fmt.Stringer:
		return log.StringValue(v.String())
	case []byte:
		return log.BytesValue(v)
	case string:
		return log.StringValue(v)
	case uint:
		return log.Int64Value(int64(v))
	case uint8:
		return log.Int64Value(int64(v))
	case uint16:
		return log.Int64Value(int64(v))
	case uint32:
		return log.Int64Value(int64(v))
	case int:
		return log.Int64Value(int64(v))
	case int8:
		return log.Int64Value(int64(v))
	case int16:
		return log.Int64Value(int64(v))
	case int32:
		return log.Int64Value(int64(v))
	case int64:
		return log.Int64Value(v)
	case float32:
		return log.Float64Value(float64(v))
	case float64:
		return log.Float64Value(v)
	case bool:
		return log.BoolValue(v)
	case []any:
		list := make([]log.Value, len(v))
		for i, item := range v {
			list[i] = LogAnyValue(item)
		}
		return log.SliceValue(list...)
	case map[string]any:
		keyValues := make([]log.KeyValue, 0, len(v))
		for k, item := range v {
			keyValues = append(keyValues, log.KeyValue{
				Key:   k,
				Value: LogAnyValue(item),
			})
		}
		return log.MapValue(keyValues...)
	default:
		return log.StringValue(slog.AnyValue(v).String())
	}
}
