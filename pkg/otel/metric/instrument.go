package metric

import (
	"context"

	otelmetric "go.opentelemetry.io/otel/metric"

	"github.com/innoai-tech/infra/internal/otel"
)

type Int64Counter interface {
	Add(ctx context.Context, incr int64, options ...otelmetric.AddOption)
}

type Int64Recorder interface {
	Record(ctx context.Context, incr int64, options ...otelmetric.RecordOption)
}

type Float64Counter interface {
	Add(ctx context.Context, incr float64, options ...otelmetric.AddOption)
}

type Float64Recorder interface {
	Record(ctx context.Context, incr float64, options ...otelmetric.RecordOption)
}

func NewInt64Counter(name string, optFuncs ...OptionFunc) Int64Counter {
	o := newOption(name, optFuncs...)

	return &int64Instrument{
		option: o,
		counter: func(meter otelmetric.Meter) (Int64Counter, error) {
			return meter.Int64Counter(o.Name, otelmetric.WithUnit(o.Unit), otelmetric.WithDescription(o.Description))
		},
	}
}

func NewInt64Histogram(name string, optFuncs ...OptionFunc) Int64Recorder {
	o := newOption(name, optFuncs...)

	return &int64Instrument{
		option: o,
		histogram: func(meter otelmetric.Meter) (Int64Recorder, error) {
			return meter.Int64Histogram(o.Name, otelmetric.WithUnit(o.Unit), otelmetric.WithDescription(o.Description))
		},
	}
}

func NewFloat64UpDownCounter(name string, optFuncs ...OptionFunc) Float64Counter {
	o := newOption(name, optFuncs...)

	return &float64Instrument{
		option: o,
		counter: func(meter otelmetric.Meter) (Float64Counter, error) {
			return meter.Float64UpDownCounter(o.Name, otelmetric.WithUnit(o.Unit), otelmetric.WithDescription(o.Description))
		},
	}
}

func NewFloat64Counter(name string, optFuncs ...OptionFunc) Float64Counter {
	o := newOption(name, optFuncs...)

	return &float64Instrument{
		option: o,
		counter: func(meter otelmetric.Meter) (Float64Counter, error) {
			return meter.Float64Counter(o.Name, otelmetric.WithUnit(o.Unit), otelmetric.WithDescription(o.Description))
		},
	}
}

func NewFloat64Histogram(name string, optFuncs ...OptionFunc) Float64Recorder {
	o := newOption(name, optFuncs...)

	return &float64Instrument{
		option: o,
		histogram: func(meter otelmetric.Meter) (Float64Recorder, error) {
			return meter.Float64Histogram(o.Name, otelmetric.WithUnit(o.Unit), otelmetric.WithDescription(o.Description))
		},
	}
}

type int64Instrument struct {
	*option
	counter   func(meter otelmetric.Meter) (Int64Counter, error)
	histogram func(meter otelmetric.Meter) (Int64Recorder, error)
}

func (i *int64Instrument) Add(ctx context.Context, incr int64, options ...otelmetric.AddOption) {
	if c, err := i.counter(otel.Meter(ctx)); err == nil {
		c.Add(ctx, incr, options...)
	}
}

func (i *int64Instrument) Record(ctx context.Context, incr int64, options ...otelmetric.RecordOption) {
	if c, err := i.histogram(otel.Meter(ctx)); err == nil {
		c.Record(ctx, incr, options...)
	}
}

type float64Instrument struct {
	*option
	counter   func(meter otelmetric.Meter) (Float64Counter, error)
	histogram func(meter otelmetric.Meter) (Float64Recorder, error)
}

func (i *float64Instrument) Add(ctx context.Context, incr float64, options ...otelmetric.AddOption) {
	if c, err := i.counter(otel.Meter(ctx)); err == nil {
		c.Add(ctx, incr, options...)
	}
}

func (i *float64Instrument) Record(ctx context.Context, incr float64, options ...otelmetric.RecordOption) {
	if c, err := i.histogram(otel.Meter(ctx)); err == nil {
		c.Record(ctx, incr, options...)
	}
}
