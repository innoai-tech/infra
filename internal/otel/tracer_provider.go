package otel

import (
	"go.opentelemetry.io/otel/trace"

	contextx "github.com/octohelm/x/context"
)

var TracerProviderContext = contextx.New[TracerProvider]()

type TracerProvider = trace.TracerProvider
