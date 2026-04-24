package agent

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"github.com/octohelm/x/logr"

	"github.com/innoai-tech/infra/pkg/configuration"
)

var agentNotHandlePanic = os.Getenv("AGENT_NOT_HANDLE_PANIC") == "1"

type worker struct {
	name string
	run  func(ctx context.Context) error
}

// Agent 管理一组可优雅退出的后台 worker。
//
// 它通常作为 configuration.Server 使用：先在 Init 中确定运行身份，
// 再通过 Host 注册 worker，随后由 Serve / Shutdown 驱动完整生命周期。
type Agent struct {
	kind    string
	done    chan struct{}
	closed  atomic.Bool
	serving atomic.Bool
	wg      sync.WaitGroup

	workers []*worker
}

// Init 根据当前实例推导 agent kind，并初始化关闭信号。
func (x *Agent) Init(ctx context.Context) error {
	if v, ok := configuration.CurrentInstanceFromContext(ctx); ok {
		if kindGetter, ok := v.(interface{ GetKind() string }); ok {
			x.kind = kindGetter.GetKind()
		} else {
			t := reflect.TypeOf(v)
			for t.Kind() == reflect.Pointer {
				t = t.Elem()
			}
			x.kind = t.Name()
		}
	}

	x.done = make(chan struct{})
	return nil
}

// Disabled 返回当前 agent 是否没有可运行的 worker。
func (x *Agent) Disabled(ctx context.Context) bool {
	return len(x.workers) == 0
}

// Serve 启动所有已注册 worker，并在 Shutdown 时广播取消信号。
func (x *Agent) Serve(pctx context.Context) error {
	if x.Disabled(pctx) {
		return nil
	}

	// run once
	if x.serving.Swap(true) {
		return nil
	}

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

			c := configuration.Background(pctx)
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

// Shutdown 停止所有 worker 并等待其退出。
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

// Done 返回 agent 的关闭信号。
func (x *Agent) Done() <-chan struct{} {
	return x.done
}

// Host 注册一个 worker；开始 Serve 后新增注册会被忽略。
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

// Go 在独立 goroutine 中执行动作，并复用日志与上下文注入能力。
func (x *Agent) Go(ctx context.Context, action func(ctx context.Context) error) {
	x.wg.Add(1)

	go func() {
		// pick first to get agent/worker scope
		l := logr.FromContext(ctx)

		if !agentNotHandlePanic {
			defer func() {
				if e := recover(); e != nil {
					switch x := e.(type) {
					case error:
						l.Error(x)
					default:
						l.Error(fmt.Errorf("panic: %#v", e))
					}
				}
			}()
		}

		defer x.wg.Done()

		c := logr.LoggerInjectContext(configuration.Background(ctx), l)
		cc, l2 := l.Start(c, "Go")
		defer l2.End()

		if err := action(cc); err != nil {
			l2.Error(err)
		}
	}()
}
