// Package agent 提供一组可接入 configuration 生命周期的后台 worker 运行模型。
//
// 它负责：
//   - 将 agent / worker 任务纳入统一启动和关闭流程
//   - 处理 context 取消、panic 保护和优雅退出
//   - 让后台任务与上层 app 的 lifecycle 保持一致
//
// 它不负责：
//   - 定义具体业务任务和调度策略
//   - 替代上层命令、server 或 domain service 的职责划分
//   - 提供跨进程任务队列或分布式调度能力
//
//go:generate go tool gen .
package agent
