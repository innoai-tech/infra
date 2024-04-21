/*
Package metric GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package metric

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (v Metric) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Name":
			return []string{}, true
		case "Unit":
			return []string{}, true
		case "Description":
			return []string{}, true
		case "Views":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v View) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Instrument":
			return []string{}, true
		case "Stream":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}