package openmetrics

import (
	"iter"
)

// MetricFamilySet 是按名称索引的指标族集合。
type MetricFamilySet map[string]*MetricFamily

// Metrics 返回指标族集合中所有指标的迭代器。
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

// MetricFamily 表示一组同类型、同名称但不同标签的指标。
type MetricFamily struct {
	// Type 指标类型，如 counter、gauge、histogram
	Type        string
	// Name 指标族名称
	Name        string
	// Description 指标描述
	Description string
	// Metrics 指标数据点列表
	Metrics     []*Metric
}

// Metric 表示单个 OpenMetrics 指标数据点。
type Metric struct {
	// Name 指标名称
	Name   string
	// Labels 标签键值对
	Labels map[string]string
	// Value 指标值
	Value  string
}

// All 组合多个指标过滤器，返回同时满足所有条件的过滤器。
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

// Named 创建一个按名称匹配指标的过滤器。
func Named(name string) func(m *Metric) bool {
	return func(m *Metric) bool {
		return m.Name == name
	}
}

// Labeled 创建一个按标签键值匹配指标的过滤器。
func Labeled(key string, value string) func(m *Metric) bool {
	return func(m *Metric) bool {
		if m.Labels == nil {
			return false
		}
		v, ok := m.Labels[key]
		return ok && v == value
	}
}

// ExistsLabel 创建一个按标签键存在性匹配指标的过滤器。
func ExistsLabel(key string) func(m *Metric) bool {
	return func(m *Metric) bool {
		if m.Labels == nil {
			return false
		}
		_, ok := m.Labels[key]
		return ok
	}
}
