package otel

import (
	"context"
	"golang.org/x/exp/slog"
	"time"

	"github.com/go-courier/logr"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/innoai-tech/infra/internal/otel"
	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/innoai-tech/infra/pkg/configuration"
)

type Otel struct {
	// Log level
	LogLevel LogLevel `flag:",omitempty"`
	// Log async
	LogAsync bool `flag:",omitempty"`
	// Log filter
	LogFilter OutputFilterType `flag:",omitempty"`
	// When set, will collect traces
	TraceCollectorEndpoint string `flag:",omitempty"`

	tp              *sdktrace.TracerProvider
	log             *slog.Logger
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

		opts := []sdktrace.TracerProviderOption{
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		}

		if o.LogAsync {
			logExporter := otel.WithSpanMapExporter(
				otel.OutputFilter(o.LogFilter),
			)(otel.SlogSpanExporter(otel.NewLogger()))

			opts = append(
				opts,
				sdktrace.WithSyncer(logExporter),
			)
		} else {
			o.log = otel.NewLogger()
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
						semconv.ServiceName(info.App.Name),
						semconv.ServiceVersion(info.App.Version),
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
	_ = o.tp.ForceFlush(ctx)
	return o.tp.Shutdown(ctx)
}

func (o *Otel) InjectContext(ctx context.Context) context.Context {
	log := newSpanLogger(o.tp, trace.SpanFromContext(ctx), o.enabledLogLevel, o.log)

	if info := cli.InfoFromContext(ctx); info != nil {
		log = log.WithValues("app", info.App)
	}

	return configuration.InjectContext(
		ctx,
		configuration.InjectContextFunc(otel.ContextWithTracerProvider, trace.TracerProvider(o.tp)),
		configuration.InjectContextFunc(logr.WithLogger, log),
	)
}
