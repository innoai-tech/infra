package otel

import "github.com/innoai-tech/infra/internal/otel"

type OutputFilterType = otel.OutputFilterType

const (
	OutputFilterAlways    = otel.OutputFilterAlways
	OutputFilterOnFailure = otel.OutputFilterOnFailure
	OutputFilterNever     = otel.OutputFilterNever
)

// +gengo:enum
type LogLevel string

const (
	PanicLevel LogLevel = "panic"
	FatalLevel LogLevel = "fatal"
	ErrorLevel LogLevel = "error"
	WarnLevel  LogLevel = "warning"
	InfoLevel  LogLevel = "info"
	DebugLevel LogLevel = "debug"
	TraceLevel LogLevel = "trace"
)
