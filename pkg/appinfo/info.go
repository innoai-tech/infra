package appinfo

import (
	"fmt"
	"net/url"
)

// Info 描述应用及其当前组件的元数据。
//
// 它通常由 CLI、HTTP Server、遥测等运行面共享，用于统一服务名、
// 组件名和镜像命名空间等基础信息。
// +gengo:injectable:provider
type Info struct {
	// App 关联的应用元数据
	App       *App
	// Name 命令或组件名称
	Name      string
	// Desc 描述信息
	Desc      string
	// Component 关联的组件信息
	Component *Component
}

// App 描述应用级元数据。
type App struct {
	// Name 应用名称
	Name           string
	// Version 版本号
	Version        string
	// ImageNamespace 镜像命名空间
	ImageNamespace string
}

// String 返回 `<name>@<version>` 格式的应用标识。
func (a App) String() string {
	return fmt.Sprintf("%s@%s", a.Name, a.Version)
}

// Component 描述当前运行单元及其附加选项。
type Component struct {
	// Name 组件名称
	Name    string
	// Options 组件选项
	Options url.Values
}
