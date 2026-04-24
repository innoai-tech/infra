// Package configuration 提供 singleton 初始化、上下文注入和生命周期编排能力。
//
// 它负责：
//   - 统一驱动 `SetDefaults -> Init -> InjectContext -> Run/Serve -> Shutdown`
//   - 组合多个 configurator 的上下文注入链
//   - 处理 disabled、shutdown timeout 和 server 生命周期编排
//
// 它不负责：
//   - 决定具体业务对象如何拆分
//   - 替代上层命令或服务的入口组织
//
//go:generate go tool gen .
package configuration
