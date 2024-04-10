package otel

import (
	"context"
	"fmt"
	"time"

	"github.com/go-courier/logr"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"
)

func NewLogger(ctx context.Context, levelEnabled logr.Level) logr.Logger {
	return &logger{
		loggerContext: loggerContext{
			ctx:     ctx,
			enabled: levelEnabled,
		},
	}
}

type logger struct {
	spanContext
	loggerContext

	keyValues []log.KeyValue
}

func (t *logger) WithValues(keyAndValues ...any) logr.Logger {
	if len(keyAndValues) == 0 {
		return t
	}

	return &logger{
		spanContext:   t.spanContext,
		loggerContext: t.loggerContext,
		keyValues:     append(t.keyValues, normalizeKeyValues(keyAndValues)...),
	}
}

func (t *logger) Start(ctx context.Context, name string, keyAndValues ...any) (context.Context, logr.Logger) {
	var parentID trace.SpanID

	parentSpan := trace.SpanContextFromContext(ctx)
	if parentSpan.HasSpanID() {
		parentID = parentSpan.SpanID()
	}

	spanCtx, c := t.spanContext.Start(ctx, name)

	lgr := &logger{
		keyValues:     append(t.keyValues, normalizeKeyValues(keyAndValues)...),
		spanContext:   spanCtx,
		loggerContext: t.loggerContext.Start(c, name, parentID),
	}

	return logr.WithLogger(c, lgr), lgr
}

func (t *logger) End() {
	endAt := time.Now()
	t.span(func(s trace.Span) {
		s.End(trace.WithTimestamp(endAt))
	})
}

func (t *logger) span(do func(s trace.Span)) {
	if span := t.spanContext.span; span != nil {
		do(span)
	}
}

func (t *logger) Debug(msgOrFormat string, args ...any) {
	t.info(logr.DebugLevel, sprintf(msgOrFormat, args...), t.keyValues)
}

func (t *logger) Info(msgOrFormat string, args ...any) {
	t.info(logr.InfoLevel, sprintf(msgOrFormat, args...), t.keyValues)
}

func (t *logger) Warn(err error) {
	t.error(logr.WarnLevel, err, t.keyValues, func(err error) {
		errMsg := err.Error()
		t.span(func(s trace.Span) {
			s.RecordError(err)
			s.SetStatus(codes.Error, errMsg)
		})
	})
}

func (t *logger) Error(err error) {
	t.error(logr.ErrorLevel, err, t.keyValues, func(err error) {
		errMsg := err.Error()
		t.span(func(s trace.Span) {
			s.RecordError(err)
			s.SetStatus(codes.Error, errMsg)
		})
	})
}

type loggerContext struct {
	ctx     context.Context
	enabled logr.Level

	log.Logger

	startedAt time.Time
	parentID  trace.SpanID
}

func (l *loggerContext) emit(level logr.Level, msg fmt.Stringer, keyValues []log.KeyValue) {
	if l.Logger == nil {
		return
	}

	var rec log.Record

	switch level {
	case logr.DebugLevel:
		rec.SetSeverity(log.SeverityDebug)
	case logr.InfoLevel:
		rec.SetSeverity(log.SeverityInfo)
	case logr.WarnLevel:
		rec.AddAttributes(GetSource(3).AsKeyValues()...)
		rec.SetSeverity(log.SeverityWarn)
	case logr.ErrorLevel:
		rec.AddAttributes(GetSource(3).AsKeyValues()...)
		rec.SetSeverity(log.SeverityError)
	}

	if len(keyValues) > 0 {
		rec.AddAttributes(keyValues...)
	}

	if !l.startedAt.IsZero() {
		rec.AddAttributes(log.String("cost", time.Since(l.startedAt).String()))
	}

	if l.parentID.IsValid() {
		rec.AddAttributes(log.String("parentSpanID", l.parentID.String()))
	}

	rec.SetTimestamp(time.Now())
	rec.SetBody(log.StringValue(msg.String()))

	l.Emit(l.ctx, rec)
}

func (l *loggerContext) info(level logr.Level, msg fmt.Stringer, keyValues []log.KeyValue) {
	if level > l.enabled {
		return
	}

	l.emit(level, msg, keyValues)
}

func (l *loggerContext) error(level logr.Level, err error, keyValues []log.KeyValue, postDo func(err error)) {
	if level > l.enabled {
		return
	}

	if err == nil {
		return
	}

	l.emit(level, sprintf("%s", err), keyValues)

	postDo(err)
}

func (l loggerContext) Start(ctx context.Context, name string, parentID trace.SpanID) loggerContext {
	l.ctx = ctx
	l.Logger = Logger(ctx, name)
	l.startedAt = time.Now()
	l.parentID = parentID
	return l
}

type spanContext struct {
	name   string
	span   trace.Span
	tracer trace.Tracer
}

func (c spanContext) Start(ctx context.Context, spanName string) (spanContext, context.Context) {
	cc, span := Tracer(ctx).Start(ctx, c.name, trace.WithTimestamp(time.Now()))
	c.span = span
	c.name = spanName
	return c, cc
}

func sprintf(format string, args ...any) fmt.Stringer {
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
