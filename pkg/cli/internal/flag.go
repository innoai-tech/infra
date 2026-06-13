package internal

import (
	"bytes"
	"encoding"
	"encoding/csv"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/pflag"

	encodingx "github.com/octohelm/x/encoding"
)

// FlagVar 表示一个命令行 flag 绑定到 struct 字段的映射。
type FlagVar struct {
	// Name flag 名称
	Name string
	// Alias flag 别名
	Alias string
	// Required 是否必填
	Required bool
	// EnvVar 环境变量名
	EnvVar string
	// Desc 描述
	Desc string
	// Value 绑定的值反射
	Value reflect.Value
	// EnumValues 允许枚举值
	EnumValues []any

	// Secret 是否为敏感值
	Secret bool
	// Expose 端口暴露标识
	Expose string
	// Volume 是否为卷挂载
	Volume bool

	changed bool
}

// FromEnvVars 从环境变量中设置 flag 的值。
func (f *FlagVar) FromEnvVars(vars EnvVars) error {
	if v, ok := vars.Get(f.EnvVar); ok {
		if err := f.Set(v); err != nil {
			return fmt.Errorf("set value from %s failed: %w", f.EnvVar, err)
		}
	}
	return nil
}

// Apply 将 flag 注册到 pflag.FlagSet。
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

// String 返回 flag 值的字符串表示。
func (f *FlagVar) String() string {
	if strings.HasSuffix(f.Type(), "Slice") {
		return "[" + f.DefaultValue() + "]"
	}

	return f.DefaultValue()
}

// DefaultValue 返回 flag 的默认值字符串。
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

// Type 返回 flag 值的类型名称。
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

	if t.Kind() == reflect.Pointer {
		return t.Elem().Kind().String()
	}

	return t.Kind().String()
}

// Set 从字符串解析并设置 flag 的值。
func (f *FlagVar) Set(s string) error {
	if f.Value.Kind() == reflect.Slice {
		if s == "" && !f.Required {
			return nil
		}

		r := csv.NewReader(bytes.NewBufferString(s))
		list, err := r.Read()
		if err != nil {
			return err
		}

		if !f.changed {
			values := reflect.MakeSlice(f.Value.Type(), len(list), len(list))
			for i := 0; i < values.Len(); i++ {
				if err := f.unmarshalText(values.Index(i), []byte(list[i])); err != nil {
					return err
				}
			}
			f.Value.Set(values)
		} else {
			for i := range list {
				elemRv := reflect.New(f.Value.Type().Elem())
				if err := f.unmarshalText(elemRv, []byte(list[i])); err != nil {
					return err
				}
				f.Value.Set(reflect.Append(f.Value, elemRv.Elem()))
			}
		}

		f.changed = true
		return nil
	}

	return f.unmarshalText(f.Value, []byte(s))
}

func (f *FlagVar) unmarshalText(target any, text []byte) error {
	// 若为可选参数且值为空则跳过解析
	if len(text) == 0 {
		if !f.Required {
			return nil
		}
	}
	return encodingx.UnmarshalText(target, text)
}

// Usage 返回 flag 的使用说明字符串。
func (f *FlagVar) Usage() string {
	s := strings.Builder{}

	s.WriteString(f.Desc)

	if len(f.EnumValues) > 0 {
		s.WriteString(" (ALLOW VALUES: ")

		for i := range f.EnumValues {
			if i > 0 {
				s.WriteString(", ")
			}
			fmt.Fprintf(&s, "%v", f.EnumValues[i])
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

// Info 返回 flag 的环境变量名和当前值。
func (f *FlagVar) Info() string {
	if s, ok := f.Value.Interface().(interface{ SecurityString() string }); ok {
		return fmt.Sprintf("%s = %s", f.EnvVar, s.SecurityString())
	}
	if f.Secret {
		return fmt.Sprintf("%s = %s", f.EnvVar, strings.Repeat("-", len(f.DefaultValue())))
	}
	return fmt.Sprintf("%s = %s", f.EnvVar, f.DefaultValue())
}
