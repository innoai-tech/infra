package agent

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
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
	kind    string
	done    chan struct{}
	closed  atomic.Bool
	serving atomic.Bool
	wg      sync.WaitGroup

	workers []*worker
}

func (x *Agent) Init(ctx context.Context) error {
	if v, ok := configuration.CurrentInstanceFromContext(ctx); ok {
		if kindGetter, ok := v.(interface{ GetKind() string }); ok {
			x.kind = kindGetter.GetKind()
		} else {
			t := reflect.TypeOf(v)
			for t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			x.kind = t.Name()
		}
	}

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

		if x.kind != "" {
			l = l.WithValues(slog.String("agent.kind", x.kind))
		}

		if w.name != "" {
			l = l.WithValues(slog.String("agent.worker", w.name))
		}

		l.Info("serving")

		x.wg.Add(1)
		eg.Go(func() error {
			defer x.wg.Done()

			c := configuration.ContextInjectorInjectContext(context.Background(), contextInjector)
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
		// pick first to get agent/worker scope
		l := logr.FromContext(ctx)

		defer func() {
			if e := recover(); e != nil {
				l.Error(fmt.Errorf("panic: %#v", e))
			}
		}()

		defer x.wg.Done()

		c := logr.LoggerInjectContext(configuration.Background(ctx), l)
		cc, l2 := l.Start(c, "Go")
		defer l2.End()

		if err := action(cc); err != nil {
			l2.Error(err)
		}
	}()
}
