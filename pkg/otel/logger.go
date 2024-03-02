package otel

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-courier/logr"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/innoai-tech/infra/internal/otel"
)

func newLogger(
	ctx context.Context,
	tp trace.TracerProvider,
	directLogger *slog.Logger,
	levelEnabled logr.Level,
) logr.Logger {
	return &logger{
		enabled:      levelEnabled,
		tp:           tp,
		directLogger: directLogger,
		spanContext:  &spanContext{span: trace.SpanFromContext(ctx)},
	}
}

type logger struct {
	enabled      logr.Level
	tp           trace.TracerProvider
	directLogger *slog.Logger
	spanContext  *spanContext
	attributes   []attribute.KeyValue
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

func (t *logger) WithValues(keyAndValues ...any) logr.Logger {
	if len(keyAndValues) == 0 {
		return t
	}

	name, attrs := attrsFromKeyAndValues(t.spanContext.name, keyAndValues...)

	return &logger{
		enabled:      t.enabled,
		tp:           t.tp,
		directLogger: t.directLogger,
		spanContext:  t.spanContext.withName(name),
		attributes:   append(t.attributes, attrs...),
	}
}

func (t *logger) Start(ctx context.Context, name string, keyAndValues ...any) (context.Context, logr.Logger) {
	c, attrs, spanCtx := t.spanContext.start(ctx, t.tp, name, keyAndValues...)

	lgr := &logger{
		enabled:      t.enabled,
		tp:           t.tp,
		spanContext:  spanCtx,
		directLogger: t.directLogger,
		attributes:   append(t.attributes, attrs...),
	}

	return logr.WithLogger(c, lgr), lgr
}

func appendName(name string, name2 string) string {
	if name == "" {
		return name2
	}
	return name + "/" + name2
}

func (t *logger) End() {
	endAt := time.Now()

	t.spanContext.span.End(trace.WithTimestamp(endAt))

	if l := t.directLogger; l != nil {
		attrs := t.spanContext.toAttrs(t.attributes)
		attrs = append(attrs, slog.String("spanCost", endAt.Sub(t.spanContext.startedAt).String()))

		if t.enabled >= logr.DebugLevel {
			attrs = append(attrs, slog.Any("source", otel.Source(3)))
		}

		l.LogAttrs(context.Background(), slog.LevelDebug, "done", attrs...)
	}
}

func (t *logger) span() trace.Span {
	return t.spanContext.span
}

func (t *logger) info(level logr.Level, msg fmt.Stringer) {
	if level > t.enabled {
		return
	}

	msgStr := msg.String()

	t.span().AddEvent(
		msgStr,
		trace.WithTimestamp(time.Now()),
		trace.WithAttributes(append(t.attributes, attribute.String("@level", level.String()))...),
	)

	if l := t.directLogger; l != nil {
		attrs := t.spanContext.toAttrs(t.attributes)

		if t.enabled >= logr.DebugLevel {
			attrs = append(attrs, slog.Any("source", otel.Source(3)))
		}

		switch level {
		case logr.DebugLevel:
			l.LogAttrs(context.Background(), slog.LevelDebug, msgStr, attrs...)
		case logr.InfoLevel:
			l.LogAttrs(context.Background(), slog.LevelInfo, msgStr, attrs...)
		default:
		}
	}
}

func (t *logger) error(level logr.Level, err error) {
	if level > t.enabled {
		return
	}

	if err == nil {
		return
	}

	attributes := t.attributes

	t.span().RecordError(
		err,
		trace.WithTimestamp(time.Now()),
		trace.WithAttributes(append(t.attributes, attribute.String("@level", level.String()))...),
	)

	if l := t.directLogger; l != nil {
		attrs := t.spanContext.toAttrs(attributes)

		if t.enabled >= logr.DebugLevel {
			attrs = append(attrs, slog.Any("source", otel.Source(3)))
		}

		switch level {
		case logr.WarnLevel:
			l.LogAttrs(context.Background(), slog.LevelWarn, err.Error(), attrs...)
		case logr.ErrorLevel:
			l.LogAttrs(context.Background(), slog.LevelError, err.Error(), attrs...)
		default:

		}
	}
}

func (t *logger) Debug(msgOrFormat string, args ...any) {
	t.info(logr.DebugLevel, Sprintf(msgOrFormat, args...))
}

func (t *logger) Info(msgOrFormat string, args ...any) {
	t.info(logr.InfoLevel, Sprintf(msgOrFormat, args...))
}

func (t *logger) Warn(err error) {
	t.error(logr.WarnLevel, err)
}

func (t *logger) Error(err error) {
	t.error(logr.ErrorLevel, err)
}

func attrsFromKeyAndValues(name string, keysAndValues ...any) (string, []attribute.KeyValue) {
	fields := make([]attribute.KeyValue, 0, len(keysAndValues))

	i := 0
	for i < len(keysAndValues) {
		k := keysAndValues[i]

		switch key := k.(type) {
		case []attribute.KeyValue:
			fields = append(fields, key...)
			i++
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
