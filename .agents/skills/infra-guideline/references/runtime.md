# Runtime

这份文档说明 `infra` 的运行模型和最小组合方式，供调用方 agent 直接使用，不依赖额外翻找实现文档。

## 生命周期顺序

`configuration` 会按固定顺序驱动 configurator：

1. `SetDefaults`
   只补默认值，不做外部副作用。
2. `Init`
   校验配置、建立内部状态。
3. `InjectContext`
   把当前 configurator 注入上下文，供后续阶段读取。
4. `Run`
   执行一次性逻辑。
5. `Serve`
   并行启动长生命周期服务。
6. `Shutdown`
   做优雅关闭。

关键规则：

- `Run` 总在 `Serve` 之前。
- 只有存在启用中的 server 时，才会进入阻塞的 serve 阶段。
- 被 `Disabled(ctx) == true` 判掉的对象不会进入 `Serve` 或 `Shutdown`。

## 最常见的组合方式

最常见的职责拆分是：

- `cli` 负责入口、flag、env。
- `configuration` 负责 lifecycle orchestration。
- `http` / `agent` / `otel` 负责具体运行时能力。

典型结构：

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

如果某个命令只需要一次性执行逻辑，可以只有 `Run` 能力；不必强行引入 server。

## 什么时候该把对象做成 configurator 字段

满足下面任一条件时，优先作为命令 struct 字段参与生命周期，而不是在方法体里临时 new：

- 需要 `SetDefaults` 或 `Init`。
- 需要读取 env / flags / configuration 注入值。
- 需要统一 `Shutdown`。
- 需要和其他 configurator 共享 context 注入链。

只有短小、纯函数式、没有生命周期状态的局部 helper，才适合留在方法体内直接调用。

## 错误表达规则

生命周期相关错误优先使用：

`<stage> <target>: %w`

例如：

- `init *otel.Otel: invalid log level: ...`
- `run *example.SyncJob: context deadline exceeded`
- `serve *http.Server: listen tcp ...`
- `shutdown *agent.Agent: context deadline exceeded`

这样做的目的是让 agent 在日志和测试里立刻知道：

- 失败发生在哪个阶段。
- 是哪个 configurator 或 server 失败。

不要为了统一把原始错误拍平成字符串；优先保留 `%w` 错误链。

## 上下文规则

- 需要向后续阶段暴露值时，用 `InjectContext`。
- 需要把已有注入链带到后台 goroutine 时，用 `configuration.Background(ctx)`。
- `CurrentInstanceFromContext` 只应用于初始化阶段识别“当前正在初始化的对象”，不要把它当成通用 service locator。

## 公开 API 核对入口

如果需要核对生命周期入口、上下文函数或公开类型，优先使用：

- `go doc github.com/innoai-tech/infra/pkg/configuration`
- `go doc github.com/innoai-tech/infra/pkg/cli`
- `go doc github.com/innoai-tech/infra/pkg/http`
- `go doc github.com/innoai-tech/infra/pkg/http/webapp`
- `go doc github.com/innoai-tech/infra/pkg/otel`
- `go doc github.com/innoai-tech/infra/pkg/agent`

如果这里描述的生命周期约定、错误包装方式或推荐组合发生变化，应先保证对应公共包的 `doc.go` 已同步，让 `go doc` 能看到关键信息。

`go doc` 只用于确认公开 API；不要为了理解基本生命周期而先跳进具体实现。
