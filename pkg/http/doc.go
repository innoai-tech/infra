// Package http 提供基于 courier 的 HTTP 服务装配能力。
//
// 它负责：
//   - 将 courier router 组装成可运行的 HTTP server
//   - 统一接入 context injector、压缩、日志、指标、pprof 与健康检查中间件
//   - 暴露服务地址、TLS provider 与 router/global handler 的装配入口
//
// 它不负责：
//   - 定义业务 API 契约
//   - 承载业务实现逻辑
//   - 替代上层应用的 routes 组织策略
//
//go:generate go tool gen .
package http
