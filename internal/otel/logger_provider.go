package otel

import (
	"go.opentelemetry.io/otel/log"

	contextx "github.com/octohelm/x/context"
)

type LoggerProvider = log.LoggerProvider

var LoggerProviderContext = contextx.New[LoggerProvider]()

// +gengo:enum
type LogLevel string

const (
	ErrorLevel LogLevel = "error"
	WarnLevel  LogLevel = "warn"
	InfoLevel  LogLevel = "info"
	DebugLevel LogLevel = "debug"
)

// +gengo:enum
type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)
