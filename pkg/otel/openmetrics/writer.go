package openmetrics

import (
	"bytes"
	"fmt"
	"io"
	"iter"
	"strconv"

	"github.com/prometheus/otlptranslator"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	syncx "github.com/octohelm/x/sync"
)

const (
	ContentType = "application/openmetrics-text; version=1.0.0; charset=utf-8"
)

func WriteResourceMetrics(w io.Writer, metrics *metricdata.ResourceMetrics) error {
	translationStrategy := otlptranslator.UnderscoreEscapingWithSuffixes
	ow := &openMetricsWriter{
		Writer:      w,
		metricNamer: otlptranslator.NewMetricNamer("", translationStrategy),
		labelNamer:  otlptranslator.LabelNamer{UTF8Allowed: !translationStrategy.ShouldEscape()},
		unitNamer:   otlptranslator.UnitNamer{UTF8Allowed: !translationStrategy.ShouldEscape()},
	}
	return ow.WriteResourceMetrics(metrics)
}

type openMetricsWriter struct {
	io.Writer

	metricNamer otlptranslator.MetricNamer
	labelNamer  otlptranslator.LabelNamer
	unitNamer   otlptranslator.UnitNamer

	metrics syncx.Map[string, struct{}]
}

func (c *openMetricsWriter) WriteResourceMetrics(metrics *metricdata.ResourceMetrics) error {
	for _, scopeMetrics := range metrics.ScopeMetrics {
		for _, m := range scopeMetrics.Metrics {
			typ := c.metricType(m)
			if typ == otlptranslator.MetricTypeUnknown {
				continue
			}
			name, e := c.getName(m, typ)
			if e != nil {
				continue
			}

			switch v := m.Data.(type) {
			case metricdata.Histogram[int64]:
				if err := writeHistogramMetric(c.Writer, v, m, name, c.labelNamer); err != nil {
					return err
				}
			case metricdata.Histogram[float64]:
				if err := writeHistogramMetric(c.Writer, v, m, name, c.labelNamer); err != nil {
					return err
				}
			case metricdata.ExponentialHistogram[int64]:
				if err := writeExponentialHistogramMetric(c.Writer, v, m, name, c.labelNamer); err != nil {
					return err
				}
			case metricdata.ExponentialHistogram[float64]:
				if err := writeExponentialHistogramMetric(c.Writer, v, m, name, c.labelNamer); err != nil {
					return err
				}
			case metricdata.Sum[int64]:
				if err := writeSumMetric(c.Writer, v, m, name, c.labelNamer); err != nil {
					return err
				}
			case metricdata.Sum[float64]:
				if err := writeSumMetric(c.Writer, v, m, name, c.labelNamer); err != nil {
					return err
				}
			case metricdata.Gauge[int64]:
				if err := writeGaugeMetric(c.Writer, v, m, name, c.labelNamer); err != nil {
					return err
				}
			case metricdata.Gauge[float64]:
				if err := writeGaugeMetric(c.Writer, v, m, name, c.labelNamer); err != nil {
					return err
				}
			}
		}
	}

	_, err := c.Write([]byte("# EOF\n"))
	return err
}

func (c *openMetricsWriter) getName(m metricdata.Metrics, mt otlptranslator.MetricType) (string, error) {
	return c.metricNamer.Build(otlptranslator.Metric{
		Name: m.Name,
		Type: mt,
	})
}

func (c *openMetricsWriter) metricType(m metricdata.Metrics) otlptranslator.MetricType {
	switch v := m.Data.(type) {
	case metricdata.ExponentialHistogram[int64], metricdata.ExponentialHistogram[float64]:
		return otlptranslator.MetricTypeHistogram
	case metricdata.Histogram[int64], metricdata.Histogram[float64]:
		return otlptranslator.MetricTypeHistogram
	case metricdata.Sum[float64]:
		if v.IsMonotonic {
			return otlptranslator.MetricTypeMonotonicCounter
		}
		return otlptranslator.MetricTypeNonMonotonicCounter
	case metricdata.Sum[int64]:
		if v.IsMonotonic {
			return otlptranslator.MetricTypeMonotonicCounter
		}
		return otlptranslator.MetricTypeNonMonotonicCounter
	case metricdata.Gauge[int64], metricdata.Gauge[float64]:
		return otlptranslator.MetricTypeGauge
	case metricdata.Summary:
		return otlptranslator.MetricTypeSummary
	}
	return otlptranslator.MetricTypeUnknown
}

