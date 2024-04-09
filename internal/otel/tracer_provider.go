package otel

import (
	"context"
	contextx "github.com/octohelm/x/context"
	"go.opentelemetry.io/otel/trace"
)

var TracerProviderContext = contextx.New[TracerProvider]()

type TracerProvider = trace.TracerProvider

func Tracer(ctx context.Context) trace.Tracer {
	return TracerProviderContext.From(ctx).Tracer("")
}
