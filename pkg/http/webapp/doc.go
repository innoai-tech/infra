// Package webapp 提供面向 SPA / 静态站点的 HTTP 托管能力。
//
// 它负责：
//   - 托管嵌入或传入的静态资源文件系统
//   - 处理 `index.html` 占位符替换、base href 和 history fallback
//   - 暴露可接入 configuration 生命周期的 Server
//
// 它不负责：
//   - 生成前端构建产物
//   - 定义前端路由、页面结构或业务状态
//   - 替代普通 HTTP API routes 的组织方式
//
//go:generate go tool gen .
package webapp
