package otel

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

var emptyMetricProvider = noop.NewMeterProvider()

type contextMetricProvider struct {
}

func MeterProviderFromContext(ctx context.Context) metric.MeterProvider {
	if mp, ok := ctx.Value(contextMetricProvider{}).(metric.MeterProvider); ok {
		return mp
	}
	return emptyMetricProvider
}

func ContextWithMeterProvider(ctx context.Context, meterProvider metric.MeterProvider) context.Context {
	return context.WithValue(ctx, contextMetricProvider{}, meterProvider)
}

type contextGatherer struct {
}

func GathererFromContext(ctx context.Context) prometheus.Gatherer {
	if mp, ok := ctx.Value(contextGatherer{}).(prometheus.Gatherer); ok {
		return mp
	}
	return prometheus.DefaultGatherer
}

func ContextWithGatherer(ctx context.Context, gatherer prometheus.Gatherer) context.Context {
	return context.WithValue(ctx, contextGatherer{}, gatherer)
}
