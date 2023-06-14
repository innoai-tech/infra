package metric

import (
	"sync"

	"github.com/pkg/errors"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

type Factory interface {
	register(r RegistryAdder) error
	metric() Metric
}

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

var factories = sync.Map{}

func register[T Factory](f T) T {
	m := f.metric()
	if _, ok := factories.Load(m.Name); ok {
		panic(errors.Errorf("metric %s already register", m.Name))
	}
	factories.Store(m.Name, f)
	return f
}

func RegisteredViews() (views []View) {
	factories.Range(func(key, value any) bool {
		f := value.(Factory)
		views = append(views, f.metric().Views...)
		return true
	})
	return
}

func AddToRegistry(r RegistryAdder) (err error) {
	factories.Range(func(key, value any) bool {
		f := value.(Factory)
		if e := f.register(r); e != nil {
			err = e
			return false
		}
		return true
	})
	return
}
