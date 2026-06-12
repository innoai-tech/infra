package testingutil

import (
	"context"
	"os"

	"golang.org/x/sync/errgroup"

	. "github.com/octohelm/x/testing/v2"

	"github.com/innoai-tech/infra/pkg/configuration"
)

// BuildContext 创建并初始化测试上下文，返回注入后的 context 和配置对象指针。
func BuildContext[T any](t TB, initial func(c *T)) (context.Context, *T) {
	c := new(T)
	initial(c)
	return NewContext(t, c), c
}

// NewContext 基于已有配置对象创建测试上下文，并自动启动服务和管理清理。
func NewContext[T any](t TB, v *T) context.Context {
	tmp := t.TempDir()
	t.Cleanup(func() {
		_ = os.RemoveAll(tmp)
	})

	cwd := MustValue(t, os.Getwd)

	t.Chdir(tmp)
	defer func() {
		_ = os.Chdir(cwd)
	}()

	ctx := t.Context()
	if v != nil {
		singletons := configuration.SingletonsFromStruct(v)

		ctx = MustValue(t, func() (context.Context, error) {
			return singletons.Init(ctx)
		})

		for s := range singletons.Configurators() {
			if r, ok := s.(configuration.Runner); ok {
				Must(t, func() error {
					return r.Run(ctx)
				})
			}
		}

		// 启动异步服务
		go func() {
			g, c := errgroup.WithContext(ctx)
			for s := range singletons.Configurators() {
				if server, ok := s.(configuration.Server); ok {
					g.Go(func() error {
						return server.Serve(c)
					})
				}
			}
			_ = g.Wait()
		}()

		t.Cleanup(func() {
			c := configuration.ContextInjectorFromContext(ctx).InjectContext(ctx)
			for s := range singletons.Configurators() {
				if canShutdown, ok := s.(configuration.CanShutdown); ok {
					_ = configuration.Shutdown(c, canShutdown)
				}
			}
		})
	}

	return configuration.ContextInjectorFromContext(ctx).InjectContext(ctx)
}
