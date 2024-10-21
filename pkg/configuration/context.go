package configuration

import (
	"context"

	contextx "github.com/octohelm/x/context"
)

func InjectContext(ctx context.Context, contextInjectors ...ContextInjector) context.Context {
	for i := range contextInjectors {
		ctx = contextInjectors[i].InjectContext(ctx)
	}
	return ctx
}

func ComposeContextInjector(configurators ...any) ContextInjector {
	contextInjectors := make([]ContextInjector, 0, len(configurators))
	for i := range configurators {
		if ci, ok := configurators[i].(ContextInjector); ok {
			contextInjectors = append(contextInjectors, ci)
		}
	}

	return &composeContextInjector{contextInjectors}
}

type composeContextInjector struct {
	contextInjectors []ContextInjector
}

func (c *composeContextInjector) InjectContext(ctx context.Context) context.Context {
	return InjectContext(ctx, c.contextInjectors...)
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

type ContextInjector interface {
	InjectContext(ctx context.Context) context.Context
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

func Background(ctx context.Context) context.Context {
	return ContextInjectorFromContext(ctx).InjectContext(context.Background())
}
