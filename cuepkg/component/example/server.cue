
package example

import (
	"github.com/innoai-tech/runtime/cuepkg/kube"
)

#Server: kube.#App & {
	app: {
		name: string | *"server"
	}
  
	config: "EXAMPLE_LOG_LEVEL": string | *"info"
  
	config: "EXAMPLE_LOG_FILTER": string | *"Always"
  
	config: "EXAMPLE_TRACE_COLLECTOR_ENDPOINT": string | *""
  
	config: "EXAMPLE_SERVER_ENABLE_DEBUG": string | *"false"

	services: "\(app.name)": {
		selector: "app": app.name
		ports:     containers."server".ports
	}

	containers: "server": {

		ports: "http": _ | *80

		env: "EXAMPLE_SERVER_ADDR": _ | *":\(ports."http")"

		readinessProbe: kube.#ProbeHttpGet & {
			httpGet: {path: "/", port: ports."http"}
		}
		livenessProbe: readinessProbe
  
	}
  
	containers: "server": {
		image: {
			name: _ | *"ghcr.io/octohelm/example"
			tag:  _ | *"\(app.version)"
		}

		args: [
		"serve",
		]
	}
}