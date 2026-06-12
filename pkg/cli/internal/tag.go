package internal

import (
	"net/url"
	"strings"
)

// Tag 表示 struct tag 中解析出的名值对信息。
type Tag struct {
	Name string

	url.Values
}

// ParseTag 解析逗号分隔的名值对标签字符串。
func ParseTag(v string) *Tag {
	parts := strings.Split(v, ",")
	t := &Tag{
		Values: url.Values{},
	}

	for i, v := range parts {
		if i == 0 {
			t.Name = v
			continue
		}

		kv := strings.SplitN(v, "=", 2)
		if len(kv) == 2 {
			t.Values[kv[0]] = append(t.Values[kv[0]], kv[1])
		} else {
			t.Values[kv[0]] = []string{}
		}
	}

	return t
}
