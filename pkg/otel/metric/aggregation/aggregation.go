package aggregation

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql"
	"go.opentelemetry.io/otel/attribute"

	"github.com/innoai-tech/infra/pkg/otel/metric"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func NewReader(views []metric.View, getMeter func() otelmetric.Meter) (sdkmetric.Reader, error) {
	a := &aggregationOverTime{
		temporalitySelector: sdkmetric.DefaultTemporalitySelector,
		aggregationSelector: sdkmetric.DefaultAggregationSelector,
		collectors:          map[string]*collector{},
	}

	for _, v := range views {
		if v.Instrument.Kind == sdkmetric.InstrumentKindObservableGauge {
			baseAndTypeRule := strings.SplitN(v.Instrument.Name, "__", 2)
			if len(baseAndTypeRule) == 2 {
				typeAndOverTime := strings.SplitN(baseAndTypeRule[1], ".", 2)
				if len(typeAndOverTime) == 2 {
					typ := typeAndOverTime[0]
					d, _ := time.ParseDuration(typeAndOverTime[1])
					c := &collector{
						Instrument: v.Instrument,
						Base:       baseAndTypeRule[0],
						Type:       typ,
						Duration:   d,
						GetMeter:   getMeter,
					}

					a.collectors[c.Base] = c
				}

			}
		}
	}

	return sdkmetric.NewPeriodicReader(a, sdkmetric.WithInterval(10*time.Second)), nil
}

type aggregationOverTime struct {
	temporalitySelector sdkmetric.TemporalitySelector
	aggregationSelector sdkmetric.AggregationSelector
	collectors          map[string]*collector
}

func (e *aggregationOverTime) Temporality(kind sdkmetric.InstrumentKind) metricdata.Temporality {
	return e.temporalitySelector(kind)
}

func (e *aggregationOverTime) Aggregation(kind sdkmetric.InstrumentKind) aggregation.Aggregation {
	return e.aggregationSelector(kind)
}

func (e *aggregationOverTime) Export(ctx context.Context, metrics *metricdata.ResourceMetrics) error {
	for i := range metrics.ScopeMetrics {
		scopeMetrics := metrics.ScopeMetrics[i]

		for j := range scopeMetrics.Metrics {
			m := scopeMetrics.Metrics[j]

			if c, ok := e.collectors[m.Name]; ok {
				if err := c.Collect(ctx, &m); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (e *aggregationOverTime) ForceFlush(ctx context.Context) error {
	return nil
}

func (e *aggregationOverTime) Shutdown(ctx context.Context) error {
	return nil
}

type collector struct {
	Instrument sdkmetric.Instrument
	Base       string
	Type       string
	Duration   time.Duration

	GetMeter func() otelmetric.Meter
	initOnce sync.Once

	data *DataChain
}

func (c *collector) ToMetric() metricdata.Metrics {
	return metricdata.Metrics{
		Name: c.Base,
		Unit: c.Instrument.Unit,
	}
}

var unitSuffixes = map[string]string{
	"1":  "_ratio",
	"By": "_bytes",
	"ms": "_milliseconds",
}

func (c *collector) getName(m metricdata.Metrics) string {
	name := sanitizeName(m.Name)
	if suffix, ok := unitSuffixes[m.Unit]; ok {
		name += suffix
	}
	return name
}

func sanitizeName(n string) string {
	// This algorithm is based on strings.Map from Go 1.19.
	const replacement = '_'

	valid := func(i int, r rune) bool {
		// Taken from
		// https://github.com/prometheus/common/blob/dfbc25bd00225c70aca0d94c3c4bb7744f28ace0/model/metric.go#L92-L102
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || r == ':' || (r >= '0' && r <= '9' && i > 0) {
			return true
		}
		return false
	}

	// This output buffer b is initialized on demand, the first time a
	// character needs to be replaced.
	var b strings.Builder
	for i, c := range n {
		if valid(i, c) {
			continue
		}

		if i == 0 && c >= '0' && c <= '9' {
			// Prefix leading number with replacement character.
			b.Grow(len(n) + 1)
			_ = b.WriteByte(byte(replacement))
			break
		}
		b.Grow(len(n))
		_, _ = b.WriteString(n[:i])
		_ = b.WriteByte(byte(replacement))
		width := utf8.RuneLen(c)
		n = n[i+width:]
		break
	}

	// Fast path for unchanged input.
	if b.Cap() == 0 { // b.Grow was not called above.
		return n
	}

	for _, c := range n {
		// Due to inlining, it is more performant to invoke WriteByte rather then
		// WriteRune.
		if valid(1, c) { // We are guaranteed to not be at the start.
			_ = b.WriteByte(byte(c))
		} else {
			_ = b.WriteByte(byte(replacement))
		}
	}

	return b.String()
}

func (c *collector) init() {
	m := c.GetMeter()

	name := c.getName(c.ToMetric())

	_, err := m.Float64ObservableGauge(
		fmt.Sprintf("%s__%s_%ds", name, c.Type, int64(c.Duration.Seconds())),
		otelmetric.WithUnit(c.Type),
		otelmetric.WithFloat64Callback(func(ctx context.Context, observer otelmetric.Float64Observer) error {
			if d := c.data; d != nil {
				queryable := NewQueryable(d.SeriesSet())
				e := promql.NewEngine(promql.EngineOpts{
					Timeout:    time.Second * 5,
					MaxSamples: d.Len() * len(d.DataPoints) * 1000,
				})

				promQL := fmt.Sprintf("%s(%s[%ds])", c.Type, name, int64(c.Duration.Seconds()))
				query, err := e.NewInstantQuery(
					ctx,
					queryable,
					nil,
					promQL,
					time.Now(),
				)
				if err != nil {
					return err
				}

				res := query.Exec(ctx)
				vec, err := res.Vector()
				if err == nil {
					for _, s := range vec {
						if s.H == nil {
							observer.Observe(s.F, otelmetric.WithAttributes(attrsFromLabels(s.Metric)...))
						} else {
							observer.Observe(s.H.Sum, otelmetric.WithAttributes(attrsFromLabels(s.Metric)...))
						}
					}
				}

			}
			return nil
		}),
	)

	if err != nil {
		panic(err)
	}
}

func (c *collector) Collect(ctx context.Context, m *metricdata.Metrics) error {
	c.initOnce.Do(func() {
		c.init()
	})
	c.data = NewDataChain(c.Base, m, c.data, c.Duration)
	return nil
}

func attrsFromLabels(ls labels.Labels) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(ls))

	ls.Range(func(l labels.Label) {
		if l.Name != "__name__" {
			attrs = append(attrs, attribute.Key(l.Name).String(l.Value))
		}
	})

	return attrs
}
