package metric

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel/sdk/metric"
)

func newOption(name string, optFuncs ...OptionFunc) *option {
	o := &option{}
	o.Name = name

	for i := range optFuncs {
		optFuncs[i](o)
	}

	return o
}

type option struct {
	Metric
}

func (o *option) metric() Metric {
	return o.Metric
}

type OptionFunc = func(o *option)

func WithUnit(unit string) OptionFunc {
	return func(o *option) {
		o.Unit = unit
	}
}

func WithDescription(description string) OptionFunc {
	return func(o *option) {
		o.Description = description
	}
}

func WithView(view func(m Metric) View) OptionFunc {
	return func(o *option) {
		o.Views = append(o.Views, view(o.Metric))
	}
}

func WithAggregation(aggregation metric.Aggregation) OptionFunc {
	return WithView(func(m Metric) View {
		return View{
			Instrument: metric.Instrument{
				Name:        m.Name,
				Unit:        m.Unit,
				Description: m.Description,
			},
			Stream: metric.Stream{
				Name:        m.Name,
				Aggregation: aggregation,
			},
		}
	})
}

func WithAggregationFunc(typ string, d time.Duration) OptionFunc {
	return WithView(func(m Metric) View {
		return View{
			Instrument: metric.Instrument{
				Kind: metric.InstrumentKindObservableGauge,
				Name: fmt.Sprintf("%s__%s.%0.0fs", m.Name, typ, d.Seconds()),
				Unit: m.Unit,
			},
		}
	})
}
