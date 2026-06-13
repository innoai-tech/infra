package metrichttp

import (
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/innoai-tech/infra/pkg/otel/metric"
)

var (
	// ServerDuration 记录入站 HTTP 请求的耗时。
	ServerDuration = metric.NewFloat64Histogram(
		"http.server.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Measures the duration of inbound HTTP requests"),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: DurationHistogramBoundaries,
		}),
	)

	// ServerActiveRequest 记录当前正在进行中的 HTTP 请求并发数。
	ServerActiveRequest = metric.NewInt64Counter(
		"http.server.active_requests",
		metric.WithDescription("Measures the number of concurrent HTTP requests that are currently in-flight"),
	)

	// ServerRequestSize 记录入站 HTTP 请求的请求体大小。
	ServerRequestSize = metric.NewInt64Histogram(
		"http.server.request.size",
		metric.WithUnit("By"),
		metric.WithDescription("Measures the size of HTTP response messages"),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: SizeHistogramBoundaries,
		}),
	)

	// ServerResponseSize 记录入站 HTTP 请求的响应体大小。
	ServerResponseSize = metric.NewInt64Histogram(
		"http.server.response.size",
		metric.WithUnit("By"),
		metric.WithDescription("Measures the size of HTTP response messages"),
		metric.WithAggregation(sdkmetric.AggregationExplicitBucketHistogram{
			Boundaries: SizeHistogramBoundaries,
		}),
	)
)
