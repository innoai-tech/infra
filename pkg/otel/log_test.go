package otel

import (
	"context"
	"testing"
	"time"

	"github.com/go-courier/logr"
	"github.com/innoai-tech/infra/pkg/configuration"
	testingx "github.com/octohelm/x/testing"
	"github.com/pkg/errors"
)

func Setup(t testing.TB, c any) context.Context {
	t.Helper()

	ctx := context.Background()
	err := configuration.Init(ctx, c)
	testingx.Expect(t, err, testingx.Be[error](nil))

	t.Cleanup(func() {
		_ = configuration.Shutdown(ctx, c)
	})

	return configuration.InjectContext(ctx, c.(configuration.ContextInjector))
}

func TestLog(t *testing.T) {
	t.Run("FilterAlways", func(t *testing.T) {
		t.Run("async", func(t *testing.T) {
			ctx := Setup(t, &Otel{
				LogFilter: OutputFilterAlways,
				LogLevel:  DebugLevel,
				LogAsync:  true,
			})
			doLog(ctx)
		})

		t.Run("sync", func(t *testing.T) {
			ctx := Setup(t, &Otel{
				LogFilter: OutputFilterAlways,
				LogLevel:  DebugLevel,
			})
			doLog(ctx)
		})
	})

	t.Run("OutputOnNever", func(t *testing.T) {
		ctx := Setup(t, &Otel{
			LogFilter: OutputFilterNever,
			LogLevel:  DebugLevel,
		})

		ctx, log := logr.FromContext(ctx).Start(ctx, "op")
		defer log.End()
		doLog(ctx)
	})

	t.Run("OnFailure", func(t *testing.T) {
		ctx := Setup(t, &Otel{
			LogFilter: OutputFilterOnFailure,
			LogLevel:  DebugLevel,
		})
		doLog(ctx)
	})
}

func doLog(ctx context.Context) {
	ctx, log := logr.Start(ctx, "op")
	defer log.End()

	otherActions(ctx)
	someActionWithSpan(ctx)
}

func someActionWithSpan(ctx context.Context) {
	_, log := logr.Start(ctx, "SomeActionWithSpan")
	defer log.End()

	log.Info("info msg")
	log.Debug("debug msg")
	log.Warn(errors.New("warn msg"))
}

func otherActions(ctx context.Context) {
	log := logr.FromContext(ctx)

	time.Sleep(200 * time.Millisecond)

	log.WithValues("test2", 2).Info("test")
	log.Error(errors.New("other action failed."))
}
