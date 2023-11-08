package aggregation

import (
	"context"
	"fmt"
	"github.com/prometheus/prometheus/tsdb/chunks"
	"testing"
	"time"

	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/tsdb/chunkenc"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/storage"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestPromQL(t *testing.T) {
	now := time.Now()

	s := storage.NewListSeries(labels.FromMap(map[string]string{
		"__name__": "http_server_response_size_bytes",
	}), []chunks.Sample{
		&SampleData{
			at:        now.Add(-30 * time.Second),
			valueType: chunkenc.ValHistogram,
			h: &histogram.Histogram{
				Count: 1,
				Sum:   1000,
			},
		},
		&SampleData{
			at:        now.Add(-20 * time.Second),
			valueType: chunkenc.ValHistogram,
			h: &histogram.Histogram{
				Count: 3,
				Sum:   5000,
			},
		},
		&SampleData{
			at:        now.Add(-10 * time.Second),
			valueType: chunkenc.ValHistogram,
			h: &histogram.Histogram{
				Count: 5,
				Sum:   10000,
			},
		},
	})

	queryable := NewQueryable(NewMockSeriesSet(s))

	e := promql.NewEngine(promql.EngineOpts{
		Timeout:    time.Second * 5,
		MaxSamples: 1000,
	})
	promQL := "increase(http_server_response_size_bytes[1m])"

	query, err := e.NewInstantQuery(
		context.Background(),
		queryable,
		nil,
		promQL,
		now,
	)
	if err != nil {
		t.Error(err)
	}

	res := query.Exec(context.Background())
	if res.Err != nil {
		t.Error(err)
	}
}

func Test_histogramFromDataPoint(t *testing.T) {
	bucketCounts := make([]uint64, len(SizeHistogramBoundaries))
	bucketCounts[2] = 1

	h := &metricdata.HistogramDataPoint[int64]{
		Sum:          4934,
		Count:        1,
		Bounds:       SizeHistogramBoundaries,
		BucketCounts: bucketCounts,
	}

	fmt.Println(histogramFromDataPoint(h).String())
}

const B = 1
const KiB = 1024 * B
const MiB = 1024 * KiB

var SizeHistogramBoundaries = []float64{
	512 * B,
	1 * KiB,
	64 * KiB,
	128 * KiB,
	256 * KiB,
	512 * KiB,
	1 * MiB,
	2 * MiB,
	5 * MiB,
	10 * MiB,
	20 * MiB,
	50 * MiB,
}
