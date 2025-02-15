package agent

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/go-courier/logr"
	"github.com/innoai-tech/infra/pkg/configuration"
	"golang.org/x/sync/errgroup"
)

type worker struct {
	name string
	run  func(ctx context.Context) error
}

type Agent struct {
	done    chan struct{}
	closed  atomic.Bool
	serving atomic.Bool
	wg      sync.WaitGroup
	workers []*worker
}

func (x *Agent) Init(ctx context.Context) error {
	x.done = make(chan struct{})
	return nil
}

func (x *Agent) Disabled(ctx context.Context) bool {
	return len(x.workers) == 0
}

func (x *Agent) Serve(pctx context.Context) error {
	if x.Disabled(pctx) {
		return nil
	}

	// run once
	if x.serving.Swap(true) {
		return nil
	}

	contextInjector := configuration.ContextInjectorFromContext(pctx)

	eg := &errgroup.Group{}

	for _, w := range x.workers {

		l := logr.FromContext(pctx)

		if w.name != "" {
			l = l.WithValues(slog.String("worker.name", w.name))
		}

		l.Info("agent worker serving...")

		x.wg.Add(1)
		eg.Go(func() error {
			defer x.wg.Done()

			c := contextInjector.InjectContext(context.Background())
			c = configuration.ContextWithContextInjector(c, contextInjector)
			c = logr.LoggerInjectContext(c, l)

			ctx, cancel := context.WithCancel(c)
			go func() {
				<-x.done
				cancel()
			}()

			if err := w.run(ctx); err != nil {
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
		// graceful shutdown
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

func (x *Agent) Done() <-chan struct{} {
	return x.done
}

// Host register worker func to workers
// never added new one once serving
func (x *Agent) Host(name string, run func(ctx context.Context) error) {
	if x.serving.Load() {
		return
	}

	a := &worker{
		name: name,
		run:  run,
	}

	x.workers = append(x.workers, a)
}

// Go exec with go func to support it could graceful shutdown
func (x *Agent) Go(ctx context.Context, action func(ctx context.Context) error) {
	x.wg.Add(1)

	go func() {
		defer x.wg.Done()

		if err := action(ctx); err != nil {
			logr.FromContext(ctx).Error(err)
		}
	}()
}
