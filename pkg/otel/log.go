package otel

import (
	"context"
	"go.uber.org/zap/zapcore"
	"os"
	"time"

	"github.com/go-courier/logr"
	"github.com/innoai-tech/infra/internal/otel"
	"github.com/innoai-tech/infra/pkg/configuration"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Otel struct {
	// Log level
	LogLevel LogLevel `flag:",omitempty"`
	// Log filter
	LogFilter OutputFilterType `flag:",omitempty"`
	// When set, will collect traces
	TraceCollectorEndpoint string `flag:",omitempty"`

	tp              *sdktrace.TracerProvider
	enabledLogLevel logr.Level
}

func (l *Otel) SetDefaults() {
	if l.LogLevel == "" {
		l.LogLevel = InfoLevel
	}
	if l.LogFilter == "" {
		l.LogFilter = OutputFilterAlways
	}
}

func (l *Otel) Init(ctx context.Context) error {
	if l.tp == nil {
		logExporter := otel.WithSpanMapExporter(
			otel.OutputFilter(l.LogFilter),
		)(otel.ZapSpanExporter(newZapCore()))

		opts := []sdktrace.TracerProviderOption{
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSyncer(logExporter),
		}

		if l.TraceCollectorEndpoint != "" {
			client := otlptracegrpc.NewClient(
				otlptracegrpc.WithEndpoint(l.TraceCollectorEndpoint),
				otlptracegrpc.WithTimeout(1*time.Second),
			)
			z, err := otlptrace.New(context.Background(), client)
			if err != nil {
				return err
			}

			opts = append(opts, sdktrace.WithBatcher(
				otel.WithSpanMapExporter(otel.OutputFilter(l.LogFilter))(
					otel.WithErrIgnoreExporter()(z),
				),
			))
		}

		l.enabledLogLevel, _ = logr.ParseLevel(string(l.LogLevel))
		l.tp = sdktrace.NewTracerProvider(opts...)
	}

	return nil
}

func (l *Otel) Shutdown(ctx context.Context) error {
	if l.tp == nil {
		return nil
	}
	return l.tp.Shutdown(ctx)
}

func (l *Otel) InjectContext(ctx context.Context) context.Context {
	return configuration.InjectContext(
		ctx,
		configuration.InjectContextFunc(otel.ContextWithTracerProvider, trace.TracerProvider(l.tp)),
		configuration.InjectContextFunc(logr.WithLogger, newSpanLogger(nil, l.enabledLogLevel)),
	)
}

func newZapCore() zapcore.Core {
	if os.Getenv("GOENV") == "DEV" {
		return zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			os.Stdout,
			zap.DebugLevel,
		)
	}
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		os.Stdout,
		zap.DebugLevel,
	)
}
