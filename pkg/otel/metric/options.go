package metric

import (
	"fmt"
	"time"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	syncx "github.com/octohelm/x/sync"
)

var metricViews = syncx.Map[string, []View]{}

// GetMetricViewsOption 收集所有已注册的指标视图选项并返回 OTel 配置选项。
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

	// 记录 views
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

// OptionFunc 用于配置指标选项的函数类型。
type OptionFunc = func(o *option)

// WithUnit 设置指标单位。
func WithUnit(unit string) OptionFunc {
	return func(o *option) {
		o.Unit = unit
	}
}

// WithDescription 设置指标描述。
func WithDescription(description string) OptionFunc {
	return func(o *option) {
		o.Description = description
	}
}

// WithView 注册一个自定义指标视图。
func WithView(view func(m Metric) View) OptionFunc {
	return func(o *option) {
		o.Views = append(o.Views, view(o.Metric))
	}
}

// WithAggregation 为指标注册一个特定的聚合策略。
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

// WithAggregationFunc 为给定类型和周期的指标注册聚合视图。
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
