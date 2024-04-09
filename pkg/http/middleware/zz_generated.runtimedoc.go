/*
Package middleware GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package middleware

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (CORSOption) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{
		"CORSOption represents a functional option for configuring the CORS middleware.",
	}, true
}
func (OriginValidator) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{
		"OriginValidator takes an origin string and returns whether or not that origin is allowed.",
	}, true
}
