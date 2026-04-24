package configuration

import (
	"context"
	"iter"

	contextx "github.com/octohelm/x/context"
)

// Background 基于当前上下文中的注入器创建一个新的后台上下文。
func Background(ctx context.Context) context.Context {
	i := ContextInjectorFromContext(ctx)
	return ContextInjectorInjectContext(i.InjectContext(context.Background()), i)
}

// InjectContext 依次执行给定的上下文注入器。
func InjectContext(ctx context.Context, contextInjectors ...ContextInjector) context.Context {
	for _, contextInjector := range contextInjectors {
		ctx = contextInjector.InjectContext(ctx)
	}
	return ctx
}

// ContextInjector 表示可向上下文注入值的对象。
type ContextInjector interface {
	InjectContext(ctx context.Context) context.Context
}

// ComposeContextInjector 将多个配置对象组合为一个注入器。
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
		enabled := true
		if canDisabled, ok := contextInjector.(CanDisabled); ok {
			enabled = !canDisabled.Disabled(ctx)
		}
		if enabled {
			ctx = contextInjector.InjectContext(ctx)
		}
	}
	return ctx
}

type contextInjectorCtx struct{}

// ContextInjectorInjectContext 将注入器放入上下文。
func ContextInjectorInjectContext(ctx context.Context, ci ContextInjector) context.Context {
	return contextx.WithValue(ctx, contextInjectorCtx{}, ci)
}

// ContextInjectorFromContext 从上下文中读取注入器。
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

// InjectContextFunc 使用函数包装一个上下文注入器。
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
