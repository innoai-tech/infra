/*
Package configuration GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package configuration

import _ "embed"

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

func (v *Singleton) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Name":
			return []string{}, true
		case "Configurator":
			return []string{}, true

		}

		return nil, false
	}
	return []string{}, true
}

func (*Singletons) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}
