package otel

import (
	contextx "github.com/octohelm/x/context"
	"go.opentelemetry.io/otel/trace"
)

var TracerProviderContext = contextx.New[TracerProvider]()

type TracerProvider = trace.TracerProvider
