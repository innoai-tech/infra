package configuration

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var log *slog.Logger

func init() {
	opt := &slog.HandlerOptions{
		Level: slog.LevelError,
	}

	if os.Getenv("INFRA_CLI_DEBUG") == "1" {
		opt.Level = slog.LevelDebug
	}

	log = slog.New(slog.NewTextHandler(os.Stdout, opt))
}

func RunOrServe(ctx context.Context, configurators ...any) error {
	configuratorRunners := make([]Runner, 0, len(configurators))
	configuratorServers := make([]Server, 0, len(configurators))
	configuratorCanShutdowns := make([]CanShutdown, 0, len(configurators))

	for i := range configurators {
		if x, ok := configurators[i].(Runner); ok {
			configuratorRunners = append(configuratorRunners, x)
		}
		if x, ok := configurators[i].(CanShutdown); ok {
			configuratorCanShutdowns = append(configuratorCanShutdowns, x)
		}
		if x, ok := configurators[i].(Server); ok {
			configuratorServers = append(configuratorServers, x)
		}
	}

	ci := ContextInjectorFromContext(ctx)

	cc := ci.InjectContext(ctx)

	if err := run(cc, configuratorRunners...); err != nil {
		return err
	}

	if len(configuratorServers) > 0 {
		stopCh := make(chan os.Signal)
		g, c := errgroup.WithContext(cc)

		g.Go(func() error {
			return serve(c, stopCh, configuratorServers...)
		})

		g.Go(func() error {
			signal.Notify(stopCh,
				os.Interrupt, os.Kill,
				syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT,
				syscall.SIGILL, syscall.SIGABRT, syscall.SIGFPE, syscall.SIGSEGV,
			)

			<-stopCh

			return Shutdown(cc, configuratorCanShutdowns...)
		})

		return g.Wait()
	}

	if len(configuratorCanShutdowns) > 0 {
		// shutdown as cleanup
		return Shutdown(cc, configuratorCanShutdowns...)
	}

	return nil
}

func run(ctx context.Context, configuratorRunners ...Runner) error {
	for i := range configuratorRunners {
		l := log.With(
			slog.String("type", fmt.Sprintf("%T", configuratorRunners[i])),
			slog.String("lifecycle", "Run"),
		)

		l.Debug("staring")

		if err := configuratorRunners[i].Run(ctx); err != nil {
			return err
		}

		l.Debug("done")
	}
	return nil
}

func serve(ctx context.Context, stopCh chan os.Signal, configuratorServers ...Server) error {
	g, c := errgroup.WithContext(ctx)

	for _, server := range configuratorServers {
		if d, ok := server.(CanDisabled); ok {
			if d.Disabled(ctx) {
				continue
			}
		}

		g.Go(func() error {
			l := log.With(
				slog.String("type", fmt.Sprintf("%T", server)),
				slog.String("lifecycle", "Serve"),
			)
			l.Debug("serving")
			err := server.Serve(c)
			go func() {
				stopCh <- syscall.SIGTERM
			}()
			return err
		})

		if r, ok := server.(PostServeRunner); ok {
			g.Go(func() error {
				return r.PostServeRun(ctx)
			})
		}
	}

	return g.Wait()
}

func Shutdown(c context.Context, configuratorServers ...CanShutdown) error {
	timeout := 10 * time.Second

	g := &errgroup.Group{}

	for _, canShutdown := range configuratorServers {
		g.Go(func() error {
			ctx, cancel := context.WithTimeout(c, timeout)
			defer cancel()

			l := log.With(
				slog.String("type", fmt.Sprintf("%T", canShutdown)),
				slog.String("lifecycle", "Shutdown"),
				slog.String("timeout", timeout.String()),
			)

			l.Debug("shutting down")
			defer log.Debug("done")

			done := make(chan error)
			defer close(done)

			go func() {
				done <- canShutdown.Shutdown(ctx)
			}()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case err := <-done:
				return err
			}
		})
	}

	return g.Wait()
}

func TypedInit(ctx context.Context, configurators ...ConfiguratorInit) error {
	ctx = ContextInjectorFromContext(ctx).InjectContext(ctx)

	for i := range configurators {
		configurator := configurators[i]

		if err := initConfigurator(ctx, configurator); err != nil {
			return err
		}

		if ci, ok := configurator.(ContextInjector); ok {
			ctx = ci.InjectContext(ctx)
		}
	}

	return nil
}

func Init(ctx context.Context, configurators ...any) error {
	ctx = ContextInjectorFromContext(ctx).InjectContext(ctx)

	for i := range configurators {
		configurator := configurators[i]

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
	log.With(slog.String("type", fmt.Sprintf("%T", configurator))).Debug("init")

	if c, ok := configurator.(Defaulter); ok {
		c.SetDefaults()
	}

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

type PostServeRunner interface {
	PostServeRun(ctx context.Context) error
}

type CanShutdown interface {
	Shutdown(ctx context.Context) error
}

type CanDisabled interface {
	Disabled(ctx context.Context) bool
}
