package cli

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Execute(ctx context.Context, c Command, args []string) error {
	if inputs, ok := c.(interface{ Init(args []string) }); ok {
		inputs.Init(args)
	}
	if e, ok := c.(interface {
		ExecuteContext(ctx context.Context) error
	}); ok {
		return e.ExecuteContext(ctx)
	}
	return nil
}

func NewApp(name string, version string, flags ...any) Command {
	return &app{
		Name:    Name{Name: name},
		version: version,
		flags:   flags,
	}
}

type app struct {
	Name

	version string
	c       *cobra.Command
	flags   []any
}

func (a *app) ExecuteContext(ctx context.Context) error {
	return a.c.ExecuteContext(ctx)
}

func (a *app) Init(args []string) {
	a.c = a.newCmdFrom(a, nil)
	a.c.SetArgs(args)
	// bind global flags
	for i := range a.flags {
		a.bindCommand(a.c, a.flags[i], a.Naming())
	}
}

func (a *app) PreRun(ctx context.Context) context.Context {
	for i := range a.flags {
		if preRun, ok := a.flags[i].(CanPreRun); ok {
			ctx = preRun.PreRun(ctx)
		}
	}
	return ctx
}

func (a *app) newCmdFrom(cc Command, parent Command) *cobra.Command {
	n := cc.Naming()
	n.parent = parent

	c := &cobra.Command{
		Version: a.version,
	}

	a.bindCommand(c, cc, n)

	c.Args = func(cmd *cobra.Command, args []string) error {
		if n.ValidArgs == nil {
			return nil
		}
		return errors.Wrapf(n.ValidArgs.Validate(args), "%s: wrong args", n.Name)
	}

	c.RunE = func(cmd *cobra.Command, args []string) error {
		n.Args = args

		ctx := cmd.Context()
		// run parent PreRun if exists
		parents := make([]Command, 0)
		for p := parent; p != nil; p = p.Naming().parent {
			parents = append(parents, p)
		}
		for i := range parents {
			if canPreRun, ok := parents[len(parents)-1-i].(CanPreRun); ok {
				ctx = canPreRun.PreRun(ctx)
			}
		}

		if preRun, ok := cc.(CanPreRun); ok {
			ctx = preRun.PreRun(ctx)
		}

		return cc.Run(ctx)
	}

	for i := range n.subcommands {
		c.AddCommand(a.newCmdFrom(n.subcommands[i], cc))
	}

	return c
}

func (a *app) bindCommand(c *cobra.Command, v any, n *Name) {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf("only support a ptr struct value, but got %#v", v))
	}

	rv = rv.Elem()

	if n.Name == "" {
		n.Name = strings.ToLower(rv.Type().Name())
	}

	a.bindFromReflectValue(c, rv)

	if validArgs := n.ValidArgs; validArgs != nil {
		c.Use = fmt.Sprintf("%s [flags] %s", n.Name, strings.Join(*validArgs, " "))
	} else {
		c.Use = n.Name
	}

	c.Short = n.Desc
}

func (a *app) bindFromReflectValue(c *cobra.Command, rv reflect.Value) {
	t := rv.Type()

	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		fv := rv.Field(i)

		if ft.Anonymous && ft.Type.Kind() == reflect.Struct {
			if n, ok := fv.Addr().Interface().(*Name); ok {
				n.Desc = ft.Tag.Get("desc")
				if v, ok := ft.Tag.Lookup("args"); ok {
					n.ValidArgs = ParseValidArgs(v)
				}
				continue
			}
			a.bindFromReflectValue(c, fv)
			continue
		}

		if n, ok := ft.Tag.Lookup("flag"); ok {
			parts := strings.SplitN(n, ",", 2)

			name, alias := parts[0], strings.Join(parts[1:], "")

			persistent := false

			if len(name) > 0 && name[0] == '!' {
				persistent = true
				name = name[1:]
			}

			var envVars []string
			if tagEnv, ok := ft.Tag.Lookup("env"); ok {
				envVars = strings.Split(tagEnv, ",")
			}

			defaultText, defaultExists := ft.Tag.Lookup("default")

			ff := &flagVar{
				Name:        name,
				Alias:       alias,
				EnvVars:     envVars,
				Default:     defaultText,
				Required:    !defaultExists,
				Desc:        ft.Tag.Get("desc"),
				Destination: fv.Addr().Interface(),
			}

			if persistent {
				if err := ff.Apply(c.PersistentFlags()); err != nil {
					panic(err)
				}
			} else {
				if err := ff.Apply(c.Flags()); err != nil {
					panic(err)
				}
			}

		}
	}
}

type flagVar struct {
	Name        string
	Desc        string
	Default     string
	Required    bool
	Alias       string
	EnvVars     []string
	Destination any
}

func (f *flagVar) DefaultValue() string {
	v := f.Default
	for i := range f.EnvVars {
		if found, ok := os.LookupEnv(f.EnvVars[i]); ok {
			v = found
			break
		}
	}
	return v
}

func (f *flagVar) Usage() string {
	if len(f.EnvVars) > 0 {
		s := strings.Builder{}
		s.WriteString(f.Desc)
		s.WriteString(" [")

		for i, envVar := range f.EnvVars {
			if i > 0 {
				s.WriteString(",")
			}
			s.WriteString("$")
			s.WriteString(envVar)
		}

		s.WriteString("]")
		return s.String()
	}
	return f.Desc
}

func (f *flagVar) Apply(flags *pflag.FlagSet) error {
	switch d := f.Destination.(type) {
	case *[]bool:
		var v []bool
		if sv := f.DefaultValue(); sv != "" {
			v = make([]bool, len(strings.Split(sv, ",")))
		}
		flags.BoolSliceVarP(d, f.Name, f.Alias, v, f.Usage())
	case *[]string:
		var v []string
		if sv := f.DefaultValue(); sv != "" {
			v = strings.Split(sv, ",")
		}
		flags.StringSliceVarP(d, f.Name, f.Alias, v, f.Usage())
	case *string:
		v := f.DefaultValue()
		flags.StringVarP(d, f.Name, f.Alias, v, f.Usage())
	case *int:
		var v int
		if sv := f.DefaultValue(); sv != "" {
			v, _ = strconv.Atoi(sv)
		}
		flags.IntVarP(d, f.Name, f.Alias, v, f.Usage())
	case *bool:
		var v bool
		if sv := f.DefaultValue(); sv != "" {
			v, _ = strconv.ParseBool(sv)
		}
		flags.BoolVarP(d, f.Name, f.Alias, v, f.Usage())
	default:
		return fmt.Errorf("unsupported flags type %T", d)
	}
	return nil
}
