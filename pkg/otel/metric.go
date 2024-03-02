package otel

import (
	"context"
	"time"

	"github.com/innoai-tech/infra/internal/otel"
	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/innoai-tech/infra/pkg/configuration"
	"github.com/innoai-tech/infra/pkg/otel/metric"
	"github.com/innoai-tech/infra/pkg/otel/metric/aggregation"
	"github.com/octohelm/x/ptr"
	prometheusclient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

type Metric struct {
	EnableSimpleAggregation *bool `flag:",omitempty"`

	CollectorEndpoint      string `flag:",omitempty"`
	CollectIntervalSeconds int    `flag:",omitempty"`

	gather   prometheusclient.Gatherer
	registry metric.Registry
	mp       *sdkmetric.MeterProvider
}

func (o *Metric) SetDefaults() {
	if o.EnableSimpleAggregation == nil {
		o.EnableSimpleAggregation = ptr.Ptr(false)
	}

	if o.CollectorEndpoint != "" {
		if o.CollectIntervalSeconds == 0 {
			o.CollectIntervalSeconds = 60
		}
	}
}

func (o *Metric) Init(ctx context.Context) error {
	if o.registry == nil {
		pr := prometheusclient.NewRegistry()

		if err := pr.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
			return err
		}
		if err := pr.Register(collectors.NewGoCollector()); err != nil {
			return err
		}

		o.gather = pr

		prometheusReader, err := prometheus.New(
			prometheus.WithRegisterer(pr),
			prometheus.WithoutScopeInfo(),
		)
		if err != nil {
			return err
		}

		opts := []sdkmetric.Option{
			sdkmetric.WithReader(prometheusReader),
		}

		appName := ""

		if info := cli.InfoFromContext(ctx); info != nil {
			appName = info.Name

			opts = append(
				opts,
				sdkmetric.WithResource(
					sdkresource.NewSchemaless(
						semconv.ServiceName(info.App.Name),
						semconv.ServiceVersion(info.App.Version),
					),
				),
			)
		}

		if *o.EnableSimpleAggregation {
			aggrReader, err := aggregation.NewReader(metric.RegisteredViews(), func() otelmetric.Meter {
				return o.mp.Meter(appName)
			})
			if err != nil {
				return err
			}

			opts = append(
				opts,
				sdkmetric.WithReader(aggrReader),
			)
		}

		if o.CollectorEndpoint != "" {
			exporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(o.CollectorEndpoint))
			if err != nil {
				return err
			}
			opts = append(opts, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(time.Duration(o.CollectIntervalSeconds)*time.Second))))
		}

		for _, v := range metric.RegisteredViews() {
			opts = append(opts, sdkmetric.WithView(sdkmetric.NewView(v.Instrument, v.Stream)))
		}

		o.mp = sdkmetric.NewMeterProvider(opts...)
		m := o.mp.Meter(appName)

		r, err := metric.NewRegistry(m, metric.AddToRegistry)
		if err != nil {
			return err
		}
		o.registry = r

	}
	return nil
}

func (o *Metric) Shutdown(ctx context.Context) error {
	if o.mp == nil {
		return nil
	}
	_ = o.mp.ForceFlush(ctx)
	return o.mp.Shutdown(ctx)
}

func (o *Metric) InjectContext(ctx context.Context) context.Context {
	return configuration.InjectContext(ctx,
		configuration.InjectContextFunc(otel.ContextWithGatherer, o.gather),
		configuration.InjectContextFunc(metric.ContextWithRegistry, o.registry),
	)
}
