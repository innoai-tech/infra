package deploy

import (
	"bytes"
	"text/template"
)

// Port 描述一个容器端口。
type Port struct {
	// Port 容器端口号
	Port              int
	// Protocol 默认 "TCP"
	Protocol          string
	// Endpoint 服务入口路径
	Endpoint          string
	// ReadinessEndpoint 就绪探针路径
	ReadinessEndpoint string
	// LivenessEndpoint 存活探针路径
	LivenessEndpoint  string
}

// EnvVar 描述一个环境变量配置，支持静态值和模板引用。
type EnvVar struct {
	// Value 静态值
	Value    string
	// ValueRef Go template，以 Container 为数据上下文
	ValueRef string
}

// ToValue 解析环境变量的实际值。
// 优先使用 ValueRef 执行模板，若为空则返回 Value。
func (e EnvVar) ToValue(c *Container) (string, error) {
	if e.ValueRef != "" {
		t, err := template.New("").Parse(e.ValueRef)
		if err != nil {
			return "", err
		}
		b := &bytes.Buffer{}
		if err := t.Execute(b, c); err != nil {
			return "", err
		}
		return b.String(), nil
	}
	return e.Value, nil
}

// Container 描述一个与平台无关的容器运行规格。
type Container struct {
	// ImageName 镜像名，如 "ghcr.io/octohelm/example"
	ImageName string
	// Version 版本号，如 "1.0.0"
	Version   string
	// Command 入口命令
	Command   []string
	// Args 命令参数
	Args      []string
	// Ports 暴露端口，key 为端口名
	Ports     map[string]Port
	// Env 环境变量，key 为变量名
	Env       map[string]EnvVar
}
