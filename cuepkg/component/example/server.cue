package example

import (
	kubepkg "github.com/octohelm/kubepkgspec/cuepkg/kubepkg"
)

#Server: kubepkg.#KubePkg & {
	metadata: name: string | *"server"
	spec: {
		version: _

		deploy: kind: "Deployment"

		deploy: spec: replicas: _ | *1

		config: EXAMPLE_LOG_LEVEL:                       string | *"info"
		config: EXAMPLE_TRACE_COLLECTOR_ENDPOINT:        string | *""
		config: EXAMPLE_METRIC_COLLECTOR_ENDPOINT:       string | *""
		config: EXAMPLE_METRIC_COLLECT_INTERVAL_SECONDS: string | *"0"
		config: EXAMPLE_SERVER_ENABLE_DEBUG:             string | *"false"

		services: "#": ports: containers."server".ports

		containers: server: {

			ports: http: _ | *80

			env: EXAMPLE_SERVER_ADDR: _ | *":\(ports."http")"

			readinessProbe: {
				httpGet: {
					path:   "/"
					port:   ports."http"
					scheme: "HTTP"
				}
				initialDelaySeconds: _ | *5
				timeoutSeconds:      _ | *1
				periodSeconds:       _ | *10
				successThreshold:    _ | *1
				failureThreshold:    _ | *3
			}
			livenessProbe: readinessProbe
		}

		containers: server: {
			image: {
				name: _ | *"ghcr.io/octohelm/example"
				tag:  _ | *"\(version)"
			}

			args: [
				"serve",
			]
		}
	}
}
