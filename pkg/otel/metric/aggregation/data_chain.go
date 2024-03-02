package aggregation

import (
	"context"
	"time"

	"github.com/prometheus/prometheus/util/annotations"

	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb/chunkenc"
	"github.com/prometheus/prometheus/tsdb/chunks"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func NewDataChain(name string, m *metricdata.Metrics, prev *DataChain, maxDuration time.Duration) *DataChain {
	s := &DataChain{
		Name: name,
		Prev: prev,
	}

	s.dropBefore(maxDuration)
	s.copyMetrics(m)
	return s
}

type DataChain struct {
	Name       string
	Prev       *DataChain
	At         time.Time
	DataPoints []DataPoint
}

func (d *DataChain) Len() int {
	if d == nil {
		return 0
	}
	if d.Prev == nil {
		return 1
	}
	return d.Prev.Len() + 1
}

func (d *DataChain) SeriesSet() storage.SeriesSet {
	if d == nil {
		return storage.EmptySeriesSet()
	}

	if len(d.DataPoints) == 0 {
		return storage.EmptySeriesSet()
	}

	series := make([]storage.Series, len(d.DataPoints))

	n := d.Len()

	// use latest data points as entry
	for i, dp := range d.DataPoints {
		samples := make([]chunks.Sample, 0, n)

		labelMap := map[string]string{
			"__name__": d.Name,
		}
		for _, a := range dp.Attributes.ToSlice() {
			labelMap[sanitizeName(string(a.Key))] = a.Value.Emit()
		}

		for dp2 := range d.Iter(context.Background(), &dp.Attributes) {
			samples = append(samples, dp2)
		}

		reversedSamples := make([]chunks.Sample, len(samples))

		for j := range reversedSamples {
			reversedSamples[j] = samples[len(samples)-1-j]
		}

		series[i] = storage.NewListSeries(labels.FromMap(labelMap), reversedSamples)
	}

	return NewMockSeriesSet(series...)
}

var _ chunks.Sample = &DataPoint{}

type DataPoint struct {
	Attributes attribute.Set
	SampleData
}

type SampleData struct {
	valueType chunkenc.ValueType
	at        time.Time
	f         float64
	h         *histogram.Histogram
	fh        *histogram.FloatHistogram
}

func (d *SampleData) Type() chunkenc.ValueType {
	return d.valueType
}

func (d *SampleData) T() int64 {
	return d.at.UnixMilli()
}

func (d *SampleData) F() float64 {
	return d.f
}

func (d *SampleData) H() *histogram.Histogram {
	return d.h
}

func (d *SampleData) FH() *histogram.FloatHistogram {
	return d.fh
}

func (n *DataChain) dropBefore(maxDuration time.Duration) {
	breakpoint := time.Now().Add(-maxDuration)

	nn := n
	for nn != nil {
		if prev := nn.Prev; prev != nil {
			if prev.At.Before(breakpoint) {
				nn.Prev = nil
			}
		}
		nn = nn.Prev
	}
}

func (d *DataChain) Iter(ctx context.Context, matches *attribute.Set) chan *DataPoint {
	ch := make(chan *DataPoint)
	done := make(chan struct{})

	go func() {
		select {
		case <-done:
		case <-ctx.Done():
		}

		close(ch)
	}()

	go func() {
		defer close(done)

		dd := d
		for dd != nil {
			for i := range dd.DataPoints {
				dp := dd.DataPoints[i]

				if dp.Attributes.Equals(matches) {
					ch <- &dp
					continue
				}
			}

			dd = dd.Prev
		}
	}()

	return ch
}

func (d *DataChain) copyMetrics(m *metricdata.Metrics) {
	switch x := m.Data.(type) {
	case metricdata.Gauge[int64]:
		d.DataPoints = make([]DataPoint, len(x.DataPoints))
		for i, dp := range x.DataPoints {
			d.At = dp.Time
			d.DataPoints[i] = DataPoint{
				Attributes: dp.Attributes,
				SampleData: SampleData{
					at:        dp.Time,
					valueType: chunkenc.ValFloat,
					f:         float64(dp.Value),
				},
			}
		}
	case metricdata.Gauge[float64]:
		d.DataPoints = make([]DataPoint, len(x.DataPoints))
		for i, dp := range x.DataPoints {
			d.At = dp.Time
			d.DataPoints[i] = DataPoint{
				Attributes: dp.Attributes,
				SampleData: SampleData{
					at:        dp.Time,
					valueType: chunkenc.ValFloat,
					f:         dp.Value,
				},
			}
		}

	case metricdata.Sum[int64]:
		d.DataPoints = make([]DataPoint, len(x.DataPoints))
		for i, dp := range x.DataPoints {
			d.At = dp.Time
			d.DataPoints[i] = DataPoint{
				Attributes: dp.Attributes,
				SampleData: SampleData{
					at:        dp.Time,
					valueType: chunkenc.ValFloat,
					f:         float64(dp.Value),
				},
			}
		}
	case metricdata.Sum[float64]:
		d.DataPoints = make([]DataPoint, len(x.DataPoints))
		for i, dp := range x.DataPoints {
			d.At = dp.Time
			d.DataPoints[i] = DataPoint{
				Attributes: dp.Attributes,
				SampleData: SampleData{
					at:        dp.Time,
					valueType: chunkenc.ValFloat,
					f:         dp.Value,
				},
			}
		}

	case metricdata.Histogram[int64]:
		d.DataPoints = make([]DataPoint, len(x.DataPoints))
		for i, dp := range x.DataPoints {
			d.At = dp.Time
			d.DataPoints[i] = DataPoint{
				Attributes: dp.Attributes,
				SampleData: SampleData{
					at:        dp.Time,
					valueType: chunkenc.ValHistogram,
					h:         histogramFromDataPoint(&dp),
				},
			}
		}
	case metricdata.Histogram[float64]:
		d.DataPoints = make([]DataPoint, len(x.DataPoints))
		for i, dp := range x.DataPoints {
			d.At = dp.Time
			d.DataPoints[i] = DataPoint{
				Attributes: dp.Attributes,
				SampleData: SampleData{
					at:        dp.Time,
					valueType: chunkenc.ValFloatHistogram,
					h:         histogramFromDataPoint(&dp),
				},
			}
		}
	}
}

func histogramFromDataPoint[N int64 | float64](dp *metricdata.HistogramDataPoint[N]) *histogram.Histogram {
	// FIXME add buckets
	return &histogram.Histogram{
		Count: dp.Count,
		Sum:   float64(dp.Sum),
	}
}

func NewMockSeriesSet(series ...storage.Series) storage.SeriesSet {
	return &mockSeriesSet{
		idx:    -1,
		series: series,
	}
}

type mockSeriesSet struct {
	idx    int
	series []storage.Series
}

func (m *mockSeriesSet) Next() bool {
	m.idx++
	return m.idx < len(m.series)
}

func (m *mockSeriesSet) At() storage.Series { return m.series[m.idx] }

func (m *mockSeriesSet) Err() error { return nil }

func (m *mockSeriesSet) Warnings() annotations.Annotations { return nil }
