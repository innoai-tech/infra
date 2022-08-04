package otel

import (
	"context"

	"go.opentelemetry.io/otel"

	contextx "github.com/octohelm/x/context"
	"go.opentelemetry.io/otel/trace"
)

type tpCtx struct{}

func ContextWithTracerProvider(ctx context.Context, tp trace.TracerProvider) context.Context {
	return contextx.WithValue(ctx, tpCtx{}, tp)
}

func TracerProviderFromContext(ctx context.Context) trace.TracerProvider {
	if tp, ok := ctx.Value(tpCtx{}).(trace.TracerProvider); ok {
		return tp
	}
	return otel.GetTracerProvider()
}
