package metric

import (
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type Metric struct {
	// Name 指标名称
	Name        string
	// Unit 指标单位
	Unit        string
	// Description 指标描述
	Description string
	// Views 关联的视图列表
	Views       []View
}

type View struct {
	// Instrument 仪器配置
	Instrument sdkmetric.Instrument
	// Stream 流配置
	Stream     sdkmetric.Stream
}
