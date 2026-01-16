package otel

import (
	"context"
	"time"

	"github.com/innoai-tech/infra/pkg/otel/metric"
	prometheusclient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"golang.org/x/sync/errgroup"

	"github.com/octohelm/x/logr"

	"github.com/innoai-tech/infra/internal/otel"
	"github.com/innoai-tech/infra/pkg/appinfo"
	"github.com/innoai-tech/infra/pkg/configuration"
)

type LogLevel = otel.LogLevel

const (
	ErrorLevel = otel.ErrorLevel
	WarnLevel  = otel.WarnLevel
	InfoLevel  = otel.InfoLevel
	DebugLevel = otel.DebugLevel
)

type LogFormat = otel.LogFormat

const (
	LogFormatText = otel.LogFormatText
	LogFormatJSON = otel.LogFormatJSON
)

// +gengo:injectable
type Otel struct {
	// Log level
	LogLevel LogLevel `flag:",omitempty"`
	// Log format
	LogFormat LogFormat `flag:",omitempty"`
	// When set, will collect traces
	TraceCollectorEndpoint string `flag:",omitempty"`

	MetricCollectorEndpoint      string `flag:",omitempty"`
	MetricCollectIntervalSeconds int    `flag:",omitempty"`

	tracerProvider *sdktrace.TracerProvider
	loggerProvider *sdklog.LoggerProvider
	meterProvider  *sdkmetric.MeterProvider
	promGatherer   prometheusclient.Gatherer

	enabledLevel logr.Level

	dynamicLogProcessor *dynamicLogProcessor

	info *appinfo.Info `inject:",opt"`
}

func (o *Otel) SetDefaults() {
	if o.LogLevel == "" {
		o.LogLevel = InfoLevel
	}

	if o.LogFormat == "" {
		o.LogFormat = LogFormatJSON
	}

	if o.MetricCollectorEndpoint != "" {
		if o.MetricCollectIntervalSeconds == 0 {
			o.MetricCollectIntervalSeconds = 60
		}
	}
}

func (o *Otel) InjectContext(ctx context.Context) context.Context {
	ctx = configuration.InjectContext(
		ctx,
		configuration.InjectContextFunc(otel.TracerProviderContext.Inject, otel.TracerProvider(o.tracerProvider)),
		configuration.InjectContextFunc(otel.LoggerProviderContext.Inject, otel.LoggerProvider(o.loggerProvider)),
	)

	l := otel.NewLogger(ctx, o.enabledLevel)

	return configuration.InjectContext(
		ctx,
		configuration.InjectContextFunc(LogProcessorRegistryInjectContext, LogProcessorRegistry(o.dynamicLogProcessor)),
		configuration.InjectContextFunc(logr.WithLogger, l),
		configuration.InjectContextFunc(otel.GathererContext.Inject, o.promGatherer),
		configuration.InjectContextFunc(otel.MeterProviderContext.Inject, otel.MeterProvider(o.meterProvider)),
	)
}

func (o *Otel) beforeInit(ctx context.Context) error {
	o.dynamicLogProcessor = &dynamicLogProcessor{}

	return nil
}

func (o *Otel) afterInit(ctx context.Context) error {
	enabledLevel, err := logr.ParseLevel(string(o.LogLevel))
	if err != nil {
		return err
	}
	o.enabledLevel = enabledLevel

	pr := prometheusclient.NewRegistry()

	if err := pr.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
		return err
	}
	if err := pr.Register(collectors.NewGoCollector()); err != nil {
		return err
	}

	o.promGatherer = pr

	prometheusReader, err := prometheus.New(
		prometheus.WithRegisterer(pr),
		prometheus.WithoutScopeInfo(),
	)
	if err != nil {
		return err
	}

	tracerOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	}

	logOpts := []sdklog.LoggerProviderOption{
		sdklog.WithProcessor(
			sdklog.NewSimpleProcessor(otel.SlogExporter(o.LogFormat)),
		),
		sdklog.WithProcessor(o.dynamicLogProcessor),
	}

	meterOpts := []sdkmetric.Option{
		sdkmetric.WithReader(prometheusReader),
	}

	if info := o.info; info != nil {
		res := resource.NewSchemaless(
			semconv.ServiceName(info.App.Name),
			semconv.ServiceVersion(info.App.Version),
		)

		tracerOpts = append(tracerOpts, sdktrace.WithResource(res))
		logOpts = append(logOpts, sdklog.WithResource(res))
		meterOpts = append(meterOpts, sdkmetric.WithResource(res))
	}

	if o.TraceCollectorEndpoint != "" {
		client := otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(o.TraceCollectorEndpoint),
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithTimeout(3*time.Second),
		)

		exporter, err := otlptrace.New(ctx, client)
		if err != nil {
			return err
		}

		tracerOpts = append(tracerOpts,
			sdktrace.WithBatcher(
				otel.IgnoreErrSpanExporter(exporter),
			),
		)
	}

	if o.MetricCollectorEndpoint != "" {
		exporter, err := otlpmetricgrpc.New(
			ctx,
			otlpmetricgrpc.WithEndpoint(o.MetricCollectorEndpoint),
			otlpmetricgrpc.WithInsecure(),
			otlpmetricgrpc.WithTimeout(3*time.Second),
		)
		if err != nil {
			return err
		}

		meterOpts = append(meterOpts,
			sdkmetric.WithReader(
				sdkmetric.NewPeriodicReader(
					exporter,
					sdkmetric.WithInterval(time.Duration(o.MetricCollectIntervalSeconds)*time.Second),
				),
			),
		)
	}

	meterOpts = append(meterOpts, metric.GetMetricViewsOption())

	o.loggerProvider = sdklog.NewLoggerProvider(logOpts...)
	o.tracerProvider = sdktrace.NewTracerProvider(tracerOpts...)
	o.meterProvider = sdkmetric.NewMeterProvider(meterOpts...)

	return nil
}

func (o *Otel) Shutdown(c context.Context) error {
	eg, ctx := errgroup.WithContext(c)

	if tp := o.tracerProvider; tp != nil {
		eg.Go(func() error {
			_ = tp.ForceFlush(ctx)

			return tp.Shutdown(ctx)
		})
	}

	if lp := o.loggerProvider; lp != nil {
		eg.Go(func() error {
			_ = lp.ForceFlush(ctx)

			return lp.Shutdown(ctx)
		})
	}

	if mp := o.meterProvider; mp != nil {
		eg.Go(func() error {
			_ = mp.ForceFlush(ctx)

			return mp.Shutdown(ctx)
		})
	}

	return eg.Wait()
}
