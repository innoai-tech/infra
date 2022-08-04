package otel

import (
	"context"

	"github.com/go-courier/logr"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

func ZapSpanExporter(log *zap.Logger) sdktrace.SpanExporter {
	return &stdoutSpanExporter{log: log}
}

type stdoutSpanExporter struct {
	log *zap.Logger
}

func (e *stdoutSpanExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (e *stdoutSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	for i := range spans {
		span := spans[i]

		for _, event := range span.Events() {
			l := e.log.Named(span.Name())

			fields := []zap.Field{
				zap.Time("ts", event.Time),
			}

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

			switch level {
			case logr.TraceLevel:
				l.Debug(event.Name, fields...)
			case logr.DebugLevel:
				l.Debug(event.Name, fields...)
			case logr.InfoLevel:
				l.Info(event.Name, fields...)
			case logr.WarnLevel:
				l.Warn(event.Name, fields...)
			case logr.ErrorLevel:
				l.Error(event.Name, fields...)
			case logr.PanicLevel:
				l.DPanic(event.Name, fields...)
			case logr.FatalLevel:
				l.Fatal(event.Name, fields...)
			}
		}
	}

	return nil
}
