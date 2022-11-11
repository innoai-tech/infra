/*
Package webapp GENERATED BY gengo:runtimedoc 
DON'T EDIT THIS FILE
*/
package webapp

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (OptFunc) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}
func (v Server) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "Env":
			return []string{
				"app env name",
			}, true
		case "BaseHref":
			return []string{
				"base href",
			}, true
		case "Config":
			return []string{
				"config",
			}, true
		case "DisableHistoryFallback":
			return []string{
				"Disable http history fallback, only used for static pages",
			}, true
		case "DisableCSP":
			return []string{
				"Disable Content-Security-Policy",
			}, true
		case "Root":
			return []string{
				"AppRoot for host in fs",
			}, true
		case "Addr":
			return []string{
				"Webapp serve on",
			}, true

		}

		return nil, false
	}
	return []string{}, true
}
