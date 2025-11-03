package testingutil

import (
	"context"
	"os"
	"testing"

	"golang.org/x/sync/errgroup"

	testingx "github.com/octohelm/x/testing"

	"github.com/innoai-tech/infra/pkg/configuration"
)

func BuildContext[T any](t testing.TB, initial func(*T)) (context.Context, *T) {
	c := new(T)
	initial(c)
	return NewContext(t, c), c
}

func NewContext[T any](t testing.TB, v *T) context.Context {
	tmp := t.TempDir()
	t.Cleanup(func() {
		_ = os.RemoveAll(tmp)
	})

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	t.Chdir(tmp)
	defer func() {
		_ = os.Chdir(cwd)
	}()

	ctx := t.Context()
	if v != nil {
		singletons := configuration.SingletonsFromStruct(v)
		c, err := singletons.Init(ctx)
		testingx.Expect(t, err, testingx.Be[error](nil))
		ctx = c

		for s := range singletons.Configurators() {
			if r, ok := s.(configuration.Runner); ok {
				err := r.Run(ctx)
				testingx.Expect(t, err, testingx.Be[error](nil))
			}
		}

		go func() {
			g, c := errgroup.WithContext(ctx)

			for s := range singletons.Configurators() {
				if server, ok := s.(configuration.Server); ok {
					g.Go(func() error {
						err := server.Serve(c)
						return err
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
