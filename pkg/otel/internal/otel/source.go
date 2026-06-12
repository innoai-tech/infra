package otel

import (
	"fmt"
	"log/slog"
	"path"
	"runtime"

	"go.opentelemetry.io/otel/log"
)

// GetSource 获取调用栈的源码位置信息。
func GetSource(skip int) Source {
	pc, _, _, _ := runtime.Caller(skip + 1)
	fs := runtime.CallersFrames([]uintptr{pc})
	f, _ := fs.Next()

	return Source{
		Function: f.Function,
		File:     f.File,
		Line:     f.Line,
	}
}

// Source 表示源码位置，作为 slog.Source 的别名。
type Source slog.Source

// AsKeyValues 将源码位置转换为 OpenTelemetry 日志键值对。
func (s Source) AsKeyValues() []log.KeyValue {
	return []log.KeyValue{
		log.String("source.func", s.Function),
		log.String("source.file", fmt.Sprintf("%s:%d", path.Base(s.File), s.Line)),
	}
}
