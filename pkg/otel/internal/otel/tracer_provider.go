package otel

import (
	"go.opentelemetry.io/otel/trace"

	contextx "github.com/octohelm/x/context"
)

// TracerProviderContext 是用于上下文注入的 TracerProvider 上下文键。
var TracerProviderContext = contextx.New[TracerProvider]()

// TracerProvider 是对 OpenTelemetry trace.TracerProvider 的类型别名。
type TracerProvider = trace.TracerProvider
