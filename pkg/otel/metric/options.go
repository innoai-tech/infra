package metric

import (
	"fmt"
	"time"

	syncx "github.com/octohelm/x/sync"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

var metricViews = syncx.Map[string, []View]{}

func GetMetricViewsOption() sdkmetric.Option {
	views := make([]sdkmetric.View, 0)

	for _, vv := range metricViews.Range {
		for _, v := range vv {
			views = append(views, sdkmetric.NewView(v.Instrument, v.Stream))
		}
	}

	return sdkmetric.WithView(views...)
}

func newOption(name string, optFuncs ...OptionFunc) *option {
	o := &option{}
	o.Name = name

	for i := range optFuncs {
		optFuncs[i](o)
	}

	// records views
	if len(o.Views) > 0 {
		metricViews.Store(name, o.Views)
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

func WithAggregation(aggregation sdkmetric.Aggregation) OptionFunc {
	return WithView(func(m Metric) View {
		return View{
			Instrument: sdkmetric.Instrument{
				Name:        m.Name,
				Unit:        m.Unit,
				Description: m.Description,
			},
			Stream: sdkmetric.Stream{
				Name:        m.Name,
				Aggregation: aggregation,
			},
		}
	})
}

func WithAggregationFunc(typ string, d time.Duration) OptionFunc {
	return WithView(func(m Metric) View {
		return View{
			Instrument: sdkmetric.Instrument{
				Kind: sdkmetric.InstrumentKindObservableGauge,
				Name: fmt.Sprintf("%s__%s.%0.0fs", m.Name, typ, d.Seconds()),
				Unit: m.Unit,
			},
		}
	})
}
