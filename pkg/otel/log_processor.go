package otel

import (
	"context"
	"sync"

	"github.com/innoai-tech/infra/internal/otel"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"golang.org/x/sync/errgroup"
)

func LogValue(v log.Value) any {
	return otel.LogValue(v)
}

type LogProcessor = sdklog.Processor

type LogRecord = sdklog.Record

// +gengo:injectable:provider
type LogProcessorRegistry interface {
	RegisterLogProcessor(p sdklog.Processor)
}

type dynamicLogProcessor struct {
	m sync.Map
}

var _ LogProcessorRegistry = &dynamicLogProcessor{}

func (d *dynamicLogProcessor) RegisterLogProcessor(p sdklog.Processor) {
	d.m.Store(p, struct{}{})
}

var _ sdklog.Processor = &dynamicLogProcessor{}

func (d *dynamicLogProcessor) OnEmit(ctx context.Context, record *sdklog.Record) error {
	for k := range d.m.Range {
		_ = k.(sdklog.Processor).OnEmit(ctx, record)
	}
	return nil
}

func (d *dynamicLogProcessor) Shutdown(ctx context.Context) error {
	g, c := errgroup.WithContext(ctx)

	for k := range d.m.Range {
		g.Go(func() error {
			return k.(sdklog.Processor).Shutdown(c)
		})
	}

	return g.Wait()
}

func (d *dynamicLogProcessor) ForceFlush(ctx context.Context) error {
	g, c := errgroup.WithContext(ctx)

	for k := range d.m.Range {
		g.Go(func() error {
			return k.(sdklog.Processor).ForceFlush(c)
		})
	}

	return g.Wait()
}
