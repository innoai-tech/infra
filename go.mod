module github.com/innoai-tech/infra

go 1.26.2

tool (
	github.com/innoai-tech/infra/tool/internal/cmd/exported-docs-check
	github.com/innoai-tech/infra/tool/internal/cmd/fmt
	github.com/innoai-tech/infra/tool/internal/cmd/gen
	github.com/innoai-tech/infra/tool/internal/cmd/skills-install
)

// +gengo:import:group=0_controlled
require (
	github.com/innoai-tech/openapi-playground v0.0.0-20251225080706-b73e3d246544
	// +skill:courier-guideline
	github.com/octohelm/courier v0.0.0-20260423104043-41a7d6803925
	// +skill:enumeration-guideline
	github.com/octohelm/enumeration v0.0.0-20260424074548-309e324da628
	// +skill:gengo-guideline
	github.com/octohelm/gengo v0.0.0-20260424104408-7ec241e3024a
	// +skill:testing-guideline
	github.com/octohelm/x v0.0.0-20260423102402-017813b113b1
)

require (
	cuelang.org/go v0.16.1
	github.com/andybalholm/brotli v1.2.1
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/fatih/color v1.19.0
	github.com/felixge/httpsnoop v1.0.4
	github.com/go-json-experiment/json v0.0.0-20260214004413-d219187c3433
	github.com/octohelm/exp v0.0.0-20250610043704-ec5e24647f61
	github.com/prometheus/otlptranslator v1.0.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	go.opentelemetry.io/contrib/instrumentation/host v0.68.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.68.0
	go.opentelemetry.io/contrib/propagators/b3 v1.43.0
	go.opentelemetry.io/otel v1.43.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.43.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.43.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.43.0
	go.opentelemetry.io/otel/log v0.19.0
	go.opentelemetry.io/otel/metric v1.43.0
	go.opentelemetry.io/otel/sdk v1.43.0
	go.opentelemetry.io/otel/sdk/log v0.19.0
	go.opentelemetry.io/otel/sdk/metric v1.43.0
	go.opentelemetry.io/otel/trace v1.43.0
	golang.org/x/net v0.53.0
	golang.org/x/sync v0.20.0
)

require (
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cockroachdb/apd/v3 v3.2.3 // indirect
	github.com/ebitengine/purego v0.10.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/juju/ansiterm v1.0.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20260330125221-c963978e514e // indirect
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.21 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/shirou/gopsutil/v4 v4.26.3 // indirect
	github.com/tklauser/go-sysconf v0.3.16 // indirect
	github.com/tklauser/numcpus v0.11.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	golang.org/x/mod v0.35.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	golang.org/x/tools v0.44.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260420184626-e10c466a9529 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260420184626-e10c466a9529 // indirect
	google.golang.org/grpc v1.80.0 // indirect
	google.golang.org/protobuf v1.36.12-0.20260120151049-f2248ac996af // indirect
	k8s.io/apimachinery v0.36.0 // indirect
	mvdan.cc/gofumpt v0.9.2 // indirect
)
