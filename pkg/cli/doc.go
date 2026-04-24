// Package cli 提供基于 struct 声明的命令行入口组装能力。
//
// 它负责：
//   - 从命令 struct 收集 args、flags、env 绑定和运行时文档
//   - 将命令树与 configuration 生命周期拼接为可执行入口
//   - 提供 `dump-k8s`、配置展示等命令层辅助能力
//
// 它不负责：
//   - 定义业务 API 契约和领域逻辑
//   - 替代应用自己的命令分层设计
//
//go:generate go tool gen .
package cli
