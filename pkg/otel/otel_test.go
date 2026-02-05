package otel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/octohelm/x/logr"
	testingv2 "github.com/octohelm/x/testing/v2"

	"github.com/innoai-tech/infra/pkg/configuration"
)

func TestLog(t *testing.T) {
	ctx := setup(t, &Otel{
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
	log.Error(errors.New("other action failed"))
}

func setup(t testing.TB, c any) context.Context {
	t.Helper()

	ctx := t.Context()

	testingv2.Must(t, func() error {
		return configuration.Init(ctx, c)
	})

	t.Cleanup(func() {
		if canShutdown, ok := c.(configuration.CanShutdown); ok {
			_ = configuration.Shutdown(ctx, canShutdown)
		}
	})

	return configuration.InjectContext(ctx, c.(configuration.ContextInjector))
}
