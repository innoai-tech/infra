package exampleserve

import (
	"github.com/innoai-tech/runtime/cuepkg/kube"
)

#ExampleServe: kube.#App & {
	app: {
		name: "example-serve"
	}
	config: "EXAMPLE_LOG_LEVEL":                string | *"info"
	config: "EXAMPLE_LOG_FILTER":               string | *"Always"
	config: "EXAMPLE_TRACE_COLLECTOR_ENDPOINT": string | *""
	config: "EXAMPLE_SERVER_ADDR":              string | *":80"
	config: "EXAMPLE_SERVER_ENABLE_DEBUG":      bool | *false

	services: "\(app.name)": {
		selector: "app": app.name
		ports: containers."example-serve".ports
	}

	containers: "example-serve": {
		image: {
			name: _ | *"ghcr.io/octohelm/example-serve"
			tag:  _ | *"\(app.version)"
		}
		ports: http: _ | *80
		readinessProbe: kube.#ProbeHttpGet & {
			httpGet: {path: "/", port: ports.http}
		}
		livenessProbe: readinessProbe
		args: [
			"serve",
		]
	}
}
