package metric

import (
	"context"

	otelmetric "go.opentelemetry.io/otel/metric"
)

type registryContext struct{}

func ContextWithRegistry(ctx context.Context, r Registry) context.Context {
	return context.WithValue(ctx, registryContext{}, r)
}

func RegistryFromContext(ctx context.Context) Registry {
	if r, ok := ctx.Value(registryContext{}).(Registry); ok {
		return r
	}
	return &discord{}
}

type discord struct {
}

func (discord) Int64Counter(name string) (Int64Counter, bool) {
	return nil, false
}

func (discord) Int64Histogram(name string) (Int64Recorder, bool) {
	return nil, false
}

func (discord) Int64Observable(name string) (Int64Observable, bool) {
	return nil, false
}

func (discord) Float64Counter(name string) (Float64Counter, bool) {
	return nil, false
}

func (discord) Float64Histogram(name string) (Float64Recorder, bool) {
	return nil, false
}

func (discord) Float64Observable(name string) (Float64Observable, bool) {
	return nil, false
}

type Registry interface {
	Int64Counter(name string) (Int64Counter, bool)
	Int64Histogram(name string) (Int64Recorder, bool)
	Int64Observable(name string) (Int64Observable, bool)

	Float64Counter(name string) (Float64Counter, bool)
	Float64Histogram(name string) (Float64Recorder, bool)
	Float64Observable(name string) (Float64Observable, bool)
}

type RegistryAdder interface {
	Meter() otelmetric.Meter

	AddInt64Counter(name string, c Int64Counter)
	AddInt64Histogram(name string, h Int64Recorder)
	AddInt64Observable(name string, o Int64Observable)

	AddFloat64Counter(name string, c Float64Counter)
	AddFloat64Histogram(name string, h Float64Recorder)
	AddFloat64Observable(name string, o Float64Observable)
}

func NewRegistry(m otelmetric.Meter, init func(r RegistryAdder) error) (Registry, error) {
	r := &metricRegister{
		m:                m,
		int64Counters:    map[string]Int64Counter{},
		int64Histograms:  map[string]Int64Recorder{},
		int64Observables: map[string]Int64Observable{},

		float64Counters:    map[string]Float64Counter{},
		float64Histograms:  map[string]Float64Recorder{},
		float64Observables: map[string]Float64Observable{},
	}

	if err := init(r); err != nil {
		return nil, err
	}

	return r, nil
}

type metricRegister struct {
	m otelmetric.Meter

	int64Counters    map[string]Int64Counter
	int64Histograms  map[string]Int64Recorder
	int64Observables map[string]Int64Observable

	float64Counters    map[string]Float64Counter
	float64Histograms  map[string]Float64Recorder
	float64Observables map[string]Float64Observable
}

func (r *metricRegister) AddInt64Observable(name string, o Int64Observable) {
	r.int64Observables[name] = o
}

func (r *metricRegister) AddFloat64Observable(name string, o Float64Observable) {
	r.float64Observables[name] = o
}

func (r *metricRegister) AddInt64Counter(name string, c Int64Counter) {
	r.int64Counters[name] = c
}

func (r *metricRegister) AddInt64Histogram(name string, h Int64Recorder) {
	r.int64Histograms[name] = h
}

func (r *metricRegister) AddFloat64Counter(name string, c Float64Counter) {
	r.float64Counters[name] = c
}

func (r *metricRegister) AddFloat64Histogram(name string, h Float64Recorder) {
	r.float64Histograms[name] = h
}

func (r *metricRegister) Meter() otelmetric.Meter {
	return r.m
}

func (r *metricRegister) Int64Counter(name string) (Int64Counter, bool) {
	if c, ok := r.int64Counters[name]; ok {
		return c, true
	}
	return nil, false
}

func (r *metricRegister) Int64Histogram(name string) (Int64Recorder, bool) {
	if h, ok := r.int64Histograms[name]; ok {
		return h, true
	}
	return nil, false
}

func (r *metricRegister) Float64Counter(name string) (Float64Counter, bool) {
	if c, ok := r.float64Counters[name]; ok {
		return c, true
	}
	return nil, false
}

func (r *metricRegister) Float64Histogram(name string) (Float64Recorder, bool) {
	if h, ok := r.float64Histograms[name]; ok {
		return h, true
	}
	return nil, false
}

func (r *metricRegister) Int64Observable(name string) (Int64Observable, bool) {
	if o, ok := r.int64Observables[name]; ok {
		return o, true
	}
	return nil, false
}

func (r *metricRegister) Float64Observable(name string) (Float64Observable, bool) {
	if o, ok := r.float64Observables[name]; ok {
		return o, true
	}
	return nil, false
}
