package v0

import (
	"reflect"
	"testing"

	apiv0 "github.com/innoai-tech/infra/internal/example/pkg/apis/org/v0"
	. "github.com/octohelm/x/testing/v2"
)

func TestOrgEndpoints(t *testing.T) {
	t.Parallel()

	createType := reflect.TypeOf(CreateOrg{})
	getType := reflect.TypeOf(GetOrg{})
	deleteType := reflect.TypeOf(DeleteOrg{})

	Then(t, "组织 endpoint 暴露预期的 body 与 path 参数",
		Expect(createType.Field(1).Type, Equal(reflect.TypeOf(apiv0.Info{}))),
		Expect(createType.Field(1).Tag.Get("in"), Equal("body")),
		Expect(getType.Field(1).Tag.Get("name"), Equal("orgName")),
		Expect(getType.Field(1).Tag.Get("in"), Equal("path")),
		Expect(deleteType.Field(1).Tag.Get("name"), Equal("orgName")),
	)
}
