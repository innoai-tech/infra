package v0

import (
	"reflect"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestArchiveEndpoint(t *testing.T) {
	t.Parallel()

	tpe := reflect.TypeOf(ArchiveZip{})

	Then(t, "archive endpoint 保持 zip 下载路径",
		Expect(tpe.Field(0).Tag.Get("path"), Equal("/archive/zip")),
	)
}
