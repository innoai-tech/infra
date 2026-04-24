package v0

import (
	"testing"
	"time"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"
)

func TestOrgTypes(t *testing.T) {
	t.Parallel()

	now := time.Now()
	detail := Detail{
		Info: Info{
			Name: "demo",
			Type: TYPE__GOV,
		},
		CreatedAt: &now,
	}

	list := DataList{
		Data:  []Info{detail.Info},
		Total: 1,
	}

	Then(t, "公开数据类型保持预期字段和值",
		Expect(detail.Name, Equal("demo")),
		Expect(detail.Type, Equal(TYPE__GOV)),
		Expect(list.Total, Equal(1)),
		Expect(list.Data, Be(cmp.Len[[]Info](1))),
		Expect(TYPE__COMPANY > TYPE__GOV, Equal(true)),
	)
}
