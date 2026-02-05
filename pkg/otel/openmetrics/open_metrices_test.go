package openmetrics

import (
	"bytes"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/octohelm/x/testing/snapshot"
	testingv2 "github.com/octohelm/x/testing/v2"
)

func TestWriteMetrics(t *testing.T) {
	t.Run("GIVEN histogram metrics", func(t *testing.T) {
		dp := metricdata.HistogramDataPoint[float64]{
			Attributes:   attribute.NewSet(attribute.String("env", "prod")),
			Bounds:       []float64{2.0, 4.0},
			BucketCounts: []uint64{1, 1, 1},
			Sum:          9.0,
			Count:        3,
		}

		metrics := &metricdata.ResourceMetrics{
			ScopeMetrics: []metricdata.ScopeMetrics{
				{
					Metrics: []metricdata.Metrics{
						{
							Name:        "test.histogram",
							Description: "A test histogram",
							Data: metricdata.Histogram[float64]{
								DataPoints: []metricdata.HistogramDataPoint[float64]{dp},
							},
						},
					},
				},
			},
		}

		testingv2.Then(t, "write correct metrics",
			testingv2.ExpectMustValue(
				func() (testingv2.Snapshot, error) {
					b := bytes.NewBuffer(nil)
					if err := WriteResourceMetrics(b, metrics); err != nil {
						return nil, err
					}
					return testingv2.SnapshotOf(snapshot.FileFromRaw("metrics", b.Bytes())), nil
				},
				testingv2.MatchSnapshot("histogram-metrics"),
			),
		)
	})

	t.Run("GIVEN sum metrics (monotonic)", func(t *testing.T) {
		dp := metricdata.DataPoint[int64]{
			Attributes: attribute.NewSet(attribute.String("method", "GET"), attribute.String("path", "/api")),
			Value:      1024,
		}

		metrics := &metricdata.ResourceMetrics{
			ScopeMetrics: []metricdata.ScopeMetrics{
				{
					Metrics: []metricdata.Metrics{
						{
							Name:        "http.request.count",
							Description: "Total HTTP requests",
							Data: metricdata.Sum[int64]{
								IsMonotonic: true, // 对应 Prometheus Counter
								Temporality: metricdata.CumulativeTemporality,
								DataPoints:  []metricdata.DataPoint[int64]{dp},
							},
						},
					},
				},
			},
		}

		testingv2.Then(t, "write correct counter metrics",
			testingv2.ExpectMustValue(
				func() (testingv2.Snapshot, error) {
					b := bytes.NewBuffer(nil)
					if err := WriteResourceMetrics(b, metrics); err != nil {
						return nil, err
					}
					return testingv2.SnapshotOf(snapshot.FileFromRaw("metrics", b.Bytes())), nil
				},
				testingv2.MatchSnapshot("sum-counter-metrics"),
			),
		)
	})

	t.Run("GIVEN gauge metrics", func(t *testing.T) {
		dp := metricdata.DataPoint[float64]{
			Attributes: attribute.NewSet(attribute.String("instance", "srv-01")),
			Value:      0.75,
		}

		metrics := &metricdata.ResourceMetrics{
			ScopeMetrics: []metricdata.ScopeMetrics{
				{
					Metrics: []metricdata.Metrics{
						{
							Name:        "system.cpu.usage",
							Description: "Current CPU usage percentage",
							Data: metricdata.Gauge[float64]{
								DataPoints: []metricdata.DataPoint[float64]{dp},
							},
						},
					},
				},
			},
		}

		testingv2.Then(t, "write correct gauge metrics",
			testingv2.ExpectMustValue(
				func() (testingv2.Snapshot, error) {
					b := bytes.NewBuffer(nil)
					if err := WriteResourceMetrics(b, metrics); err != nil {
						return nil, err
					}
					return testingv2.SnapshotOf(snapshot.FileFromRaw("metrics", b.Bytes())), nil
				},
				testingv2.MatchSnapshot("gauge-metrics"),
			),
		)
	})

	t.Run("GIVEN exponential histogram metrics", func(t *testing.T) {
		dp := metricdata.ExponentialHistogramDataPoint[float64]{
			Attributes: attribute.NewSet(attribute.String("type", "exponential")),
			Count:      100,
			Sum:        500.5,
			Min:        metricdata.NewExtrema(1.2),
			Max:        metricdata.NewExtrema(45.0),
		}

		metrics := &metricdata.ResourceMetrics{
			ScopeMetrics: []metricdata.ScopeMetrics{
				{
					Metrics: []metricdata.Metrics{
						{
							Name:        "test.exponential.histogram",
							Description: "An exponential histogram test",
							Data: metricdata.ExponentialHistogram[float64]{
								DataPoints: []metricdata.ExponentialHistogramDataPoint[float64]{dp},
							},
						},
					},
				},
			},
		}

		testingv2.Then(t, "write correct summary part of exponential histogram",
			testingv2.ExpectMustValue(
				func() (testingv2.Snapshot, error) {
					b := bytes.NewBuffer(nil)
					if err := WriteResourceMetrics(b, metrics); err != nil {
						return nil, err
					}
					return testingv2.SnapshotOf(snapshot.FileFromRaw("metrics", b.Bytes())), nil
				},
				testingv2.MatchSnapshot("exponential-histogram-metrics"),
			),
		)
	})
}
