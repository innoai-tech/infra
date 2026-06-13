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
			// 监听地址
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
			// +optional
			"EXAMPLE_LOG_LEVEL": {
				Value: "info",
			},
			// 日志格式
			// +optional
			"EXAMPLE_LOG_FORMAT": {
				Value: "json",
			},
			// 设置后将启用 trace 采集
			// +optional
			"EXAMPLE_TRACE_COLLECTOR_ENDPOINT": {
				Value: "",
			},
			// 指标采集器地址
			// +optional
			"EXAMPLE_METRIC_COLLECTOR_ENDPOINT": {
				Value: "",
			},
			// 指标采集间隔（秒）
			// +optional
			"EXAMPLE_METRIC_COLLECT_INTERVAL_SECONDS": {
				Value: "0",
			},
			// 启用调试模式
			// +optional
			"EXAMPLE_SERVER_ENABLE_DEBUG": {
				Value: "false",
			},
			// 监听地址
			"EXAMPLE_SERVER_ADDR": {
				ValueRef: `:{{ .Ports["http"].Port }}`,
			},
		},
	}

	return c
}
