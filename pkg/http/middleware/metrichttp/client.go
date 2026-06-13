package metrichttp

import (
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/innoai-tech/infra/pkg/otel/metric"
)

var (
	// ClientDuration 记录出站 HTTP 请求的耗时。
	ClientDuration = metric.NewFloat64Histogram(
		"http.client.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Measures the duration of outbound HTTP requests"),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: DurationHistogramBoundaries,
		}),
	)

	// ClientRequestSize 记录出站 HTTP 请求的请求体大小。
	ClientRequestSize = metric.NewInt64Histogram(
		"http.client.request.size",
		metric.WithUnit("By"),
		metric.WithDescription("Measures the size of HTTP response messages"),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: SizeHistogramBoundaries,
		}),
	)

	// ClientResponseSize 记录出站 HTTP 请求的响应体大小。
	ClientResponseSize = metric.NewInt64Histogram(
		"http.client.response.size",
		metric.WithUnit("By"),
		metric.WithDescription("Measures the size of HTTP response messages"),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: SizeHistogramBoundaries,
		}),
	)
)
