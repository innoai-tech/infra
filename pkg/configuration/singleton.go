package configuration

import (
	"context"
	"go/ast"
	"iter"
	"reflect"
	"slices"
)

func SingletonsFromStruct(v any) (singletons Singletons) {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	w := &singletonWalker{}
	w.walk(rv)

	return w.singletons
}

type singletonWalker struct {
	singletons Singletons
}

func (w *singletonWalker) walk(rv reflect.Value) {
	st := rv.Type()
	for i := 0; i < st.NumField(); i++ {
		ft := st.Field(i)

		if !ast.IsExported(ft.Name) {
			continue
		}

		if ft.Type.Kind() == reflect.Struct {
			v := rv.Field(i).Addr().Interface()

			switch x := v.(type) {
			case Runner, Server, ContextInjector, CanInit:
				name := ft.Name

				if ft.Anonymous {
					name = ""
				}

				w.singletons = append(w.singletons, Singleton{
					Name:         name,
					Configurator: x,
				})
				continue
			}

			if ft.Anonymous {
				w.walk(rv.Field(i))
			}
		}
	}
}

type Singleton struct {
	Name         string
	Configurator any
}

type Singletons []Singleton

func (list Singletons) Configurators() iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, t := range list {
			if !yield(t.Configurator) {
				return
			}
		}
	}
}

func (list Singletons) Init(ctx context.Context) (context.Context, error) {
	configurators := slices.Collect(list.Configurators())
	ctx = ContextWithContextInjector(ctx, ComposeContextInjector(configurators...))
	return ctx, Init(ctx, configurators...)
}

func (list Singletons) RunOrServe(ctx context.Context) error {
	return RunOrServe(ctx, slices.Collect(list.Configurators())...)
}
