# Runtime

理解这类任务时，优先参考以下文件：

1. [docs/core.md](../../../docs/core.md)
   这里定义了 `SetDefaults -> Init -> InjectContext -> Run/Serve -> Shutdown` 的时序。
2. [docs/error-style.md](../../../docs/error-style.md)
   这里定义了初始化、运行和关闭阶段的错误表达方式。
3. [internal/example/cmd/example](../../../internal/example/cmd/example)
   这里给出 `var App + func init()`、`Serve`、`Webapp` 和 routes 组装的实际写法。

本仓库当前推荐：

- 命令入口用 `cli.NewApp(...)` + `cli.AddTo(...)` 组装。
- 生命周期用 `configuration` 接口组合驱动。
- HTTP 服务优先使用 `pkg/http.Server` 或 `pkg/http/webapp.Server`。
- 遥测接入优先使用 `otel.Otel`。
- 后台 worker 优先使用 `agent.Agent`。
