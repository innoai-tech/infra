package v0

import (
	"reflect"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	apiv0 "github.com/innoai-tech/infra/internal/example/pkg/apis/org/v0"
)

func TestOrgEndpoints(t *testing.T) {
	t.Parallel()

	createType := reflect.TypeFor[CreateOrg]()
	getType := reflect.TypeFor[GetOrg]()
	deleteType := reflect.TypeFor[DeleteOrg]()

	Then(
		t, "组织 endpoint 暴露预期的 body 与 path 参数",
		Expect(createType.Field(1).Type, Equal(reflect.TypeFor[apiv0.Info]())),
		Expect(createType.Field(1).Tag.Get("in"), Equal("body")),
		Expect(getType.Field(1).Tag.Get("name"), Equal("orgName")),
		Expect(getType.Field(1).Tag.Get("in"), Equal("path")),
		Expect(deleteType.Field(1).Tag.Get("name"), Equal("orgName")),
	)
}
