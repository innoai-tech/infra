/*
Package main GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package main

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

func (v *Serve) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Server":
			return []string{}, true

		}
		if doc, ok := runtimeDoc(&v.Otel, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{"Start serve"}, true
}

func (v *Webapp) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {

		}
		if doc, ok := runtimeDoc(&v.Otel, "", names...); ok {
			return doc, ok
		}
		if doc, ok := runtimeDoc(&v.Server, "", names...); ok {
			return doc, ok
		}

		return nil, false
	}
	return []string{"Start webapp serve"}, true
}
