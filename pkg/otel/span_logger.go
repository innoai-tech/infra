package otel

import (
	"context"
	"fmt"
	"golang.org/x/exp/slog"
	"time"

	"github.com/go-courier/logr"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func newSpanLogger(tp trace.TracerProvider, span trace.Span, levelEnabled logr.Level, slog *slog.Logger) logr.Logger {
	return &spanLogger{
		tp:          tp,
		slog:        slog,
		enabled:     levelEnabled,
		spanContext: &spanContext{span: span},
	}
}

type spanLogger struct {
	enabled     logr.Level
	tp          trace.TracerProvider
	slog        *slog.Logger
	spanContext *spanContext
	attributes  []attribute.KeyValue
}

type spanContext struct {
	span      trace.Span
	name      string
	startedAt time.Time
}

func (c *spanContext) toAttrs(attributes []attribute.KeyValue) []slog.Attr {
	attrs := make([]slog.Attr, len(attributes))
	for i := range attrs {
		a := attributes[i]
		attrs[i] = slog.Any(string(a.Key), a.Value.AsInterface())
	}

	spanCtx := c.span.SpanContext()
	if traceID := spanCtx.TraceID(); traceID.IsValid() {
		attrs = append(attrs, slog.String("traceID", traceID.String()))
	}

	if spanCtx.HasSpanID() {
		attrs = append(
			attrs,
			slog.String("spanID", spanCtx.SpanID().String()),
		)
	}

	if c.name != "" {
		attrs = append(attrs, slog.String("spanName", c.name))
	}

	return attrs
}

func (c spanContext) withName(name string) *spanContext {
	c.name = name
	return &c
}

func (c spanContext) start(ctx context.Context, tp trace.TracerProvider, spanName string, keyAndValues ...any) (context.Context, []attribute.KeyValue, *spanContext) {
	c.startedAt = time.Now()
	c.name = appendName(c.name, spanName)

	n, attrs := attrsFromKeyAndValues(c.name, keyAndValues...)
	c.name = n

	cc, span := tp.Tracer("").Start(ctx, c.name, trace.WithTimestamp(c.startedAt))
	c.span = span
	return cc, attrs, &c
}

func (t *spanLogger) WithValues(keyAndValues ...any) logr.Logger {
	if len(keyAndValues) == 0 {
		return t
	}

	name, attrs := attrsFromKeyAndValues(t.spanContext.name, keyAndValues...)

	return &spanLogger{
		enabled:     t.enabled,
		tp:          t.tp,
		slog:        t.slog,
		spanContext: t.spanContext.withName(name),
		attributes:  append(t.attributes, attrs...),
	}
}

func (t *spanLogger) Start(ctx context.Context, name string, keyAndValues ...any) (context.Context, logr.Logger) {
	c, attrs, spanCtx := t.spanContext.start(ctx, t.tp, name, keyAndValues...)

	lgr := &spanLogger{
		enabled:     t.enabled,
		tp:          t.tp,
		slog:        t.slog,
		spanContext: spanCtx,
		attributes:  append(t.attributes, attrs...),
	}

	return logr.WithLogger(c, lgr), lgr
}

func appendName(name string, name2 string) string {
	if name == "" {
		return name2
	}
	return name + "/" + name2
}

func (t *spanLogger) End() {
	endAt := time.Now()

	t.spanContext.span.End(trace.WithTimestamp(endAt))

	if l := t.slog; l != nil {
		attrs := t.spanContext.toAttrs(t.attributes)
		attrs = append(attrs, slog.String("spanCost", endAt.Sub(t.spanContext.startedAt).String()))
		l.LogAttrs(context.Background(), slog.LevelInfo, "done", attrs...)
	}
}

func (t *spanLogger) info(level logr.Level, msg fmt.Stringer) {
	if level > t.enabled {
		return
	}

	span := t.spanContext.span

	msgStr := msg.String()

	span.AddEvent(
		msgStr,
		trace.WithTimestamp(time.Now()),
		trace.WithAttributes(append(t.attributes, attribute.String("@level", level.String()))...),
	)

	if l := t.slog; l != nil {
		attrs := t.spanContext.toAttrs(t.attributes)

		switch level {
		case logr.DebugLevel:
			l.LogAttrs(context.Background(), slog.LevelDebug, msgStr, attrs...)
		case logr.InfoLevel:
			l.LogAttrs(context.Background(), slog.LevelInfo, msgStr, attrs...)
		}
	}
}

func (t *spanLogger) error(level logr.Level, err error) {
	if level > t.enabled {
		return
	}

	if err == nil {
		return
	}

	span := t.spanContext.span

	attributes := t.attributes

	if level <= logr.ErrorLevel {
		attributes = append(attributes, attribute.String("exception.stack", fmt.Sprintf("%+v", err)))
	}

	span.RecordError(err,
		trace.WithTimestamp(time.Now()),
		trace.WithAttributes(append(attributes, attribute.String("@level", level.String()))...),
	)

	if l := t.slog; l != nil {
		attrs := t.spanContext.toAttrs(attributes)

		attrs = append(attrs, slog.String("exception.message", err.Error()))

		switch level {
		case logr.WarnLevel:
			l.LogAttrs(context.Background(), slog.LevelWarn, "", attrs...)
		case logr.ErrorLevel:
			l.LogAttrs(context.Background(), slog.LevelError, "", attrs...)
		}
	}
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

func attrsFromKeyAndValues(name string, keysAndValues ...any) (string, []attribute.KeyValue) {
	fields := make([]attribute.KeyValue, 0, len(keysAndValues))

	i := 0
	for i < len(keysAndValues) {
		k := keysAndValues[i]

		switch key := k.(type) {
		case attribute.KeyValue:
			fields = append(fields, key)
			i++
		case string:
			if i+1 < len(keysAndValues) {
				i++
				v := keysAndValues[i]
				i++

				if key == "@name" {
					name = appendName(name, v.(string))
					continue
				}

				switch x := v.(type) {
				case fmt.Stringer:
					fields = append(fields, attribute.Stringer(key, x))
				case string:
					fields = append(fields, attribute.String(key, x))
				case int:
					fields = append(fields, attribute.Int(key, x))
				case float64:
					fields = append(fields, attribute.Float64(key, x))
				case bool:
					fields = append(fields, attribute.Bool(key, x))
				default:
					fields = append(fields, attribute.String(key, fmt.Sprintf("%v", x)))
				}
			}
		}
	}

	return name, fields
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
