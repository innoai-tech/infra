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
		tp:      tp,
		span:    span,
		enabled: levelEnabled,
		slog:    slog,
	}
}

type spanLogger struct {
	enabled    logr.Level
	tp         trace.TracerProvider
	spanName   string
	span       trace.Span
	attributes []attribute.KeyValue
	slog       *slog.Logger
}

func (t *spanLogger) withName(name string) *spanLogger {
	return &spanLogger{
		tp:         t.tp,
		slog:       t.slog,
		span:       t.span,
		spanName:   name,
		enabled:    t.enabled,
		attributes: t.attributes,
	}
}

func (t *spanLogger) WithValues(keyAndValues ...any) logr.Logger {
	if len(keyAndValues) == 0 {
		return t
	}

	name, attrs := attrsFromKeyAndValues(t.spanName, keyAndValues...)

	return &spanLogger{
		tp:         t.tp,
		slog:       t.slog,
		span:       t.span,
		enabled:    t.enabled,
		spanName:   name,
		attributes: append(t.attributes, attrs...),
	}
}

func (t *spanLogger) start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tp.Tracer("").Start(ctx, spanName, opts...)
}

func (t *spanLogger) Start(ctx context.Context, name string, keyAndValues ...any) (context.Context, logr.Logger) {
	name = appendName(t.spanName, name)

	n, attrs := attrsFromKeyAndValues(name, keyAndValues...)

	c, span := t.start(
		ctx,
		n,
		trace.WithAttributes(attrs...),
		trace.WithTimestamp(time.Now()),
	)

	lgr := &spanLogger{
		enabled:    t.enabled,
		tp:         t.tp,
		slog:       t.slog,
		span:       span,
		spanName:   name,
		attributes: attrs,
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
	t.span.End(trace.WithTimestamp(time.Now()))
}

func (t *spanLogger) info(level logr.Level, msg fmt.Stringer) {
	if level > t.enabled {
		return
	}

	span := t.span

	if span == nil {
		_, span = t.start(context.Background(), "")
		defer span.End()
	}

	attributes := append(t.attributes, attribute.String("@level", level.String()))

	msgStr := msg.String()

	span.AddEvent(
		msgStr,
		trace.WithTimestamp(time.Now()),
		trace.WithAttributes(attributes...),
	)

	if l := t.slog; l != nil {
		attrs := make([]slog.Attr, len(t.attributes))
		for i := range attrs {
			a := t.attributes[i]
			attrs[i] = slog.Any(string(a.Key), a.Value.AsInterface())
		}

		spanContext := span.SpanContext()
		attrs = append(attrs, slog.String("traceID", spanContext.TraceID().String()))

		if spanContext.HasSpanID() {
			attrs = append(
				attrs,
				slog.String("spanID", spanContext.SpanID().String()),
			)
		}

		if t.spanName != "" {
			attrs = append(attrs, slog.String("spanName", t.spanName))
		}

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

	span := t.span
	if span == nil {
		_, span = t.start(context.Background(), "")
		defer span.End()
	}

	attributes := t.attributes

	if level <= logr.ErrorLevel {
		attributes = append(attributes, attribute.String("exception.stack", fmt.Sprintf("%+v", err)))
	}

	span.RecordError(err,
		trace.WithTimestamp(time.Now()),
		trace.WithAttributes(append(attributes, attribute.String("@level", level.String()))...),
	)

	if l := t.slog; l != nil {
		attrs := make([]slog.Attr, len(attributes))
		for i := range attrs {
			a := attributes[i]
			attrs[i] = slog.Any(string(a.Key), a.Value.AsInterface())
		}

		spanContext := span.SpanContext()
		attrs = append(attrs, slog.String("traceID", spanContext.TraceID().String()))

		if spanContext.HasSpanID() {
			attrs = append(
				attrs,
				slog.String("spanID", spanContext.SpanID().String()),
			)
		}

		attrs = append(attrs, slog.String("exception.message", err.Error()))

		if t.spanName != "" {
			attrs = append(attrs, slog.String("spanName", t.spanName))
		}

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
	n := len(keysAndValues)
	if n > 0 && n%2 == 0 {
		fields := make([]attribute.KeyValue, len(keysAndValues)/2)
		for i := range fields {
			k, v := keysAndValues[2*i], keysAndValues[2*i+1]

			if k == "@spanName" {
				name = appendName(name, v.(string))
				continue
			}

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
		return name, fields
	}
	return name, nil
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
