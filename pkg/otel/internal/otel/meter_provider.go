package otel

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	contextx "github.com/octohelm/x/context"
)

// MeterProvider 是对 OpenTelemetry metric.MeterProvider 的类型别名。
// MetricReader 是对 OpenTelemetry sdkmetric.Reader 的类型别名。
type (
	MeterProvider = metric.MeterProvider
	MetricReader  = sdkmetric.Reader
)

// MetricReaderContext 是用于上下文注入的 MetricReader 上下文键。
var MetricReaderContext = contextx.New[MetricReader]()

// MeterProviderContext 是用于上下文注入的 MeterProvider 上下文键。
var MeterProviderContext = contextx.New[MeterProvider](contextx.WithDefaultsFunc(func() metric.MeterProvider {
	return noop.NewMeterProvider()
}))

// Meter 从上下文中获取 OpenTelemetry Meter 实例。
func Meter(ctx context.Context) metric.Meter {
	return MeterProviderContext.From(ctx).Meter("")
}
