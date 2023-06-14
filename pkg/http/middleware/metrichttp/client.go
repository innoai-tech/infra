package metrichttp

import (
	"github.com/innoai-tech/infra/pkg/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
)

var (
	ClientDuration = metric.NewFloat64Histogram(
		"http.client.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Measures the duration of outbound HTTP requests"),
		metric.WithAggregation(aggregation.ExplicitBucketHistogram{
			Boundaries: DurationHistogramBoundaries,
		}),
	)

	ClientRequestSize = metric.NewInt64Histogram(
		"http.client.request.size",
		metric.WithUnit("By"),
		metric.WithDescription("Measures the size of HTTP response messages"),
		metric.WithAggregation(aggregation.ExplicitBucketHistogram{
			Boundaries: SizeHistogramBoundaries,
		}),
	)

	ClientResponseSize = metric.NewInt64Histogram(
		"http.client.response.size",
		metric.WithUnit("By"),
		metric.WithDescription("Measures the size of HTTP response messages"),
		metric.WithAggregation(aggregation.ExplicitBucketHistogram{
			Boundaries: SizeHistogramBoundaries,
		}),
	)
)
