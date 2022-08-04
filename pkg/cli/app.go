package cli

import (
	"context"
	"fmt"
	"go/ast"
	"os"
	"reflect"
	"strings"

	contextx "github.com/octohelm/x/context"
	"github.com/spf13/pflag"

	"github.com/innoai-tech/infra/pkg/configuration"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type App struct {
	Name    string
	Version string
}

type infoCtx struct {
}

func InfoFromContext(ctx context.Context) *C {
	return ctx.Value(infoCtx{}).(*C)
}

func ContextWithInfo(ctx context.Context, c *C) context.Context {
	return contextx.WithValue(ctx, infoCtx{}, c)
}

func NewApp(name string, version string) Command {
	return &app{
		C: C{
			Name: name,
		},
		version: version,
	}
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
	root    *cobra.Command
	version string
}

func (a *app) newFrom(cc Command, parent Command) *cobra.Command {
	info := cc.CmdInfo()
	info.App = App{
		Name:    a.Name,
		Version: a.version,
	}

	c := &cobra.Command{
		Version: a.version,
	}

	for i := range info.subcommands {
		c.AddCommand(a.newFrom(info.subcommands[i], cc))
	}

	a.bindCommand(c, info, cc)

	c.Args = func(cmd *cobra.Command, args []string) error {
		if info.ValidArgs == nil {
			return nil
		}
		return errors.Wrapf(info.ValidArgs.Validate(args), "%s: wrong args, got %v", info.Name, args)
	}

	dumpK8s := false
	showConfiguration := false

	for i := range info.configurators {
		if _, ok := info.configurators[i].(configuration.ConfiguratorServer); ok {
			c.Flags().BoolVarP(&dumpK8s, "dump-k8s", "", false, "dump k8s of command")
			c.Flags().BoolVarP(&dumpK8s, "list-configuration", "c", false, "show configuration")
			break
		}
	}

	c.RunE = func(cmd *cobra.Command, args []string) error {
		if len(info.configurators) == 0 {
			return cmd.Help()
		}
		info.Args = args

		ctx := cmd.Context()

		configuration.SetDefaults(ctx, info.configurators...)

		if dumpK8s {
			return info.dumpK8sConfiguration(ctx, a.Name, fmt.Sprintf("./cuepkg/component"))
		}

		envVars := make(map[string]string)

		for _, kv := range os.Environ() {
			parts := strings.SplitN(kv, "=", 2)
			envVars[strings.ToUpper(parts[0])] = parts[1]
		}

		for i := range info.flagVars {
			f := info.flagVars[i]
			if err := f.FromEnvVars(envVars); err != nil {
				return err
			}

			if showConfiguration {
				fmt.Println(f.Info())
			}
		}

		ctx = ContextWithInfo(ctx, info)

		ctx = configuration.ContextWithContextInjector(ctx, configuration.ComposeContextInjector(info.configurators...))

		if err := configuration.Init(ctx, info.configurators...); err != nil {
			return err
		}

		if err := configuration.ServeOrRun(ctx, info.configurators...); err != nil {
			return err
		}

		return nil
	}

	return c
}

func (a *app) bindCommand(c *cobra.Command, info *C, v any) {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf("only support a ptr struct value, but got %#v", v))
	}

	rv = rv.Elem()

	if info.Name == "" {
		info.Name = strings.ToLower(rv.Type().Name())
	}

	a.bindCommandFromStruct(info, rv, c.Flags())

	if validArgs := info.ValidArgs; validArgs != nil {
		c.Use = fmt.Sprintf("%s [flags] %s", info.Name, strings.Join(*validArgs, " "))
	} else {
		c.Use = info.Name
	}

	c.Short = info.Desc
}

func (a *app) bindCommandFromStruct(cmdInfo *C, rv reflect.Value, flags *pflag.FlagSet) {
	st := rv.Type()

	if v, ok := rv.Interface().(canRuntimeDoc); ok {
		lines, ok := v.RuntimeDoc()
		if ok && len(lines) > 0 {
			cmdInfo.Desc = lines[0]
		}
	}

	for i := 0; i < st.NumField(); i++ {
		ft := st.Field(i)

		if !ast.IsExported(ft.Name) {
			continue
		}

		fv := rv.Field(i)

		if n, ok := fv.Addr().Interface().(*C); ok {
			if v, ok := ft.Tag.Lookup("args"); ok {
				n.ValidArgs = ParseValidArgs(v)
			}
			continue
		}

		name := ft.Name

		if ft.Anonymous {
			name = ""
		}

		if ft.Type.Kind() == reflect.Struct {
			cmdInfo.addConfigurator(fv, flags, name, a.Name)
		}
	}
}
