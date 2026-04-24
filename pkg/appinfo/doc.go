// Package appinfo 定义应用、命令与组件元数据。
//
// 这些元数据会被 cli、http、otel 等包复用，用于服务名、组件信息、
// 镜像命名空间和注入元数据的统一传递。
//
//go:generate go tool gen .
package appinfo
