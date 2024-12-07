package internal

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"

	encodingx "github.com/octohelm/x/encoding"
	"github.com/spf13/pflag"
)

type FlagVar struct {
	Name       string
	Alias      string
	Required   bool
	EnvVar     string
	Desc       string
	Value      reflect.Value
	EnumValues []any

	Secret bool
	Expose string
	Volume bool

	changed bool
}

func (f *FlagVar) FromEnvVars(vars map[string]string) error {
	if v, ok := vars[f.EnvVar]; ok {
		if err := f.Set(v); err != nil {
			return fmt.Errorf("set value from %s failed: %w", f.EnvVar, err)
		}
	}
	return nil
}

func (f *FlagVar) Apply(flags *pflag.FlagSet) {
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

func (f *FlagVar) String() string {
	if strings.HasSuffix(f.Type(), "Slice") {
		return "[" + f.DefaultValue() + "]"
	}
	return f.DefaultValue()
}

func (f *FlagVar) DefaultValue() string {
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

func (f *FlagVar) Type() string {
	if _, ok := f.Value.Interface().(encoding.TextMarshaler); ok {
		return "string"
	}
	if f.Value.Kind() == reflect.Slice {
		return f.typ(f.Value.Type().Elem()) + "Slice"
	}
	return f.typ(f.Value.Type())
}

var textMarshaler = reflect.TypeFor[encoding.TextMarshaler]()

func (f *FlagVar) typ(t reflect.Type) string {
	if ok := t.Implements(textMarshaler); ok {
		return "string"
	}
	if t.Kind() == reflect.Ptr {
		return t.Elem().Kind().String()
	}
	return t.Kind().String()
}

func (f *FlagVar) Set(s string) error {
	if f.Value.Kind() == reflect.Slice {
		if s == "" && !f.Required {
			return nil
		}

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
				f.Value.Set(reflect.Append(f.Value, elemRv.Elem()))
			}
		}
		f.changed = true
		return nil
	}

	return encodingx.UnmarshalText(f.Value, []byte(s))
}

func (f *FlagVar) Usage() string {
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

func (f *FlagVar) Info() string {
	if s, ok := f.Value.Interface().(interface{ SecurityString() string }); ok {
		return fmt.Sprintf("%s = %s", f.EnvVar, s.SecurityString())
	}
	if f.Secret {
		return fmt.Sprintf("%s = %s", f.EnvVar, strings.Repeat("-", len(f.DefaultValue())))
	}
	return fmt.Sprintf("%s = %s", f.EnvVar, f.DefaultValue())
}
