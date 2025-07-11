package cli

import (
	"context"
	"fmt"
	"go/ast"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/innoai-tech/infra/pkg/appinfo"
	"github.com/innoai-tech/infra/pkg/cli/internal"
	"github.com/innoai-tech/infra/pkg/configuration"
)

type AppOptionFunc = func(*appinfo.App)

func WithImageNamespace(imageNamespace string) AppOptionFunc {
	return func(a *appinfo.App) {
		a.ImageNamespace = imageNamespace
	}
}

func NewApp(name string, version string, fns ...AppOptionFunc) Command {
	a := &app{
		version: version,

		a: &appinfo.App{
			Name:    name,
			Version: version,
		},

		C: C{
			info: appinfo.Info{
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
	a       *appinfo.App
	root    *cobra.Command
	version string
}

func (a *app) newFrom(cc Command, parent Command) *cobra.Command {
	c := cc.Cmd()
	c.info.App = a.a

	cmd := &cobra.Command{
		Version: a.version,
	}

	a.bindCommand(cmd, c, cc)

	if parent != nil {
		c.cmdPath = append(parent.Cmd().cmdPath, c.info.Name)
	}

	for i := range c.subcommands {
		cmd.AddCommand(a.newFrom(c.subcommands[i], cc))
	}

	dumpK8s := false
	showConfiguration := false

	cmd.Flags().BoolVarP(&showConfiguration, "list-configuration", "c", os.Getenv("ENV") == "DEV", "show configuration")

	if c.info.Component != nil {
		cmd.Flags().BoolVarP(&dumpK8s, "dump-k8s", "", false, "dump k8s component")
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := c.args.Validate(args); err != nil {
			return err
		}

		ctx := cmd.Context()

		if dumpK8s {
			return c.dumpK8sConfiguration(ctx, "./cuepkg/component")
		}

		envVars := internal.EnvVarsFromEnviron(os.Environ())

		for i := range c.flagVars {
			f := c.flagVars[i]

			if err := f.FromEnvVars(envVars); err != nil {
				return err
			}

			if showConfiguration {
				fmt.Println(f.Info())
			}
		}

		singletons := append(
			configuration.Singletons{{
				Configurator: &c.info,
			}},
			c.singletons...,
		)

		ctx, err := singletons.Init(ctx)
		if err != nil {
			return err
		}

		return singletons.RunOrServe(ctx)
	}

	return cmd
}

func (a *app) bindCommand(c *cobra.Command, info *C, v any) {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf("only support a ptr struct value, but got %#v", v))
	}

	rv = rv.Elem()

	if info.info.Name == "" {
		info.info.Name = strings.ToLower(rv.Type().Name())
	}

	a.bindCommandFromStruct(info, rv, c.Flags())

	if len(info.args) > 0 {
		c.Use = fmt.Sprintf("%s [flags] %s", info.info.Name, info.args)
	} else {
		c.Use = info.info.Name
	}

	c.Short = info.info.Desc
}

func (a *app) bindCommandFromStruct(c *C, rv reflect.Value, flags *pflag.FlagSet) {
	st := rv.Type()

	if v, ok := rv.Interface().(CanRuntimeDoc); ok {
		lines, ok := v.RuntimeDoc()
		if ok && len(lines) > 0 {
			c.info.Desc = lines[0]
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
				n.info.Name = name
			}
			if component, ok := ft.Tag.Lookup("component"); ok {
				tag := internal.ParseTag(component)

				n.info.Component = &appinfo.Component{
					Name:    tag.Name,
					Options: tag.Values,
				}
			}

			if envPrefix, ok := ft.Tag.Lookup("envprefix"); ok {
				n.envPrefix = envPrefix
			}

			continue
		}
	}

	c.singletons = configuration.SingletonsFromStruct(rv)
	for _, s := range c.singletons {
		addConfigurator(c, flags, s.Configurator, s.Name, a.info.Name)
	}
}
