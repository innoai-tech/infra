package metric

import (
	"context"
	"sync"

	otelmetric "go.opentelemetry.io/otel/metric"
)

type Int64Counter interface {
	Add(ctx context.Context, incr int64, options ...otelmetric.AddOption)
}

type Int64Recorder interface {
	Record(ctx context.Context, incr int64, options ...otelmetric.RecordOption)
}

type Int64Observable interface {
	AddObserve(ctx context.Context, callback otelmetric.Int64Callback) func()
}

type Float64Counter interface {
	Add(ctx context.Context, incr float64, options ...otelmetric.AddOption)
}
type Float64Recorder interface {
	Record(ctx context.Context, incr float64, options ...otelmetric.RecordOption)
}

type Float64Observable interface {
	AddObserve(ctx context.Context, callback otelmetric.Float64Callback) func()
}

func NewInt64UpDownCounter(name string, optFuncs ...OptionFunc) Int64Counter {
	o := newOption(name, optFuncs...)

	return register(&int64Instrument{
		option: o,
		createCounter: func(meter otelmetric.Meter) (Int64Counter, error) {
			return meter.Int64UpDownCounter(o.Name, otelmetric.WithUnit(o.Unit), otelmetric.WithDescription(o.Description))
		},
	})
}

func NewInt64ObservableGauge(name string, optFuncs ...OptionFunc) Int64Observable {
	o := newOption(name, optFuncs...)

	return register(&int64Instrument{
		option: o,
		createObservable: func(meter otelmetric.Meter) (Int64Observable, error) {
			io := &int64Observable{}

			_, err := meter.Int64ObservableGauge(
				o.Name,
				otelmetric.WithUnit(o.Unit),
				otelmetric.WithDescription(o.Description),
				otelmetric.WithInt64Callback(io.Callback),
			)

			return io, err
		},
	})
}

func NewInt64Counter(name string, optFuncs ...OptionFunc) Int64Counter {
	o := newOption(name, optFuncs...)

	return register(&int64Instrument{
		option: o,
		createCounter: func(meter otelmetric.Meter) (Int64Counter, error) {
			return meter.Int64Counter(o.Name, otelmetric.WithUnit(o.Unit), otelmetric.WithDescription(o.Description))
		},
	})
}

func NewInt64Histogram(name string, optFuncs ...OptionFunc) Int64Recorder {
	o := newOption(name, optFuncs...)

	return register(&int64Instrument{
		option: o,
		createHistogram: func(meter otelmetric.Meter) (Int64Recorder, error) {
			return meter.Int64Histogram(o.Name, otelmetric.WithUnit(o.Unit), otelmetric.WithDescription(o.Description))
		},
	})
}

func NewFloat64UpDownCounter(name string, optFuncs ...OptionFunc) Float64Counter {
	o := newOption(name, optFuncs...)

	return register(&float64Instrument{
		option: o,
		createCounter: func(meter otelmetric.Meter) (Float64Counter, error) {
			return meter.Float64UpDownCounter(o.Name, otelmetric.WithUnit(o.Unit), otelmetric.WithDescription(o.Description))
		},
	})
}

func NewFloat64ObservableGauge(name string, optFuncs ...OptionFunc) Float64Observable {
	o := newOption(name, optFuncs...)

	return register(&float64Instrument{
		option: o,
		createObservable: func(meter otelmetric.Meter) (Float64Observable, error) {
			fo := &float64Observable{}

			_, err := meter.Float64ObservableGauge(
				o.Name,
				otelmetric.WithUnit(o.Unit),
				otelmetric.WithDescription(o.Description),
				otelmetric.WithFloat64Callback(fo.Callback),
			)

			return fo, err
		},
	})
}

func NewFloat64Counter(name string, optFuncs ...OptionFunc) Float64Counter {
	o := newOption(name, optFuncs...)

	return register(&float64Instrument{
		option: o,
		createCounter: func(meter otelmetric.Meter) (Float64Counter, error) {
			return meter.Float64Counter(o.Name, otelmetric.WithUnit(o.Unit), otelmetric.WithDescription(o.Description))
		},
	})
}

func NewFloat64Histogram(name string, optFuncs ...OptionFunc) Float64Recorder {
	o := newOption(name, optFuncs...)

	return register(&float64Instrument{
		option: o,
		createHistogram: func(meter otelmetric.Meter) (Float64Recorder, error) {
			return meter.Float64Histogram(o.Name, otelmetric.WithUnit(o.Unit), otelmetric.WithDescription(o.Description))
		},
	})
}

