package otel

import (
	contextx "github.com/octohelm/x/context"
	"go.opentelemetry.io/otel/log"
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
