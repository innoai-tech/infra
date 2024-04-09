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
		if canShutdown, ok := c.(configuration.CanShutdown); ok {
			_ = configuration.Shutdown(ctx, canShutdown)
		}
	})

	return configuration.InjectContext(ctx, c.(configuration.ContextInjector))
}

func TestLog(t *testing.T) {
	ctx := Setup(t, &Otel{
		LogLevel: DebugLevel,
	})

	doLog(ctx)
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
