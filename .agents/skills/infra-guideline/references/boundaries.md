# Boundaries

需要判断“该放哪一层”时，优先看这些文件：

1. [docs/boundaries.md](../../../docs/boundaries.md)
   定义 `pkg/*` 与 `internal/*` 的仓库边界。
2. [docs/mechanisms.md](../../../docs/mechanisms.md)
   解释反射、生成和目录约定的交叉点。
3. [docs/api-review.md](../../../docs/api-review.md)
   总结当前 `pkg/cli`、`pkg/configuration`、`pkg/http`、`pkg/otel` 的公开 API 复盘结果。

当前边界的简化判断：

- 可复用、稳定、愿意承担兼容成本的能力，才考虑进入 root `pkg/*`。
- 只服务本仓库示例或单个 app 装配的实现，优先留在 `internal/*`。
- `pkg/http` / `pkg/otel` 这类 root package 允许新增显式入口，但不应继续扩大隐式上下文前置条件。
