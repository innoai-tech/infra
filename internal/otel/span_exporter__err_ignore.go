package otel

import (
	"context"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func WithErrIgnoreExporter() func(spanExporter sdktrace.SpanExporter) sdktrace.SpanExporter {
	return func(spanExporter sdktrace.SpanExporter) sdktrace.SpanExporter {
		return &errIgnoreExporter{
			SpanExporter: spanExporter,
		}
	}
}

type errIgnoreExporter struct {
	sdktrace.SpanExporter
}

func (e *errIgnoreExporter) ExportSpans(ctx context.Context, spanData []sdktrace.ReadOnlySpan) error {
	_ = e.SpanExporter.ExportSpans(ctx, spanData)
	return nil
}
