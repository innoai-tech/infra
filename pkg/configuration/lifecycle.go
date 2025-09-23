package configuration

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
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
	configuratorServers := make([]Server, 0, len(configurators))
	configuratorCanShutdowns := make([]CanShutdown, 0, len(configurators))

	for _, configurator := range configurators {
		if x, ok := configurator.(Server); ok {
			configuratorServers = append(configuratorServers, x)
		}

		if x, ok := configurator.(CanShutdown); ok {
			configuratorCanShutdowns = append(configuratorCanShutdowns, x)
		}
	}

	ci := ContextInjectorFromContext(ctx)

	if err := run(
		ci.InjectContext(ctx),
		func(yield func(Runner) bool) {
			for _, configurator := range configurators {
				if x, ok := configurator.(Runner); ok {
					if !yield(x) {
						return
					}
				}
			}
		},
	); err != nil {
		return err
	}

	if len(configuratorServers) > 0 {
		chStop := make(chan os.Signal)

		signal.Notify(chStop,
			os.Interrupt, os.Kill,
			syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT,
			syscall.SIGILL, syscall.SIGABRT, syscall.SIGFPE, syscall.SIGSEGV,
		)

		c, cancel := context.WithCancel(ci.InjectContext(ctx))
		g, gc := errgroup.WithContext(c)

		g.Go(func() error {
			defer cancel()
			<-chStop

			return Shutdown(gc, configuratorCanShutdowns...)
		})

		g.Go(func() error {
			return serve(gc, chStop, slices.Values(configuratorServers))
		})

		return g.Wait()
	}

	if len(configuratorCanShutdowns) > 0 {
		// shutdown as cleanup
		return Shutdown(ci.InjectContext(ctx), configuratorCanShutdowns...)
	}

	return nil
}

func run(ctx context.Context, configuratorRunners iter.Seq[Runner]) error {
	for runner := range configuratorRunners {
		l := log.With(
			slog.String("type", fmt.Sprintf("%T", runner)),
			slog.String("lifecycle", "Run"),
		)

		l.Debug("staring")

		if err := runner.Run(ctx); err != nil {
			return err
		}

		l.Debug("done")
	}
	return nil
}

func serve(ctx context.Context, stopCh chan os.Signal, configuratorServers iter.Seq[Server]) error {
	g, c := errgroup.WithContext(ctx)

	for server := range configuratorServers {
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
			defer l.Debug("exit")

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

func Shutdown(c context.Context, configuratorCanShutdowns ...CanShutdown) error {
	timeout := 10 * time.Second

	g := &errgroup.Group{}

	for _, canShutdown := range configuratorCanShutdowns {
		if d, ok := canShutdown.(CanDisabled); ok {
			if d.Disabled(c) {
				continue
			}
		}

		g.Go(func() error {
			ctx, cancel := context.WithTimeout(c, timeout)
			defer cancel()

			l := log.With(
				slog.String("type", fmt.Sprintf("%T", canShutdown)),
				slog.String("lifecycle", "Shutdown"),
			)

			l.With(slog.String("timeout", timeout.String())).Debug("shutting down")
			defer l.Debug("done")

			done := make(chan error)

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

func Init(ctx context.Context, configurators ...any) error {
	ctx = ContextInjectorFromContext(ctx).InjectContext(ctx)

	for _, configurator := range configurators {
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

	if c, ok := configurator.(CanInit); ok {
		return c.Init(CurrentInstanceInjectContext(ctx, configurator))
	}
	return nil
}

type Defaulter interface {
	SetDefaults()
}

type CanInit interface {
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
