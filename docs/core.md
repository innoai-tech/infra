# Framework Runtime Model

`infra` 的运行模型围绕一组可组合的 configurator 展开。一个 configurator 可以实现 `SetDefaults`、`Init`、`InjectContext`、`Run`、`Serve`、`Shutdown` 中的任意子集，框架按固定时序驱动它们。

## Lifecycle Order

1. `SetDefaults`
   只负责补默认值，不读取外部副作用。
2. `Init`
   负责校验配置、建立内部状态，并可通过 `configuration.CurrentInstanceFromContext` 读取当前正在初始化的实例。
3. `InjectContext`
   将当前 configurator 暴露给后续阶段使用。多个注入器会按初始化顺序串联。
4. `Run`
   顺序执行一次性逻辑，适合启动前准备、索引构建、异步任务注册。
5. `Serve`
   并行启动长生命周期服务。至少有一个启用的 `Server` 时才会进入 serve 模式。
6. `Shutdown`
   在 `Serve` 结束或无 server 的 cleanup 阶段执行优雅关闭。

## Execution Rules

- `Run` 总是先于 `Serve`。
- `Serve` 只会运行 `Disabled(ctx) == false` 的 server。
- 如果全部 server 都被禁用，框架不会进入阻塞的 serve 等待，而是直接进入 cleanup。
- `Shutdown` 也会跳过 `Disabled(ctx) == true` 的对象。
- `WithShutdownTimeout` 可以为单个 configurator 自定义关闭超时。

## Context Model

- `configuration.ContextInjector` 是上下文装配的基础接口。
- `configuration.ComposeContextInjector(...)` 会把多个注入器按顺序组合起来，并自动跳过禁用项。
- `configuration.Background(ctx)` 会从当前上下文复制注入链，用于启动后台 goroutine 或异步 worker。
- `configuration.CurrentInstanceInjectContext` / `CurrentInstanceFromContext` 只用于初始化阶段标识“当前实例”。

## Composition Shape

最常见的组合顺序是：

`cli` 负责入口和 flag/env 绑定，`configuration` 负责生命周期编排，`http`/`agent`/`otel` 作为具体 configurator 接入。

```go
var App = cli.NewApp("example", "0.0.0")

func init() {
	cli.AddTo(App, &Serve{})
}

type Serve struct {
	cli.C `component:"server"`

	otel.Otel
	http.Server
	agent.Agent
}
```

## Recommended Boundaries

- 把稳定可复用能力放在 `pkg/*`。
- 把示例、组装细节、仓库内部实现放在 `internal/*`。
- 当一个类型只为某个 app 的装配服务，不应提升到 root `pkg/*`。
