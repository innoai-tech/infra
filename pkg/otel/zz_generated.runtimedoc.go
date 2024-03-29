/*
Package otel GENERATED BY gengo:runtimedoc
DON'T EDIT THIS FILE
*/
package otel

// nolint:deadcode,unused
func runtimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}

func (LogLevel) RuntimeDoc(names ...string) ([]string, bool) {
	return []string{}, true
}
func (v Otel) RuntimeDoc(names ...string) ([]string, bool) {
	if len(names) > 0 {
		switch names[0] {
		case "LogLevel":
			return []string{
				"Log level",
			}, true
		case "ExportFilter":
			return []string{
				"Log filter",
			}, true
		case "TraceCollectorEndpoint":
			return []string{
				"When set, will collect traces",
			}, true

		}

		return nil, false
	}
	return []string{}, true
}
