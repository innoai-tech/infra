/*
Package cron GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package cron

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (v IntervalSchedule) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Interval":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v Job) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Cron":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}
