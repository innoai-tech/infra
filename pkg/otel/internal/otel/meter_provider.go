package otel

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	contextx "github.com/octohelm/x/context"
)

type (
	MeterProvider = metric.MeterProvider
	MetricReader  = sdkmetric.Reader
)

var MetricReaderContext = contextx.New[MetricReader]()

var MeterProviderContext = contextx.New[MeterProvider](contextx.WithDefaultsFunc(func() metric.MeterProvider {
	return noop.NewMeterProvider()
}))

func Meter(ctx context.Context) metric.Meter {
	return MeterProviderContext.From(ctx).Meter("")
}
