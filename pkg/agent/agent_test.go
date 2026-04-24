package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	contextx "github.com/octohelm/x/context"
	. "github.com/octohelm/x/testing/v2"

	"github.com/innoai-tech/infra/pkg/configuration"
)

type ctxKey string

type namedKind struct{}

func (namedKind) GetKind() string { return "custom-kind" }

type fallbackKind struct{}

func TestInitUsesKindGetter(t *testing.T) {
	t.Parallel()

	a := &Agent{}

	Must(t, func() error {
		return a.Init(configuration.CurrentInstanceInjectContext(context.Background(), namedKind{}))
	})

	Then(t, "Init 优先使用实例暴露的 kind",
		Expect(a.kind, Equal("custom-kind")),
		Expect(a.Done() == nil, Equal(false)),
	)
}

func TestInitFallsBackToTypeName(t *testing.T) {
	t.Parallel()

	a := &Agent{}

	Must(t, func() error {
		return a.Init(configuration.CurrentInstanceInjectContext(context.Background(), &fallbackKind{}))
	})

	Then(t, "未实现 GetKind 时回退到类型名",
		Expect(a.kind, Equal("fallbackKind")),
	)
}

func TestDisabledAndHost(t *testing.T) {
	t.Parallel()

	a := &Agent{}

	Must(t, func() error {
		return a.Init(context.Background())
	})

	Then(t, "默认没有 worker 时视为禁用",
		Expect(a.Disabled(context.Background()), Equal(true)),
	)

	a.Host("worker", func(ctx context.Context) error { return nil })

	Then(t, "注册 worker 后可运行",
		Expect(a.Disabled(context.Background()), Equal(false)),
		Expect(len(a.workers), Equal(1)),
	)
}

func TestServeInjectsContextAndShutdownCancels(t *testing.T) {
	t.Parallel()

	a := &Agent{}

	Must(t, func() error {
		return a.Init(context.Background())
	})

	type runState struct {
		value     string
		cancelled bool
	}

	stateCh := make(chan runState, 1)
	started := make(chan struct{}, 1)

	a.Host("worker", func(ctx context.Context) error {
		started <- struct{}{}
		<-ctx.Done()
		stateCh <- runState{
			value:     valueFromContext(ctx, ctxKey("k")),
			cancelled: true,
		}
		return nil
	})

	ctx := configuration.ContextInjectorInjectContext(context.Background(), configuration.InjectContextFunc(
		func(ctx context.Context, input string) context.Context {
			return contextx.WithValue(ctx, ctxKey("k"), input)
		},
		"v",
	))

	serveDone := make(chan error, 1)
	go func() {
		serveDone <- a.Serve(ctx)
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatalf("worker did not start")
	}

	a.Host("late", func(ctx context.Context) error { return nil })

	Must(t, func() error {
		return a.Shutdown(context.Background())
	})

	state := <-stateCh

	Then(t, "Serve 会注入上下文并在关闭时取消 worker",
		Expect(<-serveDone, Equal(error(nil))),
		Expect(state.value, Equal("v")),
		Expect(state.cancelled, Equal(true)),
		Expect(len(a.workers), Equal(1)),
	)
}

func TestServeReturnsWorkerError(t *testing.T) {
	t.Parallel()

	a := &Agent{}

	Must(t, func() error {
		return a.Init(context.Background())
	})

	expected := errors.New("boom")
	a.Host("worker", func(ctx context.Context) error {
		return expected
	})

	Then(t, "worker 错误会向上传递",
		ExpectDo(func() error {
			return a.Serve(context.Background())
		}, ErrorIs(expected)),
	)
}

func TestShutdownIdempotent(t *testing.T) {
	t.Parallel()

	a := &Agent{}

	Must(t, func() error {
		return a.Init(context.Background())
	})

	Must(t, func() error {
		return a.Shutdown(context.Background())
	})

	select {
	case <-a.Done():
	default:
		t.Fatalf("done channel should be closed")
	}

	Then(t, "重复关闭不会报错且关闭信号已发出",
		ExpectDo(func() error {
			return a.Shutdown(context.Background())
		}),
	)
}

func TestGoUsesBackgroundContextInjector(t *testing.T) {
	t.Parallel()

	a := &Agent{}

	done := make(chan string, 1)

	ctx := configuration.ContextInjectorInjectContext(context.Background(), configuration.InjectContextFunc(
		func(ctx context.Context, input string) context.Context {
			return contextx.WithValue(ctx, ctxKey("async"), input)
		},
		"ok",
	))

	a.Go(ctx, func(ctx context.Context) error {
		done <- valueFromContext(ctx, ctxKey("async"))
		return nil
	})

	Then(t, "Go 会在后台上下文里保留注入值",
		Expect(<-done, Equal("ok")),
	)
}

func valueFromContext(ctx context.Context, key ctxKey) string {
	v, _ := ctx.Value(key).(string)
	return v
}
