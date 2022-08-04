package cli

import (
	"net/url"
	"strings"
)

func parseTag(v string) *tag {
	parts := strings.Split(v, ",")
	t := &tag{
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

type tag struct {
	Name string
	url.Values
}
