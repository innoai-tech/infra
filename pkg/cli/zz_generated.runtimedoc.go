/*
Package cli GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package cli

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (v App) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Name":
			return []string{}, true
		case "Version":
			return []string{}, true
		case "ImageNamespace":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v Component) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Name":
			return []string{}, true
		case "Options":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v Info) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "App":
			return []string{}, true
		case "Name":
			return []string{}, true
		case "Desc":
			return []string{}, true
		case "Component":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}
