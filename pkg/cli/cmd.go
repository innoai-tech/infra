package cli

import (
	"context"
	"fmt"
	"go/ast"
	"reflect"
	"strings"

	"github.com/innoai-tech/infra/pkg/configuration"
	"github.com/octohelm/gengo/pkg/camelcase"
	"github.com/spf13/pflag"
)

type Command interface {
	Cmd() *C
}

type CanPreRun interface {
	PreRun(ctx context.Context) error
}

func AddTo[T Command](parent Command, c T) T {
	cc := parent.Cmd()
	cc.subcommands = append(cc.subcommands, c)
	return c
}

type C struct {
	i           Info
	cmdPath     []string
	args        args
	flagVars    []*flagVar
	singletons  configuration.Singletons
	subcommands []Command
}

func (c *C) Cmd() *C {
	return c
}

type CanRuntimeDoc interface {
	RuntimeDoc(names ...string) ([]string, bool)
}

func addConfigurator(c *C, fv reflect.Value, flags *pflag.FlagSet, name string, appName string) {
	c.singletons = append(c.singletons, fv.Addr().Interface())
	collectFlagsFromConfigurator(c, flags, fv, name, appName)
}

func collectFlagsFromConfigurator(c *C, flags *pflag.FlagSet, rv reflect.Value, prefix string, appName string) {
	var docer CanRuntimeDoc

	if rv.CanAddr() {
		vv := rv.Addr().Interface()

		if defaulter, ok := vv.(configuration.Defaulter); ok {
			defaulter.SetDefaults()
		}

		if v, ok := vv.(CanRuntimeDoc); ok {
			docer = v
		}
	}

	st := rv.Type()

	for i := 0; i < st.NumField(); i++ {
		ft := st.Field(i)

		if !ast.IsExported(ft.Name) {
			continue
		}

		fv := rv.Field(i)

		if n, ok := ft.Tag.Lookup("arg"); ok {
			argName := ft.Name

			tt := parseTag(n)
			if n := tt.Name; n != "" {
				argName = n
			}

			if prefix != "" {
				argName = prefix + "_" + argName
			}

			a := &arg{
				Name:  argName,
				Value: fv,
			}

			c.args = append(c.args, a)

			continue
		}

		ff := &flagVar{
			Value: fv,
		}

		flagName := ft.Name

		if n, ok := ft.Tag.Lookup("flag"); ok {
			tt := parseTag(n)
			if name := tt.Name; name != "" {
				flagName = name
			}

			ff.Required = !tt.Has("omitempty")

			ff.Expose = tt.Get("expose")
			ff.Secret = tt.Has("secret")
			ff.Volume = tt.Has("volume")

			if alias, ok := ft.Tag.Lookup("alias"); ok {
				ff.Alias = alias
			}
		}

		if prefix != "" {
			flagName = prefix + "_" + flagName
		}

		if ft.Type.Kind() == reflect.Struct && ff.Type() != "string" {
			if ft.Anonymous {
				collectFlagsFromConfigurator(c, flags, fv, prefix, appName)
			} else {
				collectFlagsFromConfigurator(c, flags, fv, flagName, appName)
			}
			continue
		}

		if docer != nil {
			lines, ok := docer.RuntimeDoc(ft.Name)
			if ok {
				ff.Desc = strings.Join(lines, "\n")
			}
		}

		if can, ok := fv.Interface().(interface{ EnumValues() []any }); ok {
			ff.EnumValues = can.EnumValues()
		}

		ff.Name = camelcase.LowerKebabCase(flagName)
		ff.EnvVar = camelcase.UpperSnakeCase(fmt.Sprintf("%s_%s", appName, flagName))

		c.flagVars = append(c.flagVars, ff)
		ff.Apply(flags)
	}
}
