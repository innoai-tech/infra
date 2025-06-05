module github.com/innoai-tech/infra

go 1.24.3

tool github.com/innoai-tech/infra/cmd/example

tool (
	github.com/innoai-tech/infra/internal/cmd/devtool
	mvdan.cc/gofumpt
)

require (
	github.com/go-courier/logr v0.3.2
	github.com/innoai-tech/openapi-playground v0.0.0-20250422032150-6455c64a7612
	github.com/octohelm/courier v0.0.0-20250605014724-269e1e89a1c1
	github.com/octohelm/enumeration v0.0.0-20250115091102-ddf41951d733
	github.com/octohelm/gengo v0.0.0-20250418031154-d14a228621d4
	github.com/octohelm/x v0.0.0-20250409031213-9c254440c2b8
)

require (
	cuelang.org/go v0.13.0
	github.com/andybalholm/brotli v1.1.1
	github.com/fatih/color v1.18.0
	github.com/felixge/httpsnoop v1.0.4
	github.com/go-json-experiment/json v0.0.0-20250517221953-25912455fbc8
	github.com/prometheus/client_golang v1.22.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/spf13/cobra v1.9.1
	github.com/spf13/pflag v1.0.6
	go.opentelemetry.io/contrib/propagators/b3 v1.36.0
	go.opentelemetry.io/otel v1.36.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.36.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.36.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.36.0
	go.opentelemetry.io/otel/exporters/prometheus v0.58.0
	go.opentelemetry.io/otel/log v0.12.2
	go.opentelemetry.io/otel/metric v1.36.0
	go.opentelemetry.io/otel/sdk v1.36.0
	go.opentelemetry.io/otel/sdk/log v0.12.2
	go.opentelemetry.io/otel/sdk/metric v1.36.0
	go.opentelemetry.io/otel/trace v1.36.0
	golang.org/x/net v0.40.0
	golang.org/x/sync v0.14.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.2 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cockroachdb/apd/v3 v3.2.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/juju/ansiterm v1.0.0 // indirect
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/onsi/gomega v1.37.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.64.0 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/proto/otlp v1.7.0 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/tools v0.33.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/grpc v1.72.2 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apimachinery v0.33.1 // indirect
	mvdan.cc/gofumpt v0.8.0 // indirect
)
