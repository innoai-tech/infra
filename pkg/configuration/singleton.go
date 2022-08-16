package configuration

import (
	"context"
	"go/ast"
	"reflect"
)

func SingletonsFromStruct(v any) (singletons Singletons) {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	st := rv.Type()

	for i := 0; i < st.NumField(); i++ {
		ft := st.Field(i)

		if !ast.IsExported(ft.Name) {
			continue
		}

		if ft.Type.Kind() == reflect.Struct {
			singletons = append(singletons, rv.Field(i).Addr().Interface())
		}
	}

	return
}

type Singletons []any

func (list Singletons) Init(ctx context.Context) (context.Context, error) {
	ctx = ContextWithContextInjector(ctx, ComposeContextInjector(list...))
	return ctx, Init(ctx, list...)
}

func (list Singletons) RunOrServe(ctx context.Context) error {
	return RunOrServe(ctx, list...)
}
