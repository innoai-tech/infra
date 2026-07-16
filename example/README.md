# Example Layout

`internal/example` 是仓库内的第一参考实现，用来表达当前推荐的应用分层方式。

## Boundary

- `cmd/example`:
  即 [internal/example/cmd/example](./cmd/example)，承载示例应用入口、命令注册、routes 组装与静态资源承载。
- `pkg/apis`: 示例域的公开数据契约。
- `pkg/endpoints`: 示例域的 endpoint 契约。
- `domain`: 示例业务实现与 provider。

示例目录用于说明框架推荐用法，不作为 root `pkg/*` 的稳定公共 API。

## Reading Order

- 先看 [cmd/example](./cmd/example) 了解入口与运行形态。
- 再看 [pkg/apis](./pkg/apis) 与 [pkg/endpoints](./pkg/endpoints) 了解 courier 契约层。
- 最后看 [domain](./domain) 了解实现如何被 routes 连接。
