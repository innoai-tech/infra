package configuration

import "context"

type currentInstanceCtx struct{}

func CurrentInstanceInjectContext(ctx context.Context, v any) context.Context {
	return context.WithValue(ctx, currentInstanceCtx{}, v)
}

func CurrentInstanceFromContext(ctx context.Context) (any, bool) {
	if v := ctx.Value(currentInstanceCtx{}); v != nil {
		return v, true
	}
	return nil, false
}
