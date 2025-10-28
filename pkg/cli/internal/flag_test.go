package internal

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/octohelm/x/testing/bdd"
)

type SomeStruct struct {
	Name string
}

func TestReflect(t *testing.T) {
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

	spew.Dump(x)
}

func TestFlagVar(t *testing.T) {
	bdd.FromT(t).Given("slice flag var", func(b bdd.T) {
		v := &FlagVar{
			Name:  "list",
			Value: reflect.New(reflect.TypeFor[[]string]()).Elem(),
		}

		b.Then("could set single value",
			bdd.NoError(v.Set("1")),
			bdd.Equal([]string{"1"}, v.Value.Interface().([]string)),
		)

		v1 := &FlagVar{
			Name:  "list",
			Value: reflect.New(reflect.TypeFor[[]string]()).Elem(),
		}
		b.Then("could set multiple value",
			bdd.NoError(v1.Set("1,2,3")),
			bdd.Equal([]string{"1", "2", "3"}, v1.Value.Interface().([]string)),
		)

		v2 := &FlagVar{
			Name:  "list",
			Value: reflect.New(reflect.TypeFor[[]string]()).Elem(),
		}
		b.Then("could set multiple value contains comma",
			bdd.NoError(v2.Set(`"1,1","2,2",3`)),
			bdd.Equal([]string{"1,1", "2,2", "3"}, v2.Value.Interface().([]string)),
		)
	})
}
