package otel

import (
	"context"
	"go.uber.org/zap/zapcore"

	"github.com/go-courier/logr"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

func ZapSpanExporter(log zapcore.Core) sdktrace.SpanExporter {
	return &stdoutSpanExporter{log: log}
}

type stdoutSpanExporter struct {
	log zapcore.Core
}

func (e *stdoutSpanExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (e *stdoutSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	for i := range spans {
		span := spans[i]

		for _, event := range span.Events() {
			fields := make([]zap.Field, 0, len(event.Attributes)+3)

			level := logr.TraceLevel

			for _, kv := range event.Attributes {
				k := string(kv.Key)

				switch k {
				case "@level":
					lvl, err := logr.ParseLevel(kv.Value.AsString())
					if err != nil {
						continue
					}
					level = lvl
				default:
					fields = append(fields, zap.Any(k, kv.Value.AsInterface()))
				}
			}

			fields = append(fields, zap.Stringer("traceID", span.SpanContext().TraceID()))

			if span.SpanContext().HasSpanID() {
				fields = append(fields, zap.Stringer("spanID", span.SpanContext().SpanID()))
			}

			if span.Parent().IsValid() {
				fields = append(fields, zap.Stringer("parentSpanID", span.Parent().SpanID()))
			}

			entry := zapcore.Entry{}
			entry.LoggerName = span.Name()
			entry.Time = event.Time
			entry.Message = event.Name

			switch level {
			case logr.TraceLevel, logr.DebugLevel:
				entry.Level = zapcore.DebugLevel
			case logr.InfoLevel:
				entry.Level = zapcore.InfoLevel
			case logr.WarnLevel:
				entry.Level = zapcore.WarnLevel
			case logr.ErrorLevel:
				entry.Level = zapcore.ErrorLevel
			case logr.PanicLevel:
				entry.Level = zapcore.PanicLevel
			case logr.FatalLevel:
				entry.Level = zapcore.FatalLevel
			}

			if err := e.log.Write(entry, fields); err != nil {
				return err
			}
		}
	}

	_ = e.log.Sync()
	return nil
}
