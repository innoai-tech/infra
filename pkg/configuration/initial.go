package configuration

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-courier/logr"
	"golang.org/x/sync/errgroup"
)

func Serve(ctx context.Context, configurators ...any) error {
	go func() {
		if err := serve(ctx, configurators...); err != nil {
			logr.FromContext(ctx).Error(err)
		}
	}()

	stopCh := make(chan os.Signal, 1)
	signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	<-stopCh

	timeout := 10 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logr.FromContext(ctx).Info("shutdowning in %s", timeout)

	return Shutdown(ctx, configurators...)
}

func serve(ctx context.Context, configurators ...any) error {
	ci := ComposeContextInjector(configurators...)

	g, c := errgroup.WithContext(ctx)

	c = ContextWithContextInjector(c, ci)

	for i := range configurators {
		configurator := configurators[i]

		if server, ok := configurator.(ConfiguratorServer); ok {
			g.Go(func() error {
				return server.Serve(c)
			})
		}
	}

	return g.Wait()
}

func Shutdown(ctx context.Context, configurators ...any) error {
	g, c := errgroup.WithContext(ctx)

	for i := range configurators {
		if s, ok := configurators[i].(ConfiguratorShutdown); ok {
			g.Go(func() error {
				return s.Shutdown(c)
			})
		}
	}

	return g.Wait()
}

func Init(ctx context.Context, configurators ...any) error {
	g, c := errgroup.WithContext(ctx)

	for i := range configurators {
		configurator := configurators[i]

		g.Go(func() error {
			return initConfigurator(c, configurator)
		})
	}

	return g.Wait()
}

func initConfigurator(ctx context.Context, configurator any) (err error) {
	if c, ok := configurator.(Defaulter); ok {
		c.SetDefaults()
	}

	if c, ok := configurator.(Configurator); ok {
		return c.Init(ctx)
	}

	return nil
}

type Defaulter interface {
	SetDefaults()
}

type ConfiguratorServer interface {
	Serve(ctx context.Context) error
}

type Configurator interface {
	Init(ctx context.Context) error
}

type ConfiguratorShutdown interface {
	Shutdown(ctx context.Context) error
}
