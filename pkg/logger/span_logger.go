package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/innoai-tech/infra/internal/otel"

	"github.com/go-courier/logr"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func newSpanLogger(span trace.Span, levelEnabled logr.Level) logr.Logger {
	return &spanLogger{span: span, enabled: levelEnabled}
}

type spanLogger struct {
	enabled    logr.Level
	span       trace.Span
	name       string
	attributes []attribute.KeyValue
}

func (t *spanLogger) withName(name string) *spanLogger {
	return &spanLogger{
		span:       t.span,
		name:       name,
		enabled:    t.enabled,
		attributes: t.attributes,
	}
}

func (t *spanLogger) WithValues(keyAndValues ...any) logr.Logger {
	return &spanLogger{
		span:       t.span,
		name:       t.name,
		enabled:    t.enabled,
		attributes: append(t.attributes, attrsFromKeyAndValues(keyAndValues...)...),
	}
}

func (t *spanLogger) Start(ctx context.Context, name string, keyAndValues ...any) (context.Context, logr.Logger) {
	if t.name != "" {
		name = t.name + "/" + name
	}

	c, span := otel.TracerProviderFromContext(ctx).Tracer("").Start(
		ctx,
		name,
		trace.WithAttributes(attrsFromKeyAndValues(keyAndValues...)...),
		trace.WithTimestamp(time.Now()),
	)

	lgr := &spanLogger{
		enabled: t.enabled,
		span:    span,
		name:    name,
	}

	return logr.WithLogger(c, lgr), lgr
}

func (t *spanLogger) End() {
	if t.span == nil {
		return
	}
	t.span.End(trace.WithTimestamp(time.Now()))
}

func (t *spanLogger) info(level logr.Level, msg fmt.Stringer, keyAndValues ...any) {
	if level > t.enabled {
		return
	}

	attributes := append(t.attributes, attrsFromKeyAndValues(keyAndValues...)...)
	attributes = append(attributes, attribute.String("@level", level.String()))

	t.span.AddEvent(
		msg.String(),
		trace.WithTimestamp(time.Now()),
		trace.WithAttributes(attributes...),
	)
}

func (t *spanLogger) error(level logr.Level, err error, keyAndValues ...any) {
	if level > t.enabled {
		return
	}

	if t.span == nil || err == nil || !t.span.IsRecording() {
		return
	}

	attributes := append(t.attributes, attrsFromKeyAndValues(keyAndValues...)...)

	attributes = append(attributes, attribute.String("@level", level.String()))

	if level <= logr.ErrorLevel {
		attributes = append(attributes, attribute.String("stack", fmt.Sprintf("%+v", err)))
	}

	t.span.RecordError(err,
		trace.WithTimestamp(time.Now()),
		trace.WithAttributes(attributes...),
	)
}

func (t *spanLogger) Trace(msgOrFormat string, args ...any) {
	t.info(logr.TraceLevel, Sprintf(msgOrFormat, args...))
}

func (t *spanLogger) Debug(msgOrFormat string, args ...any) {
	t.info(logr.DebugLevel, Sprintf(msgOrFormat, args...))
}

func (t *spanLogger) Info(msgOrFormat string, args ...any) {
	t.info(logr.InfoLevel, Sprintf(msgOrFormat, args...))
}

func (t *spanLogger) Warn(err error) {
	t.error(logr.WarnLevel, err)
}

func (t *spanLogger) Error(err error) {
	t.error(logr.ErrorLevel, err)
}

func (t *spanLogger) Fatal(err error) {
	t.error(logr.FatalLevel, err)
}

func (t *spanLogger) Panic(err error) {
	t.error(logr.PanicLevel, err)
	panic(err)
}

func attrsFromKeyAndValues(keysAndValues ...any) []attribute.KeyValue {
	n := len(keysAndValues)
	if n > 0 && n%2 == 0 {
		fields := make([]attribute.KeyValue, len(keysAndValues)/2)
		for i := range fields {
			k, v := keysAndValues[2*i], keysAndValues[2*i+1]

			if key, ok := k.(string); ok {
				switch x := v.(type) {
				case fmt.Stringer:
					fields[i] = attribute.Stringer(key, x)
				case string:
					fields[i] = attribute.String(key, x)
				case int:
					fields[i] = attribute.Int(key, x)
				case float64:
					fields[i] = attribute.Float64(key, x)
				case bool:
					fields[i] = attribute.Bool(key, x)
				default:
					fields[i] = attribute.String(key, fmt.Sprintf("%v", x))
				}
			}
		}
		return fields
	}
	return nil
}

func Sprintf(format string, args ...any) fmt.Stringer {
	return &printer{format: format, args: args}
}

type printer struct {
	format string
	args   []any
}

func (p *printer) String() string {
	if len(p.args) == 0 {
		return p.format
	}
	return fmt.Sprintf(p.format, p.args...)
}
