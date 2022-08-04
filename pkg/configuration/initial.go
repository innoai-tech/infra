package configuration

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-courier/logr"
	"golang.org/x/sync/errgroup"
)

func ServeOrRun(ctx context.Context, configurators ...any) error {
	configuratorRunners := make([]ConfiguratorRunner, 0, len(configurators))
	configuratorServers := make([]ConfiguratorServer, 0, len(configurators))

	for i := range configurators {
		switch x := configurators[i].(type) {
		case ConfiguratorRunner:
			configuratorRunners = append(configuratorRunners, x)
		case ConfiguratorServer:
			configuratorServers = append(configuratorServers, x)
		}
	}

	for i := range configuratorRunners {
		if err := configuratorRunners[i].Run(ctx); err != nil {
			return err
		}
	}

	if len(configuratorServers) == 0 {
		return nil
	}

	go func() {
		if err := serve(ctx, configuratorServers...); err != nil {
			logr.FromContext(ctx).Error(err)
			fmt.Printf("%s\n", err)
		}
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
	<-stopCh

	timeout := 10 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Printf("shutdowning in %s\n", timeout)

	return Shutdown(ctx, configurators...)
}

func serve(ctx context.Context, configuratorServers ...ConfiguratorServer) error {
	g, c := errgroup.WithContext(ctx)

	for i := range configuratorServers {
		configurator := configuratorServers[i]

		g.Go(func() error {
			return configurator.Serve(c)
		})
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

func SetDefaults(ctx context.Context, configurators ...any) {
	for i := range configurators {
		if c, ok := configurators[i].(Defaulter); ok {
			c.SetDefaults()
		}
	}
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
	if c, ok := configurator.(ConfiguratorInit); ok {
		return c.Init(ctx)
	}

	return nil
}

type Defaulter interface {
	SetDefaults()
}

type ConfiguratorRunner interface {
	Run(ctx context.Context) error
}

type ConfiguratorServer interface {
	Serve(ctx context.Context) error
}

type ConfiguratorInit interface {
	Init(ctx context.Context) error
}

type ConfiguratorShutdown interface {
	Shutdown(ctx context.Context) error
}
