// Package cron 提供面向配置层的 cron 表达式公共类型。
//
// 它负责在配置和运行面之间传递 cron spec，并提供最小的解析、
// 调度和序列化能力，不负责具体任务编排。
//
//go:generate go tool gen .
package cron
