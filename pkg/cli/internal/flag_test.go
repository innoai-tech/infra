package internal

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"

	. "github.com/octohelm/x/testing/v2"
)

type SomeStruct struct {
	Name string
}

func TestReflect(t *testing.T) {
	t.Run("reflect map element mutation", func(t *testing.T) {
		x := struct {
			Map map[string]*SomeStruct
		}{
			Map: map[string]*SomeStruct{
				"A": {},
			},
		}

		rv := reflect.ValueOf(&x).Elem()
		m := rv.FieldByName("Map").MapRange()

		for m.Next() {
			m.Value().Elem().FieldByName("Name").SetString("aa")
		}

		Then(t, "map element should be mutated",
			Expect(x.Map["A"].Name, Equal("aa")),
		)

		if testing.Verbose() {
			spew.Dump(x)
		}
	})
}

func TestFlagVar(t *testing.T) {
	t.Run("GIVEN a slice flag var", func(t *testing.T) {
		t.Run("WHEN setting a single value", func(t *testing.T) {
			v := &FlagVar{
				Name:  "list",
				Value: reflect.New(reflect.TypeFor[[]string]()).Elem(),
			}

			Must(t, func() error { return v.Set("1") })

			Then(t, "it should contain one element",
				Expect(v.Value.Interface().([]string), Equal([]string{"1"})),
			)
		})

		t.Run("WHEN setting multiple values via comma", func(t *testing.T) {
			v := &FlagVar{
				Name:  "list",
				Value: reflect.New(reflect.TypeFor[[]string]()).Elem(),
			}

			Must(t, func() error { return v.Set("1,2,3") })

			Then(t, "it should parse into a slice",
				Expect(v.Value.Interface().([]string), Equal([]string{"1", "2", "3"})),
			)
		})

		t.Run("WHEN setting values with quoted commas", func(t *testing.T) {
			v := &FlagVar{
				Name:  "list",
				Value: reflect.New(reflect.TypeFor[[]string]()).Elem(),
			}

			Must(t, func() error { return v.Set(`"1,1","2,2",3`) })

			Then(t, "it should respect quotes and ignore escaped commas",
				Expect(v.Value.Interface().([]string), Equal([]string{"1,1", "2,2", "3"})),
			)
		})
	})
}
