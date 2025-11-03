package metrichttp

import (
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/innoai-tech/infra/pkg/otel/metric"
)

var (
	ServerDuration = metric.NewFloat64Histogram(
		"http.server.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Measures the duration of inbound HTTP requests"),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: DurationHistogramBoundaries,
		}),
	)

	ServerActiveRequest = metric.NewInt64Counter(
		"http.server.active_requests",
		metric.WithDescription("Measures the number of concurrent HTTP requests that are currently in-flight"),
	)

	ServerRequestSize = metric.NewInt64Histogram(
		"http.server.request.size",
		metric.WithUnit("By"),
		metric.WithDescription("Measures the size of HTTP response messages"),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: SizeHistogramBoundaries,
		}),
	)

	ServerResponseSize = metric.NewInt64Histogram(
		"http.server.response.size",
		metric.WithUnit("By"),
		metric.WithDescription("Measures the size of HTTP response messages"),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: SizeHistogramBoundaries,
		}),
	)
)
