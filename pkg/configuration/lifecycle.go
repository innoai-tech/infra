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

	stopCh := make(chan os.Signal, 1)

	g.Go(func() error {
		signal.Notify(stopCh,
			os.Interrupt, os.Kill,
			syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT,
			syscall.SIGILL, syscall.SIGABRT, syscall.SIGFPE, syscall.SIGSEGV,
		)
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
		return serve(c, stopCh, configuratorServers...)
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

func serve(ctx context.Context, stopCh chan os.Signal, configuratorServers ...Server) error {
	g, c := errgroup.WithContext(ctx)

	for i := range configuratorServers {
		server := configuratorServers[i]

		if d, ok := server.(CanDisabled); ok {
			if d.Disabled(ctx) {
				continue
			}
		}

		g.Go(func() error {
			err := server.Serve(c)
			go func() {
				stopCh <- syscall.SIGTERM
			}()
			return err
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

func Init(ctx context.Context, configurators ...any) error {
	ctx = ContextInjectorFromContext(ctx).InjectContext(ctx)

	for i := range configurators {
		configurator := configurators[i]

		if c, ok := configurator.(Defaulter); ok {
			c.SetDefaults()
		}

		if err := initConfigurator(ctx, configurator); err != nil {
			return err
		}

		if ci, ok := configurator.(ContextInjector); ok {
			ctx = ci.InjectContext(ctx)
		}
	}

	return nil
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

type CanDisabled interface {
	Disabled(ctx context.Context) bool
}
