package otel

import (
	"context"
	"github.com/innoai-tech/infra/internal/otel"
	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/innoai-tech/infra/pkg/configuration"
	prometheusclient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

type Metric struct {
	mp     *sdkmetric.MeterProvider
	gather prometheusclient.Gatherer
	views  []sdkmetric.View
}

func (o *Metric) RegisterViews(views ...sdkmetric.View) {
	o.views = views
}

func (o *Metric) Init(ctx context.Context) error {
	if o.mp == nil {
		registry := prometheusclient.NewRegistry()

		if err := registry.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
			return err
		}
		if err := registry.Register(collectors.NewGoCollector()); err != nil {
			return err
		}

		prometheusReader, err := prometheus.New(
			prometheus.WithRegisterer(registry),
			prometheus.WithoutScopeInfo(),
		)
		if err != nil {
			return err
		}

		opts := []sdkmetric.Option{
			sdkmetric.WithReader(prometheusReader),
		}

		if info := cli.InfoFromContext(ctx); info != nil {
			opts = append(
				opts,
				sdkmetric.WithResource(
					resource.NewSchemaless(
						semconv.ServiceName(info.App.Name),
						semconv.ServiceVersion(info.App.Version),
					),
				),
			)
		}

		opts = append(opts, sdkmetric.WithView(
			append(o.views,
				sdkmetric.NewView(
					sdkmetric.Instrument{
						Name: "http.server.duration",
					},
					sdkmetric.Stream{
						Aggregation: aggregation.ExplicitBucketHistogram{
							Boundaries: []float64{0, 0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10},
						},
					},
				),
				sdkmetric.NewView(
					sdkmetric.Instrument{
						Name: "http.client.duration",
					},
					sdkmetric.Stream{
						Aggregation: aggregation.ExplicitBucketHistogram{
							Boundaries: []float64{0, 0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10},
						},
					},
				))...,
		))

		o.mp = sdkmetric.NewMeterProvider(opts...)
		o.gather = registry
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
		configuration.InjectContextFunc(otel.ContextWithMeterProvider, metric.MeterProvider(o.mp)),
	)
}

func Meter(ctx context.Context, name string, opts ...metric.MeterOption) metric.Meter {
	return otel.MeterProviderFromContext(ctx).Meter(name, opts...)
}
