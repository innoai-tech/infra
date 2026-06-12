package example

import "github.com/innoai-tech/infra/pkg/deploy"

// Preset 返回预设的容器部署规格。
func Preset() *deploy.Container {
	c := &deploy.Container{
		ImageName: "ghcr.io/octohelm/example",
		Version:   "1.0.0",
		Command:   []string{"example"},
		Args: []string{
			"serve",
		},
		Ports: map[string]deploy.Port{
			"http": {
				Port:              80,
				Protocol:          "TCP",
				Endpoint:          "/",
				ReadinessEndpoint: "/",
				LivenessEndpoint:  "/",
			},
		},
		Env: map[string]deploy.EnvVar{
			// 日志级别
			"EXAMPLE_LOG_LEVEL": {
				Value: "info",
			},
			// 日志格式
			"EXAMPLE_LOG_FORMAT": {
				Value: "json",
			},
			// 设置后将启用 trace 采集
			"EXAMPLE_TRACE_COLLECTOR_ENDPOINT": {
				Value: "",
			},
			"EXAMPLE_METRIC_COLLECTOR_ENDPOINT": {
				Value: "",
			},
			"EXAMPLE_METRIC_COLLECT_INTERVAL_SECONDS": {
				Value: "0",
			},
			"EXAMPLE_SERVER_ENABLE_DEBUG": {
				Value: "false",
			},
			"EXAMPLE_SERVER_ADDR": {
				ValueRef: `:{{ .Ports["http"].Port }}`,
			},
		},
	}

	return c
}
