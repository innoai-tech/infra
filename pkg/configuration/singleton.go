package configuration

import (
	"context"
	"go/ast"
	"iter"
	"reflect"
	"slices"
)

// SingletonsFromStruct 从 struct 中提取支持生命周期或注入的字段。
func SingletonsFromStruct(v any) (singletons Singletons) {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	for rv.Kind() == reflect.Pointer {
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

// Singleton 表示一个可参与配置生命周期的对象。
type Singleton struct {
	Name         string
	Configurator any
}

// Singletons 表示一组 singleton 配置对象。
type Singletons []Singleton

// Configurators 返回 singleton 对应的配置对象序列。
func (list Singletons) Configurators() iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, t := range list {
			if !yield(t.Configurator) {
				return
			}
		}
	}
}

// Init 初始化 singleton 列表并返回串联后的上下文。
func (list Singletons) Init(ctx context.Context) (context.Context, error) {
	configurators := slices.Collect(list.Configurators())
	ctx = ContextInjectorInjectContext(ctx, ComposeContextInjector(configurators...))
	return ctx, Init(ctx, configurators...)
}

// RunOrServe 执行 singleton 列表的生命周期。
func (list Singletons) RunOrServe(ctx context.Context) error {
	return RunOrServe(ctx, slices.Collect(list.Configurators())...)
}
