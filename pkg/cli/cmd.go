package cli

import (
	"context"
	"fmt"
	"go/ast"
	"reflect"
	"strings"

	"github.com/innoai-tech/infra/pkg/appinfo"
	"github.com/innoai-tech/infra/pkg/cli/internal"
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
	i appinfo.Info

	cmdPath     []string
	subcommands []Command

	args       internal.Args
	flagVars   []*internal.FlagVar
	envPrefix  string
	singletons configuration.Singletons
}

func (c *C) Cmd() *C {
	return c
}

type CanRuntimeDoc interface {
	RuntimeDoc(names ...string) ([]string, bool)
}

func addConfigurator(c *C, flags *pflag.FlagSet, target any, name string, appName string) {
	envPrefix := c.envPrefix
	if envPrefix == "" {
		envPrefix = fmt.Sprintf("%s_", appName)
	}

	collectFlagsFromConfigurator(c, flags, target, name, envPrefix, "")
}

func collectFlagsFromConfigurator(c *C, flags *pflag.FlagSet, target any, prefix string, envPrefix string, parentDoc string) {
	rv, ok := target.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(target)
	}

	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

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

			tt := internal.ParseTag(n)
			if n := tt.Name; n != "" {
				argName = n
			}

			if prefix != "" {
				argName = prefix + "_" + argName
			}

			a := &internal.Arg{
				Name:  argName,
				Value: fv,
			}

			c.args = append(c.args, a)

			continue
		}

		ff := &internal.FlagVar{
			Value: fv,
		}

		flagName := ft.Name

		if n, ok := ft.Tag.Lookup("flag"); ok {
			if n == "-" {
				continue
			}

			tt := internal.ParseTag(n)
			if name := tt.Name; name != "" {
				flagName = name
			}

			ff.Required = !(tt.Has("omitempty") || tt.Has("omitzero"))

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

		doc := parentDoc

		if docer != nil {
			if lines, ok := docer.RuntimeDoc(); ok {
				if d := strings.Join(lines, "\n"); d != "" {
					if doc != "" {
						doc += ": \n"
					}
					doc += d
				}
			}

			if lines, ok := docer.RuntimeDoc(ft.Name); ok {
				if d := strings.Join(lines, "\n"); d != "" {
					if doc != "" {
						doc += ": \n"
					}
					doc += d
				}
			}
		}

		if ft.Type.Kind() == reflect.Struct && ff.Type() != "string" {
			if ft.Anonymous {
				collectFlagsFromConfigurator(c, flags, fv, prefix, envPrefix, doc)
			} else {
				collectFlagsFromConfigurator(c, flags, fv, flagName, envPrefix, doc)
			}
			continue
		}

		if can, ok := fv.Interface().(interface{ EnumValues() []any }); ok {
			ff.EnumValues = can.EnumValues()
		}

		ff.Name = camelcase.LowerKebabCase(flagName)
		ff.EnvVar = camelcase.UpperSnakeCase(fmt.Sprintf("%s%s", envPrefix, flagName))
		ff.Desc = doc

		c.flagVars = append(c.flagVars, ff)

		ff.Apply(flags)
	}
}
