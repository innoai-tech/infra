/*
Package internal GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package internal

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

func (v Arg) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Name":
			return []string{}, true
		case "Value":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (Args) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}

func (v FlagVar) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Name":
			return []string{}, true
		case "Alias":
			return []string{}, true
		case "Required":
			return []string{}, true
		case "EnvVar":
			return []string{}, true
		case "Desc":
			return []string{}, true
		case "Value":
			return []string{}, true
		case "EnumValues":
			return []string{}, true
		case "Secret":
			return []string{}, true
		case "Expose":
			return []string{}, true
		case "Volume":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (v Tag) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Name":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(v.Values, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{}, true
}