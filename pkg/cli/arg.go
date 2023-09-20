package cli

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/octohelm/gengo/pkg/camelcase"
	encodingx "github.com/octohelm/x/encoding"
)

type arg struct {
	Name  string
	Value reflect.Value
}

func (as arg) HasVariadic() bool {
	return as.Value.Kind() == reflect.Slice
}

type args []*arg

func (as args) String() string {
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

func (as args) Validate(args []string) error {
	for i := range as {
		a := as[i]

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