type int64Instrument struct {
	*option
	createObservable func(meter otelmetric.Meter) (Int64Observable, error)
	createCounter    func(meter otelmetric.Meter) (Int64Counter, error)
	createHistogram  func(meter otelmetric.Meter) (Int64Recorder, error)
}

func (i *int64Instrument) register(r RegistryAdder) error {
	if i.createObservable != nil {
		o, err := i.createObservable(r.Meter())
		if err != nil {
			return err
		}
		r.AddInt64Observable(i.Name, o)
	}

	if i.createCounter != nil {
		c, err := i.createCounter(r.Meter())
		if err != nil {
			return err
		}
		r.AddInt64Counter(i.Name, c)
	}

	if i.createHistogram != nil {
		h, err := i.createHistogram(r.Meter())
		if err != nil {
			return err
		}
		r.AddInt64Histogram(i.Name, h)
	}
	return nil
}

func (i *int64Instrument) Add(ctx context.Context, incr int64, options ...otelmetric.AddOption) {
	if c, ok := RegistryFromContext(ctx).Int64Counter(i.Name); ok {
		c.Add(ctx, incr, options...)
	}
}

func (i *int64Instrument) Record(ctx context.Context, incr int64, options ...otelmetric.RecordOption) {
	if c, ok := RegistryFromContext(ctx).Int64Histogram(i.Name); ok {
		c.Record(ctx, incr, options...)
	}
}

func (i *int64Instrument) AddObserve(ctx context.Context, callback otelmetric.Int64Callback) func() {
	if c, ok := RegistryFromContext(ctx).Int64Observable(i.Name); ok {
		return c.AddObserve(ctx, callback)
	}
	return func() {}
}

type int64Observable struct {
	lock      sync.Mutex
	callbacks []otelmetric.Int64Callback
}

func (o *int64Observable) Callback(ctx context.Context, observer otelmetric.Int64Observer) error {
	for i := range o.callbacks {
		if c := o.callbacks[i]; c != nil {
			if err := c(ctx, observer); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *int64Observable) AddObserve(ctx context.Context, callback otelmetric.Int64Callback) func() {
	o.lock.Lock()
	defer o.lock.Unlock()

	i := len(o.callbacks)
	o.callbacks = append(o.callbacks, callback)

	return func() {
		o.lock.Lock()
		defer o.lock.Unlock()

		o.callbacks[i] = nil
	}
}

type float64Instrument struct {
	*option
	createObservable func(meter otelmetric.Meter) (Float64Observable, error)
	createCounter    func(meter otelmetric.Meter) (Float64Counter, error)
	createHistogram  func(meter otelmetric.Meter) (Float64Recorder, error)
}

func (i *float64Instrument) register(r RegistryAdder) error {
	if i.createObservable != nil {
		o, err := i.createObservable(r.Meter())
		if err != nil {
			return err
		}
		r.AddFloat64Observable(i.Name, o)
	}

	if i.createCounter != nil {
		c, err := i.createCounter(r.Meter())
		if err != nil {
			return err
		}
		r.AddFloat64Counter(i.Name, c)
	}

	if i.createHistogram != nil {
		h, err := i.createHistogram(r.Meter())
		if err != nil {
			return err
		}
		r.AddFloat64Histogram(i.Name, h)
	}
	return nil
}

func (i *float64Instrument) Add(ctx context.Context, incr float64, options ...otelmetric.AddOption) {
	if c, ok := RegistryFromContext(ctx).Float64Counter(i.Name); ok {
		c.Add(ctx, incr, options...)
	}
}

func (i *float64Instrument) Record(ctx context.Context, incr float64, options ...otelmetric.RecordOption) {
	if c, ok := RegistryFromContext(ctx).Float64Histogram(i.Name); ok {
		c.Record(ctx, incr, options...)
	}
}

func (i *float64Instrument) AddObserve(ctx context.Context, callback otelmetric.Float64Callback) func() {
	if c, ok := RegistryFromContext(ctx).Float64Observable(i.Name); ok {
		return c.AddObserve(ctx, callback)
	}
	return func() {}
}

type float64Observable struct {
	lock      sync.Mutex
	callbacks []otelmetric.Float64Callback
}

func (o *float64Observable) Callback(ctx context.Context, observer otelmetric.Float64Observer) error {
	for i := range o.callbacks {
		if c := o.callbacks[i]; c != nil {
			if err := c(ctx, observer); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *float64Observable) AddObserve(ctx context.Context, callback otelmetric.Float64Callback) func() {
	o.lock.Lock()
	defer o.lock.Unlock()

	i := len(o.callbacks)
	o.callbacks = append(o.callbacks, callback)

	return func() {
		o.lock.Lock()
		defer o.lock.Unlock()

		o.callbacks[i] = nil
	}
}
