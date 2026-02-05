package openmetrics

import (
	"iter"
)

type MetricFamilySet map[string]*MetricFamily

func (s MetricFamilySet) Metrics() iter.Seq[*Metric] {
	return func(yield func(*Metric) bool) {
		for _, mf := range s {
			for _, m := range mf.Metrics {
				if !yield(m) {
					return
				}
			}
		}
	}
}

type MetricFamily struct {
	Type        string
	Name        string
	Description string
	Metrics     []*Metric
}

type Metric struct {
	Name   string
	Labels map[string]string
	Value  string
}

func All(filters ...func(m *Metric) bool) func(m *Metric) bool {
	return func(m *Metric) bool {
		if len(filters) == 0 {
			return false
		}

		for _, filter := range filters {
			if !filter(m) {
				return false
			}
		}

		return true
	}
}

func Named(name string) func(m *Metric) bool {
	return func(m *Metric) bool {
		return m.Name == name
	}
}

func Labeled(key string, value string) func(m *Metric) bool {
	return func(m *Metric) bool {
		if m.Labels == nil {
			return false
		}
		v, ok := m.Labels[key]
		return ok && v == value
	}
}

func ExistsLabel(key string) func(m *Metric) bool {
	return func(m *Metric) bool {
		if m.Labels == nil {
			return false
		}
		_, ok := m.Labels[key]
		return ok
	}
}
