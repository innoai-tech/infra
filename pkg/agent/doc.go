// Package agent 提供一组可优雅启动和关闭的后台 worker 运行模型。
//
// 它适合承载在 configuration 生命周期中运行的 agent / worker 任务，
// 重点处理启动、取消、panic 保护与优雅退出。
//
//go:generate go tool gen .
package agent
