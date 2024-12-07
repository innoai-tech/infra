/*
Package appinfo GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package appinfo

// nolint:deadcode,unused
func runtimeDoc(v any, prefix string, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		doc, ok := c.RuntimeDoc(names...)
		if ok {
			if prefix != "" && len(doc) > 0 {
				doc[0] = prefix + doc[0]
				return doc, true
			}

			return doc, true
		}
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
