package otel

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"golang.org/x/sync/errgroup"

	"github.com/innoai-tech/infra/pkg/otel/internal/otel"
)

// LogValue 将 OpenTelemetry log.Value 转换为普通 Go 值。
func LogValue(v log.Value) any {
	return otel.LogValue(v)
}

// LogProcessor 是对 sdklog.Processor 的公开别名。
type LogProcessor = sdklog.Processor

// LogRecord 是对 sdklog.Record 的公开别名。
type LogRecord = sdklog.Record

// +gengo:injectable:provider
type LogProcessorRegistry interface {
	RegisterLogProcessor(p sdklog.Processor)
}

type dynamicLogProcessor struct {
	m sync.Map
}

var _ LogProcessorRegistry = &dynamicLogProcessor{}

// RegisterLogProcessor 注册一个动态日志处理器。
func (d *dynamicLogProcessor) RegisterLogProcessor(p sdklog.Processor) {
	d.m.Store(p, struct{}{})
}

var _ sdklog.Processor = &dynamicLogProcessor{}

// Enabled 返回当前处理器是否启用。
func (d *dynamicLogProcessor) Enabled(ctx context.Context, param sdklog.EnabledParameters) bool {
	return true
}

// OnEmit 将日志记录分发给已注册处理器。
func (d *dynamicLogProcessor) OnEmit(ctx context.Context, record *sdklog.Record) error {
	for k := range d.m.Range {
		_ = k.(sdklog.Processor).OnEmit(ctx, record)
	}
	return nil
}

// Shutdown 并发关闭所有已注册处理器。
func (d *dynamicLogProcessor) Shutdown(ctx context.Context) error {
	g, c := errgroup.WithContext(ctx)

	for k := range d.m.Range {
		g.Go(func() error {
			return k.(sdklog.Processor).Shutdown(c)
		})
	}

	return g.Wait()
}

// ForceFlush 并发刷新所有已注册处理器。
func (d *dynamicLogProcessor) ForceFlush(ctx context.Context) error {
	g, c := errgroup.WithContext(ctx)

	for k := range d.m.Range {
		g.Go(func() error {
			return k.(sdklog.Processor).ForceFlush(c)
		})
	}

	return g.Wait()
}
