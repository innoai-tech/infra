package otel

import (
	"context"

	contextx "github.com/octohelm/x/context"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type (
	MeterProvider = metric.MeterProvider
	Reader        = sdkmetric.Reader
)

var GathererContext = contextx.New[prometheus.Gatherer]()

var MeterProviderContext = contextx.New[MeterProvider](contextx.WithDefaultsFunc(func() metric.MeterProvider {
	return noop.NewMeterProvider()
}))

func Meter(ctx context.Context) metric.Meter {
	return MeterProviderContext.From(ctx).Meter("")
}
