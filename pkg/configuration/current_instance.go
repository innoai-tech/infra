package configuration

import (
	"context"
)

type currentInstanceCtx struct{}

// CurrentInstanceInjectContext 将当前实例写入上下文。
func CurrentInstanceInjectContext(ctx context.Context, v any) context.Context {
	return context.WithValue(ctx, currentInstanceCtx{}, v)
}

// CurrentInstanceFromContext 从上下文读取当前实例。
func CurrentInstanceFromContext(ctx context.Context) (any, bool) {
	if v := ctx.Value(currentInstanceCtx{}); v != nil {
		return v, true
	}
	return nil, false
}
