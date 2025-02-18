package internal

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/octohelm/gengo/pkg/camelcase"
	encodingx "github.com/octohelm/x/encoding"
)

type Arg struct {
	Name  string
	Value reflect.Value
}

func (as Arg) HasVariadic() bool {
	return as.Value.Kind() == reflect.Slice
}

type Args []*Arg

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
