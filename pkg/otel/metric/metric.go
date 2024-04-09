package metric

import (
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type Metric struct {
	Name        string
	Unit        string
	Description string
	Views       []View
}

type View struct {
	Instrument sdkmetric.Instrument
	Stream     sdkmetric.Stream
}
