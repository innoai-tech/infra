package otel

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type OutputFilterType string

var (
	OutputFilterAlways    OutputFilterType = "Always"
	OutputFilterOnFailure OutputFilterType = "OnFailure"
	OutputFilterNever     OutputFilterType = "Never"
)

func OutputFilter(outputType OutputFilterType) SpanMapper {
	return func(span sdktrace.ReadOnlySpan) sdktrace.ReadOnlySpan {
		switch outputType {
		case OutputFilterOnFailure:
			if span.Status().Code == codes.Ok {
				return nil
			}
		case OutputFilterNever:
			return nil
		}
		return span
	}
}

type SpanMapper = func(data sdktrace.ReadOnlySpan) sdktrace.ReadOnlySpan

func WithSpanMapExporter(mappers ...SpanMapper) func(spanExporter sdktrace.SpanExporter) sdktrace.SpanExporter {
	return func(spanExporter sdktrace.SpanExporter) sdktrace.SpanExporter {
		return &spanMapExporter{
			mappers:      mappers,
			SpanExporter: spanExporter,
		}
	}
}

type spanMapExporter struct {
	mappers []SpanMapper
	sdktrace.SpanExporter
}

func (e *spanMapExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	finalSpanSnapshot := make([]sdktrace.ReadOnlySpan, 0, len(spans))

	mappers := e.mappers

	for i := range spans {
		span := spans[i]

		for _, m := range mappers {
			span = m(span)
		}

		if span != nil {
			finalSpanSnapshot = append(finalSpanSnapshot, span)
		}
	}

	if len(finalSpanSnapshot) == 0 {
		return nil
	}

	return e.SpanExporter.ExportSpans(ctx, finalSpanSnapshot)
}
