// Package otel 提供日志、trace 和 metric 的统一装配入口。
//
// 它负责：
//   - 基于应用信息初始化 logger、tracer 与 meter provider
//   - 将观测对象注入运行时上下文
//   - 协调观测生命周期的初始化与关闭
//
// 它不负责：
//   - 定义业务级 metric 名称和采样策略
//   - 替代上层服务的日志语义设计
//
//go:generate go tool gen .
package otel
