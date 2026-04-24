// Package webapp 提供面向 SPA / 静态站点的 HTTP 托管能力。
//
// 它负责静态资源分发、`index.html` 占位符替换、base href 处理和
// history fallback，不负责前端构建产物本身的生成。
//
//go:generate go tool gen .
package webapp
