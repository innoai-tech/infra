package otel

import (
	"context"
	"os"
	_ "time/tzdata"

	"github.com/go-courier/logr"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapCore() zapcore.Core {
	if os.Getenv("GOENV") == "DEV" {
		return zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			os.Stdout,
			zap.DebugLevel,
		)
	}

	c := zap.NewProductionEncoderConfig()
	c.EncodeTime = zapcore.ISO8601TimeEncoder

	return zapcore.NewCore(
		zapcore.NewJSONEncoder(c),
		os.Stdout,
		zap.DebugLevel,
	)
}

func ZapSpanExporter(log zapcore.Core) sdktrace.SpanExporter {
	return &stdoutSpanExporter{log: log}
}

type stdoutSpanExporter struct {
	log zapcore.Core
}

func (e *stdoutSpanExporter) Shutdown(ctx context.Context) error {
	_ = e.log.Sync()
	return nil
}

func (e *stdoutSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	for i := range spans {
		span := spans[i]

		for _, event := range span.Events() {
			fields := make([]zap.Field, 0, len(event.Attributes)+4)

			level := logr.DebugLevel

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

			spanName := span.Name()

			if spanName != "" {
				fields = append(fields, zap.Stringer("traceID", span.SpanContext().TraceID()))

				if span.SpanContext().HasSpanID() {
					fields = append(
						fields,
						zap.Stringer("spanID", span.SpanContext().SpanID()),
						zap.Stringer("spanCost", span.EndTime().Sub(span.StartTime())),
					)
				}

				if span.Parent().IsValid() {
					fields = append(fields, zap.Stringer("parentSpanID", span.Parent().SpanID()))
				}
			}

			entry := zapcore.Entry{}
			entry.LoggerName = spanName
			entry.Message = event.Name
			entry.Time = event.Time

			switch level {
			case logr.DebugLevel:
				entry.Level = zapcore.DebugLevel
			case logr.InfoLevel:
				entry.Level = zapcore.InfoLevel
			case logr.WarnLevel:
				entry.Level = zapcore.WarnLevel
			case logr.ErrorLevel:
				entry.Level = zapcore.ErrorLevel
			}

			if err := e.log.Write(entry, fields); err != nil {
				return err
			}
		}
	}
	return nil
}
