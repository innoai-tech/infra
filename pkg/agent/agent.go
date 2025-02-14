package agent

import (
	"context"
	"github.com/go-courier/logr"
	"github.com/innoai-tech/infra/pkg/configuration"
	"log/slog"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

type action struct {
	name   string
	action func(ctx context.Context) error
}

type Agent struct {
	done    chan struct{}
	closed  atomic.Bool
	serving atomic.Bool
	wg      sync.WaitGroup
	actions []*action
}

func (x *Agent) Init(ctx context.Context) error {
	x.done = make(chan struct{})
	return nil
}

func (x *Agent) Add(name string, fn func(ctx context.Context) error) {
	if x.serving.Load() {
		return
	}

	a := &action{
		name:   name,
		action: fn,
	}

	x.actions = append(x.actions, a)
}

func (x *Agent) Disabled(ctx context.Context) bool {
	return len(x.actions) == 0
}

func (x *Agent) Done() <-chan struct{} {
	return x.done
}

func (x *Agent) Serve(pctx context.Context) error {
	if x.Disabled(pctx) {
		return nil
	}

	// serve once
	if x.serving.Swap(true) {
		return nil
	}

	contextInjector := configuration.ContextInjectorFromContext(pctx)

	eg := &errgroup.Group{}

	for _, a := range x.actions {
		x.wg.Add(1)

		l := logr.FromContext(pctx)

		if a.name != "" {
			l.WithValues(slog.String("agent.name", a.name))
		}

		l.Info("agent serving...")

		eg.Go(func() error {
			defer x.wg.Done()

			c := contextInjector.InjectContext(context.Background())
			c = logr.LoggerInjectContext(c, l)

			ctx, cancel := context.WithCancel(c)
			go func() {
				<-x.done
				cancel()
			}()

			if err := a.action(ctx); err != nil {
				return err
			}

			return nil
		})
	}

	return eg.Wait()
}

func (x *Agent) Shutdown(ctx context.Context) error {
	if x.closed.Swap(true) {
		return nil
	}

	// close to stop all observable
	close(x.done)
	done := make(chan struct{})

	go func() {
		// graceful
		x.wg.Wait()

		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}

	return nil
}
