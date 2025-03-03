package configuration

import (
	"context"
	"iter"

	contextx "github.com/octohelm/x/context"
)

func Background(ctx context.Context) context.Context {
	i := ContextInjectorFromContext(ctx)
	return ContextInjectorInjectContext(i.InjectContext(context.Background()), i)
}

func InjectContext(ctx context.Context, contextInjectors ...ContextInjector) context.Context {
	for _, contextInjector := range contextInjectors {
		ctx = contextInjector.InjectContext(ctx)
	}
	return ctx
}

type ContextInjector interface {
	InjectContext(ctx context.Context) context.Context
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
	injectors iter.Seq[ContextInjector]
}

func (c *composeContextInjector) InjectContext(ctx context.Context) context.Context {
	for contextInjector := range c.injectors {
		ctx = contextInjector.InjectContext(ctx)
	}
	return ctx
}

type contextInjectorCtx struct{}

func ContextInjectorInjectContext(ctx context.Context, ci ContextInjector) context.Context {
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
