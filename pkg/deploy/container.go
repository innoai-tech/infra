package deploy

import (
	"bytes"
	"text/template"
)

// Port 描述一个容器端口。
type Port struct {
	Port              int    // 容器端口号
	Protocol          string // 默认 "TCP"
	Endpoint          string // 服务入口路径
	ReadinessEndpoint string // 就绪探针路径
	LivenessEndpoint  string // 存活探针路径
}

// EnvVar 描述一个环境变量配置，支持静态值和模板引用。
type EnvVar struct {
	Value    string // 静态值
	ValueRef string // Go template，以 Container 为数据上下文
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
	ImageName string            // 镜像名，如 "ghcr.io/octohelm/example"
	Version   string            // 版本号，如 "1.0.0"
	Command   []string          // 入口命令
	Args      []string          // 命令参数
	Ports     map[string]Port   // 暴露端口，key 为端口名
	Env       map[string]EnvVar // 环境变量，key 为变量名
}
