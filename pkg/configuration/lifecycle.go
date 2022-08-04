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

func RunOrServe(ctx context.Context, configurators ...any) error {
	configuratorRunners := make([]Runner, 0, len(configurators))
	configuratorServers := make([]Server, 0, len(configurators))

	for i := range configurators {
		if x, ok := configurators[i].(Runner); ok {
			configuratorRunners = append(configuratorRunners, x)
		}
		if x, ok := configurators[i].(Server); ok {
			configuratorServers = append(configuratorServers, x)
		}
	}

	ci := ContextInjectorFromContext(ctx)

	cc := ci.InjectContext(ctx)

	g, c := errgroup.WithContext(cc)

	if err := run(cc, configuratorRunners...); err != nil {
		return err
	}

	// Shutdown for cleanup when no server
	if len(configuratorServers) == 0 {
		return Shutdown(cc, configurators...)
	}

	g.Go(func() error {
		stopCh := make(chan os.Signal, 1)

		signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
		<-stopCh

		timeout := 10 * time.Second

		if len(configuratorServers) > 0 {
			logr.FromContext(c).Info("shutdowning server in %s", timeout)
		}

		cc, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		return Shutdown(cc, configurators...)
	})

	g.Go(func() error {
		return serve(c, configuratorServers...)
	})

	return g.Wait()
}

func run(ctx context.Context, configuratorRunners ...Runner) error {
	for i := range configuratorRunners {
		if err := configuratorRunners[i].Run(ctx); err != nil {
			return err
		}
	}
	return nil
}

func serve(ctx context.Context, configuratorServers ...Server) error {
	g, c := errgroup.WithContext(ctx)

	for i := range configuratorServers {
		server := configuratorServers[i]

		g.Go(func() error {
			return server.Serve(c)
		})
	}

	return g.Wait()
}

func Shutdown(ctx context.Context, configuratorServers ...any) error {
	g, c := errgroup.WithContext(ctx)

	for i := range configuratorServers {
		if canShutdown, ok := configuratorServers[i].(CanShutdown); ok {
			g.Go(func() error {
				return canShutdown.Shutdown(c)
			})
		}
	}

	return g.Wait()
}

func SetDefaults(ctx context.Context, configurators ...any) {
	for i := range configurators {
		if c, ok := configurators[i].(Defaulter); ok {
			c.SetDefaults()
		}
	}
}

func Init(ctx context.Context, configurators ...any) error {
	g, c := errgroup.WithContext(ContextInjectorFromContext(ctx).InjectContext(ctx))

	for i := range configurators {
		configurator := configurators[i]

		g.Go(func() error {
			return initConfigurator(c, configurator)
		})
	}

	return g.Wait()
}

func initConfigurator(ctx context.Context, configurator any) (err error) {
	if c, ok := configurator.(ConfiguratorInit); ok {
		return c.Init(ctx)
	}

	return nil
}

type Defaulter interface {
	SetDefaults()
}

type ConfiguratorInit interface {
	Init(ctx context.Context) error
}

type Runner interface {
	Run(ctx context.Context) error
}

type Server interface {
	CanShutdown
	Serve(ctx context.Context) error
}

type CanShutdown interface {
	Shutdown(ctx context.Context) error
}
