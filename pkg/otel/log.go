package otel

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"time"

	"github.com/innoai-tech/infra/pkg/cli"

	"github.com/go-courier/logr"
	"github.com/innoai-tech/infra/internal/otel"
	"github.com/innoai-tech/infra/pkg/configuration"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type Otel struct {
	// Log level
	LogLevel LogLevel `flag:",omitempty"`
	// Log filter
	LogFilter OutputFilterType `flag:",omitempty"`
	// When set, will collect traces
	TraceCollectorEndpoint string `flag:",omitempty"`

	tp *sdktrace.TracerProvider

	enabledLogLevel logr.Level
}

func (o *Otel) SetDefaults() {
	if o.LogLevel == "" {
		o.LogLevel = InfoLevel
	}
	if o.LogFilter == "" {
		o.LogFilter = OutputFilterAlways
	}
}

func (o *Otel) Init(ctx context.Context) error {
	if o.tp == nil {
		logExporter := otel.WithSpanMapExporter(
			otel.OutputFilter(o.LogFilter),
		)(otel.ZapSpanExporter(otel.NewZapCore()))

		opts := []sdktrace.TracerProviderOption{
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSyncer(logExporter),
		}

		if o.TraceCollectorEndpoint != "" {
			client := otlptracegrpc.NewClient(
				otlptracegrpc.WithEndpoint(o.TraceCollectorEndpoint),
				otlptracegrpc.WithInsecure(),
				otlptracegrpc.WithTimeout(3*time.Second),
			)
			z, err := otlptrace.New(context.Background(), client)
			if err != nil {
				return err
			}

			opts = append(opts, sdktrace.WithBatcher(
				otel.WithSpanMapExporter(otel.OutputFilter(o.LogFilter))(
					otel.WithErrIgnoreExporter()(z),
				),
			))
		}

		if info := cli.InfoFromContext(ctx); info != nil {
			opts = append(
				opts,
				sdktrace.WithResource(
					resource.NewSchemaless(
						attribute.String("service.name", info.App.String()),
					),
				),
			)
		}

		o.enabledLogLevel, _ = logr.ParseLevel(string(o.LogLevel))
		o.tp = sdktrace.NewTracerProvider(opts...)
	}

	return nil
}

func (o *Otel) Shutdown(ctx context.Context) error {
	if o.tp == nil {
		return nil
	}
	return o.tp.Shutdown(ctx)
}

func (o *Otel) InjectContext(ctx context.Context) context.Context {
	log := newSpanLogger(o.tp, nil, o.enabledLogLevel)

	if info := cli.InfoFromContext(ctx); info != nil {
		log = log.WithValues("app", info.App)
	}

	return configuration.InjectContext(
		ctx,
		configuration.InjectContextFunc(otel.ContextWithTracerProvider, trace.TracerProvider(o.tp)),
		configuration.InjectContextFunc(logr.WithLogger, log),
	)
}
