package otel

import (
	"fmt"
	"log/slog"
	"path"
	"runtime"

	"go.opentelemetry.io/otel/log"
)

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

type Source slog.Source

func (s Source) AsKeyValues() []log.KeyValue {
	return []log.KeyValue{
		log.String("source.func", s.Function),
		log.String("source.file", fmt.Sprintf("%s:%d", path.Base(s.File), s.Line)),
	}
}
