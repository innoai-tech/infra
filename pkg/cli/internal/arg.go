package internal

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/octohelm/gengo/pkg/camelcase"
	encodingx "github.com/octohelm/x/encoding"
)

// Arg 表示一个命令行位置参数。
type Arg struct {
	// Name 参数名称
	Name  string
	// Value 参数值反射
	Value reflect.Value
}

// HasVariadic 返回该位置参数是否为变长参数。
func (as Arg) HasVariadic() bool {
	return as.Value.Kind() == reflect.Slice
}

// Args 表示一组命令行位置参数。
type Args []*Arg

// String 返回参数的显示名称。
func (as Args) String() string {
	s := strings.Builder{}

	for i := range as {
		if i > 0 {
			s.WriteString(" ")
		}

		s.WriteString(camelcase.UpperSnakeCase(as[i].Name))

		if as[i].HasVariadic() {
			s.WriteString("...")
		}
	}

	return s.String()
}

// Validate 校验并设置位置参数的值。
func (as Args) Validate(args []string) error {
	for i, a := range as {
		if len(args) < i {
			return fmt.Errorf("need %d arg as %s", i, camelcase.UpperSnakeCase(a.Name))
		}

		if a.HasVariadic() {
			for _, s := range args[i:] {
				ev := reflect.New(a.Value.Type().Elem()).Elem()
				if err := encodingx.UnmarshalText(ev, []byte(s)); err != nil {
					return err
				}
				a.Value.Set(reflect.Append(a.Value, ev))
			}
		} else {
			if err := encodingx.UnmarshalText(a.Value, []byte(args[i])); err != nil {
				return err
			}
		}
	}
	return nil
}
