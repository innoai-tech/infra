/*
Package otel GENERATED BY gengo:injectable
DON'T EDIT THIS FILE
*/
package otel

import (
	context "context"

	appinfo "github.com/innoai-tech/infra/pkg/appinfo"
)

type contextLogProcessorRegistry struct{}

func LogProcessorRegistryFromContext(ctx context.Context) (LogProcessorRegistry, bool) {
	if v, ok := ctx.Value(contextLogProcessorRegistry{}).(LogProcessorRegistry); ok {
		return v, true
	}
	return nil, false
}

func LogProcessorRegistryInjectContext(ctx context.Context, tpe LogProcessorRegistry) context.Context {
	return context.WithValue(ctx, contextLogProcessorRegistry{}, tpe)
}

func (v *Otel) Init(ctx context.Context) error {
	if value, ok := appinfo.InfoFromContext(ctx); ok {
		v.info = value
	}

	if err := v.beforeInit(ctx); err != nil {
		return err
	}

	if err := v.afterInit(ctx); err != nil {
		return err
	}

	return nil
}
