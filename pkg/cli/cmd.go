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
	CmdInfo() *C
}

type CanPreRun interface {
	PreRun(ctx context.Context) error
}

func AddTo[T Command](parent Command, c T) T {
	cc := parent.CmdInfo()
	cc.subcommands = append(cc.subcommands, c)
	c.CmdInfo().parent = parent
	return c
}

type C struct {
	Name      string
	Desc      string
	Args      []string
	ValidArgs *ValidArgs
	App       App

	parent      Command
	subcommands []Command

	flagVars      []*flagVar
	configurators []any
}

func (c *C) CmdInfo() *C {
	return c
}

func ParseValidArgs(s string) *ValidArgs {
	if s == "" {
		return nil
	}

	v := make(ValidArgs, 0)

	args := strings.Split(s, " ")

	for i := range args {
		arg := strings.TrimSpace(args[i])
		v = append(v, arg)
	}

	return &v
}

type ValidArgs []string

func (as ValidArgs) HasVariadic() bool {
	for _, a := range as {
		if strings.HasSuffix(a, "...") {
			return true
		}
	}
	return false
}

func (as ValidArgs) Validate(args []string) error {
	if as.HasVariadic() {
		if len(args) < len(as) {
			return fmt.Errorf("requires at least %d arg(s), only received %d", len(as), len(args))
		}
	}
	if len(as) != len(args) {
		return fmt.Errorf("accepts %d arg(s), received %d", len(as), len(args))
	}
	return nil
}

func (c *C) addConfigurator(fv reflect.Value, flags *pflag.FlagSet, name string, appName string) {
	c.configurators = append(c.configurators, fv.Addr().Interface())
	c.collectFlagsFromConfigurator(flags, fv, name, appName)

}

func (c *C) collectFlagsFromConfigurator(flags *pflag.FlagSet, rv reflect.Value, prefix string, appName string) {
	var docer canRuntimeDoc

	if rv.CanAddr() {
		vv := rv.Addr().Interface()

		if defaulter, ok := vv.(configuration.Defaulter); ok {
			defaulter.SetDefaults()
		}

		if v, ok := vv.(canRuntimeDoc); ok {
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

		ff := &flagVar{
			Value: fv,
		}

		flagName := ft.Name

		if n, ok := ft.Tag.Lookup("flag"); ok {
			parts := strings.Split(n, ",")
			if name := parts[0]; name != "" {
				flagName = name
			}
			ff.Required = !strings.Contains(n, ",omitempty")
			ff.Secret = strings.Contains(n, ",secret")
		}

		if prefix != "" {
			flagName = prefix + "_" + flagName
		}

		if ft.Type.Kind() == reflect.Struct && ff.Type() != "string" {
			c.collectFlagsFromConfigurator(flags, fv, flagName, appName)
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
		ff.Apply(flags)
		c.flagVars = append(c.flagVars, ff)
	}
}