func writeHistogramMetric[N int64 | float64](w io.Writer, v metricdata.Histogram[N], m metricdata.Metrics, name string, namer otlptranslator.LabelNamer) error {
	mw := &metricWriter{
		w:          w,
		labelNamer: namer,
	}

	if err := mw.WriteHead("histogram", name, m.Description); err != nil {
		return err
	}

	for _, dp := range v.DataPoints {

		// 1. 渲染累加桶 (_bucket)
		if len(dp.Bounds) > 0 {
			var cumulative uint64
			for i, bound := range dp.Bounds {
				cumulative += dp.BucketCounts[i]

				if err := mw.WriteDataPoint(name+"_bucket", appendLe(attrsFromSet(dp.Attributes), strconv.FormatFloat(bound, 'g', -1, 64)), cumulative); err != nil {
					return err
				}
			}

			cumulative += dp.BucketCounts[len(dp.Bounds)]

			if err := mw.WriteDataPoint(name+"_bucket", appendLe(attrsFromSet(dp.Attributes), "+Inf"), cumulative); err != nil {
				return err
			}
		}

		if err := mw.WriteDataPoint(name+"_sum", attrsFromSet(dp.Attributes), dp.Sum); err != nil {
			return err
		}
		if err := mw.WriteDataPoint(name+"_count", attrsFromSet(dp.Attributes), dp.Count); err != nil {
			return err
		}
	}

	return nil
}

func writeExponentialHistogramMetric[N int64 | float64](w io.Writer, v metricdata.ExponentialHistogram[N], m metricdata.Metrics, name string, namer otlptranslator.LabelNamer) error {
	mw := &metricWriter{
		w:          w,
		labelNamer: namer,
	}

	if err := mw.WriteHead("histogram", name, m.Description); err != nil {
		return err
	}

	for _, dp := range v.DataPoints {
		if err := mw.WriteDataPoint(name+"_sum", attrsFromSet(dp.Attributes), dp.Sum); err != nil {
			return err
		}
		if err := mw.WriteDataPoint(name+"_count", attrsFromSet(dp.Attributes), dp.Count); err != nil {
			return err
		}
	}

	return nil
}

func writeGaugeMetric[N int64 | float64](w io.Writer, v metricdata.Gauge[N], m metricdata.Metrics, name string, namer otlptranslator.LabelNamer) error {
	mw := &metricWriter{
		w:          w,
		labelNamer: namer,
	}

	if err := mw.WriteHead("gauge", name, m.Description); err != nil {
		return err
	}

	for _, dp := range v.DataPoints {
		if err := mw.WriteDataPoint(name, attrsFromSet(dp.Attributes), dp.Value); err != nil {
			return err
		}
	}

	return nil
}

func writeSumMetric[N int64 | float64](w io.Writer, v metricdata.Sum[N], m metricdata.Metrics, name string, namer otlptranslator.LabelNamer) error {
	typ := "counter"
	if !v.IsMonotonic {
		typ = "gauge"
	}

	mw := &metricWriter{
		w:          w,
		labelNamer: namer,
	}

	if err := mw.WriteHead(typ, name, m.Description); err != nil {
		return err
	}

	for _, dp := range v.DataPoints {
		if err := mw.WriteDataPoint(name, attrsFromSet(dp.Attributes), dp.Value); err != nil {
			return err
		}
	}
	return nil
}

type metricWriter struct {
	w io.Writer

	labelNamer otlptranslator.LabelNamer
}

func (m *metricWriter) WriteHead(typ string, name string, description string) error {
	if _, err := fmt.Fprintf(m.w, "# HELP %s %s\n# TYPE %s %s\n", name, description, name, typ); err != nil {
		return err
	}
	return nil
}

func appendLe(seq iter.Seq2[string, string], leValue string) iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		seq(yield)

		if !yield("le", leValue) {
			return
		}
	}
}

func attrsFromSet(s attribute.Set) iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for i := s.Iter(); i.Next(); {
			kv := i.Attribute()
			if !yield(string(kv.Key), kv.Value.Emit()) {
				return
			}
		}
	}
}

func (m *metricWriter) WriteDataPoint(name string, attrs iter.Seq2[string, string], data any) error {
	if _, err := fmt.Fprint(m.w, name); err != nil {
		return err
	}

	labels := bytes.NewBuffer(nil)
	first := true
	for k, val := range attrs {
		if !first {
			if _, err := fmt.Fprint(labels, ","); err != nil {
				return err
			}
		}

		key, err := m.labelNamer.Build(k)
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintf(labels, "%s=%q", key, val); err != nil {
			return err
		}

		first = false
	}

	if labels.Len() > 0 {
		if _, err := fmt.Fprintf(m.w, "{%s}", labels.Bytes()); err != nil {
			return err
		}
	}

	var valStr string
	switch v := data.(type) {
	case float64:
		valStr = strconv.FormatFloat(v, 'g', -1, 64)
	case float32:
		valStr = strconv.FormatFloat(float64(v), 'g', -1, 32)
	case int, int64, uint64, int32, uint32:
		valStr = fmt.Sprintf("%v", v)
	default:
		valStr = fmt.Sprintf("%v", v)
	}

	if _, err := fmt.Fprintf(m.w, " %s\n", valStr); err != nil {
		return err
	}

	return nil
}
