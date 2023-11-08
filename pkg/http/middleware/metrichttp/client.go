package metrichttp

import (
	"github.com/innoai-tech/infra/pkg/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

var (
	ClientDuration = metric.NewFloat64Histogram(
		"http.client.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Measures the duration of outbound HTTP requests"),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: DurationHistogramBoundaries,
		}),
	)

	ClientRequestSize = metric.NewInt64Histogram(
		"http.client.request.size",
		metric.WithUnit("By"),
		metric.WithDescription("Measures the size of HTTP response messages"),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: SizeHistogramBoundaries,
		}),
	)

	ClientResponseSize = metric.NewInt64Histogram(
		"http.client.response.size",
		metric.WithUnit("By"),
		metric.WithDescription("Measures the size of HTTP response messages"),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: SizeHistogramBoundaries,
		}),
	)
)
