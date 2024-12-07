package configuration

import (
	"context"
	"iter"

	contextx "github.com/octohelm/x/context"
)

type ContextInjector interface {
	InjectContext(ctx context.Context) context.Context
}

func Background(ctx context.Context) context.Context {
	return ContextInjectorFromContext(ctx).InjectContext(context.Background())
}

func InjectContext(ctx context.Context, contextInjectors ...ContextInjector) context.Context {
	for _, contextInjector := range contextInjectors {
		ctx = contextInjector.InjectContext(ctx)
	}
	return ctx
}

func ComposeContextInjector(configurators ...any) ContextInjector {
	contextInjectors := make([]ContextInjector, 0, len(configurators))
	for _, configurator := range configurators {
		if ci, ok := configurator.(ContextInjector); ok {
			contextInjectors = append(contextInjectors, ci)
		}
	}

	return &composeContextInjector{func(yield func(ContextInjector) bool) {
		for _, ci := range contextInjectors {
			if !yield(ci) {
				return
			}
		}
	}}
}

type composeContextInjector struct {
	contextInjectors iter.Seq[ContextInjector]
}

func (c *composeContextInjector) InjectContext(ctx context.Context) context.Context {
	for contextInjector := range c.contextInjectors {
		ctx = contextInjector.InjectContext(ctx)
	}
	return ctx
}

type contextInjectorCtx struct{}

func ContextWithContextInjector(ctx context.Context, ci ContextInjector) context.Context {
	return contextx.WithValue(ctx, contextInjectorCtx{}, ci)
}

func ContextInjectorFromContext(ctx context.Context) ContextInjector {
	if ci, ok := ctx.Value(contextInjectorCtx{}).(ContextInjector); ok {
		return ci
	}
	return contextInjectorDiscord{}
}

type contextInjectorDiscord struct{}

func (contextInjectorDiscord) InjectContext(ctx context.Context) context.Context {
	return ctx
}

func InjectContextFunc[T any](fn func(ctx context.Context, input T) context.Context, input T) ContextInjector {
	return &injectContextFunc[T]{
		input:  input,
		inject: fn,
	}
}

type injectContextFunc[T any] struct {
	input  T
	inject func(ctx context.Context, input T) context.Context
}

func (f *injectContextFunc[T]) InjectContext(ctx context.Context) context.Context {
	return f.inject(ctx, f.input)
}
