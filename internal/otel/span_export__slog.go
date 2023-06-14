package otel

import (
	"context"
	"os"
	"runtime"
	_ "time/tzdata"

	"github.com/go-courier/logr"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"golang.org/x/exp/slog"
)

func Source(skip int) *slog.Source {
	pc, _, _, _ := runtime.Caller(skip)
	fs := runtime.CallersFrames([]uintptr{pc})
	f, _ := fs.Next()

	return &slog.Source{
		Function: f.Function,
		File:     f.File,
		Line:     f.Line,
	}
}

func NewLogger() *slog.Logger {
	if os.Getenv("GOENV") == "DEV" {
		return slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	return slog.New(slog.NewJSONHandler(os.Stderr, nil))
}

func SlogSpanExporter(log *slog.Logger) sdktrace.SpanExporter {
	return &stdoutSpanExporter{log: log}
}

type stdoutSpanExporter struct {
	log *slog.Logger
}

func (e *stdoutSpanExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (e *stdoutSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	for i := range spans {
		span := spans[i]

		for _, event := range span.Events() {
			attrs := make([]slog.Attr, 0, len(event.Attributes)+4)

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
					attrs = append(attrs, slog.Any(k, kv.Value.AsInterface()))
				}
			}

			spanName := span.Name()

			if spanName != "" {
				attrs = append(attrs, slog.String("traceID", span.SpanContext().TraceID().String()))

				if span.SpanContext().HasSpanID() {
					attrs = append(
						attrs,
						slog.String("spanName", spanName),
						slog.String("spanID", span.SpanContext().SpanID().String()),
						slog.String("spanCost", span.EndTime().Sub(span.StartTime()).String()),
					)
				}

				if span.Parent().IsValid() {
					attrs = append(attrs, slog.String("parentSpanID", span.Parent().SpanID().String()))
				}
			}

			msg := event.Name
			lvl := slog.LevelDebug

			switch level {
			case logr.DebugLevel:
				lvl = slog.LevelDebug
			case logr.InfoLevel:
				lvl = slog.LevelInfo
			case logr.WarnLevel:
				lvl = slog.LevelWarn
			case logr.ErrorLevel:
				lvl = slog.LevelError
			}

			r := slog.NewRecord(event.Time, lvl, msg, 0)
			r.AddAttrs(attrs...)
			_ = e.log.Handler().Handle(ctx, r)
		}
	}
	return nil
}
