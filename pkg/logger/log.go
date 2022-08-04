package logger

import (
	"context"
	"os"
	"time"

	"github.com/go-courier/logr"
	"github.com/innoai-tech/infra/internal/otel"
	"github.com/innoai-tech/infra/pkg/configuration"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type FilterType = otel.OutputFilterType

type Log struct {
	Level        logr.Level `env:""`
	Filter       FilterType `env:""`
	OtlpEndpoint string     `env:""`

	tp *sdktrace.TracerProvider
}

func (l *Log) SetDefaults() {
	if l.Level == logr.PanicLevel {
		l.Level = logr.DebugLevel
	}
	if l.Filter == "" {
		l.Filter = otel.OutputFilterAlways
	}
}

func (l *Log) Init(ctx context.Context) error {
	if l.tp == nil {
		zaprLogger, err := newZapLog()
		if err != nil {
			return err
		}

		logExporter := otel.WithSpanMapExporter(
			otel.OutputFilter(l.Filter),
		)(otel.ZapSpanExporter(zaprLogger))

		opts := []sdktrace.TracerProviderOption{
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSyncer(logExporter),
		}

		if l.OtlpEndpoint != "" {
			client := otlptracehttp.NewClient(
				otlptracehttp.WithEndpoint(l.OtlpEndpoint),
				otlptracehttp.WithTimeout(1*time.Second),
			)
			z, err := otlptrace.New(context.Background(), client)
			if err != nil {
				return err
			}

			opts = append(opts, sdktrace.WithBatcher(
				otel.WithSpanMapExporter(otel.OutputFilter(l.Filter))(
					otel.WithErrIgnoreExporter()(z),
				),
			))
		}

		l.tp = sdktrace.NewTracerProvider(opts...)
	}

	return nil
}

func (l *Log) Shutdown(ctx context.Context) error {
	if l.tp == nil {
		return nil
	}
	return l.tp.Shutdown(ctx)
}

func (l *Log) InjectContext(ctx context.Context) context.Context {
	return configuration.InjectContext(
		ctx,
		configuration.InjectContextFunc(otel.ContextWithTracerProvider, trace.TracerProvider(l.tp)),
		configuration.InjectContextFunc(logr.WithLogger, newSpanLogger(nil, l.Level)),
	)
}

func newZapLog() (*zap.Logger, error) {
	c := zap.NewProductionConfig()
	if os.Getenv("GOENV") == "DEV" {
		c = zap.NewDevelopmentConfig()
	}
	c.DisableCaller = true
	c.DisableStacktrace = true
	return c.Build()
}
