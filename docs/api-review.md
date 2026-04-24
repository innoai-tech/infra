# API Review Notes

这份文档记录当前 root `pkg/*` 的公开 API 复盘结果，目标是给后续增量演进一个稳定基线。

## `pkg/cli`

当前建议视为公开入口的只有：

- `NewApp`
- `WithImageNamespace`
- `AddTo`
- `Exec`
- `Execute`
- `Command`
- `C`

其余需要谨慎对待：

- `CanPreRun` 目前属于可选扩展点，但仓库内使用并不广。
- `CanRuntimeDoc` 主要为反射收集说明文本服务，更像框架内部约定，不建议在业务层大量自定义扩散。
- `pkg/cli/internal` 明确不是公共边界，即使位于 `pkg/cli` 之下，也不应作为外部依赖面。

结论：

- `pkg/cli` 的公开入口已经足够小，后续应避免再把 tag 解析、flag/env 细节类型提升到 root package。

## `pkg/configuration`

当前接口集合围绕生命周期能力拆分：

- `Defaulter`
- `CanInit`
- `ContextInjector`
- `Runner`
- `Server`
- `CanShutdown`
- `WithShutdownTimeout`
- `CanDisabled`
- `PostServeRunner`

这套拆分的优点是组合灵活，缺点是名字都偏“能力型”，初次阅读时容易混淆：

- `CanInit` / `CanShutdown` 是生命周期阶段。
- `ContextInjector` 是上下文能力。
- `CanDisabled` / `WithShutdownTimeout` 是修饰性能力。

结论：

- 当前接口数量仍可接受，不建议在没有明显重复语义前继续新增同类能力接口。
- 后续若要收敛，应优先通过文档和示例澄清语义，而不是急于改名。

## `pkg/http`

`pkg/http.Server` 当前公开的装配点主要有：

- 路由：`ApplyRouter`
- 服务名：`SetName`
- TLS：`SetTLSProvoder`
- 中间件：`ApplyRouterHandlers` / `ApplyGlobalHandlers`
- CORS：`SetCorsOptions`
- 指标 reader：`SetMetricReader`

这轮新增的 `SetMetricReader` 解决了一个实际问题：

- 之前 `Server` 完全依赖上下文里预先注入 metric reader。
- 现在 standalone 场景可以显式设置，也会在缺省时安全回退到 manual reader。

结论：

- `pkg/http` 的显式装配入口已经基本够用，后续应优先增加 setter / option，而不是继续扩大隐式上下文依赖。

## `pkg/otel`

`pkg/otel` 目前承担两层职责：

- 观测 provider 的初始化与关闭。
- 把 logger / tracer / meter / metric reader 注入运行时上下文。

它依然是一个偏“生命周期型 singleton”，而不是一个纯构造器包。

结论：

- 这类包继续保留 singleton 形态是合理的。
- 后续如果要增强显式构造入口，应以新增 helper 为主，不要破坏现有 `configuration` 接入模型。

## Recommended Direction

后续 public API 演进遵循三条规则：

1. 新增公开能力优先做成 root package 的显式入口，而不是让调用方必须依赖隐藏的上下文前置条件。
2. `pkg/*` 只暴露“推荐直接使用”的入口，辅助细节尽量留在子包或 `internal/*`。
3. 当行为已经强依赖反射或生成时，必须同时补文档或示例，不让调用方只靠猜。
