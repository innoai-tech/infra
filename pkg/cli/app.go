package cli

import (
	"context"
	"fmt"
	"go/ast"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/pflag"

	"github.com/innoai-tech/infra/pkg/configuration"
	"github.com/spf13/cobra"
)

func NewApp(name string, version string, fns ...AppOptionFunc) Command {
	a := &app{
		a: &App{
			Name:    name,
			Version: version,
		},
		C: C{
			i: Info{
				Name: name,
			},
		},
	}

	for i := range fns {
		fns[i](a.a)
	}

	return a
}

func (a *app) ParseArgs(args []string) {
	a.root = a.newFrom(a, nil)
	a.root.SetArgs(args)
}

func (a *app) ExecuteContext(ctx context.Context) error {
	return a.root.ExecuteContext(ctx)
}

type app struct {
	C
	a       *App
	root    *cobra.Command
	version string
}

func (a *app) newFrom(cc Command, parent Command) *cobra.Command {
	c := cc.Cmd()
	c.i.App = a.a

	cmd := &cobra.Command{
		Version: a.version,
	}

	a.bindCommand(cmd, c, cc)

	if parent != nil {
		c.cmdPath = append(parent.Cmd().cmdPath, c.i.Name)
	}

	for i := range c.subcommands {
		cmd.AddCommand(a.newFrom(c.subcommands[i], cc))
	}

	dumpK8s := false
	showConfiguration := false

	cmd.Flags().BoolVarP(&showConfiguration, "list-configuration", "c", os.Getenv("ENV") == "DEV", "show configuration")

	if c.i.Component != "" {
		cmd.Flags().BoolVarP(&dumpK8s, "dump-k8s", "", false, "dump k8s component")
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(c.configurators) == 0 {
			return cmd.Help()
		}

		if err := c.args.Validate(args); err != nil {
			return err
		}

		ctx := cmd.Context()

		configuration.SetDefaults(ctx, c.configurators...)

		if dumpK8s {
			return c.dumpK8sConfiguration(ctx, "./cuepkg/component")
		}

		envVars := make(map[string]string)

		for _, kv := range os.Environ() {
			parts := strings.SplitN(kv, "=", 2)
			envVars[strings.ToUpper(parts[0])] = parts[1]
		}

		for i := range c.flagVars {
			f := c.flagVars[i]
			if err := f.FromEnvVars(envVars); err != nil {
				return err
			}

			if showConfiguration {
				fmt.Println(f.Info())
			}
		}

		configurators := append(
			[]any{c.i},
			c.configurators...,
		)

		ci := configuration.ComposeContextInjector(configurators...)

		ctx = configuration.ContextWithContextInjector(ctx, ci)

		if err := configuration.Init(ctx, c.configurators...); err != nil {
			return err
		}

		return configuration.RunOrServe(ctx, c.configurators...)
	}

	return cmd
}

func (a *app) bindCommand(c *cobra.Command, info *C, v any) {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf("only support a ptr struct value, but got %#v", v))
	}

	rv = rv.Elem()

	if info.i.Name == "" {
		info.i.Name = strings.ToLower(rv.Type().Name())
	}

	a.bindCommandFromStruct(info, rv, c.Flags())

	if len(info.args) > 0 {
		c.Use = fmt.Sprintf("%s [flags] %s", info.i.Name, info.args)
	} else {
		c.Use = info.i.Name
	}

	c.Short = info.i.Desc
}

func (a *app) bindCommandFromStruct(c *C, rv reflect.Value, flags *pflag.FlagSet) {
	st := rv.Type()

	if v, ok := rv.Interface().(CanRuntimeDoc); ok {
		lines, ok := v.RuntimeDoc()
		if ok && len(lines) > 0 {
			c.i.Desc = lines[0]
		}
	}

	for i := 0; i < st.NumField(); i++ {
		ft := st.Field(i)

		if !ast.IsExported(ft.Name) {
			continue
		}

		fv := rv.Field(i)

		if n, ok := fv.Addr().Interface().(*C); ok {
			if name, ok := ft.Tag.Lookup("name"); ok {
				n.i.Name = name
			}
			if component, ok := ft.Tag.Lookup("component"); ok {
				n.i.Component = component
			}
			continue
		}

		name := ft.Name

		if ft.Anonymous {
			name = ""
		}

		if ft.Type.Kind() == reflect.Struct {
			addConfigurator(c, fv, flags, name, a.i.Name)
		}
	}
}
