package cli

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"

	encodingx "github.com/octohelm/x/encoding"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

type flagVar struct {
	Name       string
	Alias      string
	Required   bool
	Secret     bool
	EnvVar     string
	Desc       string
	Value      reflect.Value
	EnumValues []any

	changed bool
}

func (f *flagVar) FromEnvVars(vars map[string]string) error {
	if v, ok := vars[f.EnvVar]; ok {
		if err := f.Set(v); err != nil {
			return errors.Wrapf(err, "set value from %s failed", f.EnvVar)
		}
	}
	return nil
}

func (f *flagVar) Apply(flags *pflag.FlagSet) {
	ff := flags.VarPF(f, f.Name, f.Alias, f.Usage())

	if f.Value.Kind() == reflect.Slice {
		if f.Value.Type().Elem().Kind() == reflect.Bool {
			ff.NoOptDefVal = "true"
		}
	}

	if f.Value.Kind() == reflect.Bool {
		ff.NoOptDefVal = "true"
	}
}

func (f *flagVar) String() string {
	if strings.HasSuffix(f.Type(), "Slice") {
		return "[" + f.string() + "]"
	}
	return f.string()
}

func (f *flagVar) string() string {
	b := &strings.Builder{}

	if f.Value.Kind() == reflect.Slice {
		for i := 0; i < f.Value.Len(); i++ {
			if i > 0 {
				b.WriteByte(',')
			}
			d, _ := encodingx.MarshalText(f.Value.Index(i))
			b.Write(d)
		}
	} else {
		d, _ := encodingx.MarshalText(f.Value)
		b.Write(d)
	}

	return b.String()
}

func (f *flagVar) Type() string {
	if _, ok := f.Value.Interface().(encoding.TextMarshaler); ok {
		return "string"
	}
	if f.Value.Kind() == reflect.Slice {
		return f.typ(f.Value.Type().Elem()) + "Slice"
	}
	return f.typ(f.Value.Type())
}

var textMarshaler = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()

func (f *flagVar) typ(t reflect.Type) string {
	if ok := t.Implements(textMarshaler); ok {
		return "string"
	}
	if t.Kind() == reflect.Ptr {
		return t.Elem().Kind().String()
	}
	return t.Kind().String()
}

func (f *flagVar) Set(s string) error {
	if f.Value.Kind() == reflect.Slice {
		list := strings.Split(s, ",")

		if !f.changed {
			values := reflect.MakeSlice(f.Value.Type(), len(list), len(list))
			for i := 0; i < values.Len(); i++ {
				if err := encodingx.UnmarshalText(values.Index(i), []byte(list[i])); err != nil {
					return err
				}
			}
			f.Value.Set(values)
		} else {
			for i := range list {
				elemRv := reflect.New(f.Value.Type().Elem())
				if err := encodingx.UnmarshalText(elemRv, []byte(list[i])); err != nil {
					return err
				}
				f.Value = reflect.Append(f.Value, elemRv.Elem())
			}
		}
		f.changed = true
		return nil
	}

	return encodingx.UnmarshalText(f.Value, []byte(s))
}

func (f *flagVar) Usage() string {
	s := strings.Builder{}

	s.WriteString(f.Desc)

	if len(f.EnumValues) > 0 {
		s.WriteString(" (ALLOW VALUES: ")

		for i := range f.EnumValues {
			if i > 0 {
				s.WriteString(", ")
			}
			s.WriteString(fmt.Sprintf("%v", f.EnumValues[i]))
		}

		s.WriteString(")")
	}

	if len(f.EnvVar) > 0 {
		s.WriteString(" ${")
		s.WriteString(f.EnvVar)
		s.WriteString("}")
	}

	return s.String()
}

func (f *flagVar) Info() string {
	if f.Secret {
		return fmt.Sprintf("%s = %s", f.EnvVar, strings.Repeat("*", len(f.string())))
	}
	return fmt.Sprintf("%s = %s", f.EnvVar, f.string())
}
