package otel

import (
	"go.opentelemetry.io/otel/log"

	contextx "github.com/octohelm/x/context"
)

// LoggerProvider 是对 OpenTelemetry log.LoggerProvider 的类型别名。
type LoggerProvider = log.LoggerProvider

// LoggerProviderContext 是用于上下文注入的 LoggerProvider 上下文键。
var LoggerProviderContext = contextx.New[LoggerProvider]()

// LogLevel 表示日志输出级别。
// +gengo:enum
type LogLevel string

const (
	// ErrorLevel 表示仅输出 error 级别日志。
	ErrorLevel LogLevel = "error"
	// WarnLevel 表示输出 warn 及以上级别日志。
	WarnLevel  LogLevel = "warn"
	// InfoLevel 表示输出 info 及以上级别日志。
	InfoLevel  LogLevel = "info"
	// DebugLevel 表示输出 debug 及以上级别日志。
	DebugLevel LogLevel = "debug"
)

// LogFormat 表示日志输出格式。
// +gengo:enum
type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)
