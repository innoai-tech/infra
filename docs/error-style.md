# Error Style

这份文档定义 `infra` 当前推荐的错误返回风格，目标是让初始化失败、配置失败、运行失败在日志和测试里都更容易定位。

## Rule

优先使用下面的格式：

`<stage> <target>: %w`

例如：

- `init *otel.Otel: invalid log level: ...`
- `run *example.SyncJob: context deadline exceeded`
- `serve *http.Server: listen tcp ...`
- `shutdown *agent.Agent: context deadline exceeded`

## Why

当前框架大量通过接口和组合对象驱动生命周期，如果只返回原始错误：

- 看不出是在 `Init`、`Run`、`Serve` 还是 `Shutdown` 阶段失败。
- 看不出是哪个 configurator / server 失败。
- 在并发生命周期里尤其难排查。

## Current Scope

这一轮已经把 `pkg/configuration` 主生命周期路径统一为这个风格：

- `Init`
- `RunOrServe`
- `Shutdown`
- `PostServeRun`

并保持 `errors.Is` / `errors.As` 可用。

## Additional Guidelines

- 配置校验失败：
  直接说明配置项和原因，例如 `invalid cron spec: ...`、`index.html not found in root dir ...`
- 环境变量注入失败：
  说明来源 env var，例如 `set value from DEMO_ADDR failed: ...`
- HTTP handler 级错误：
  除非需要转换成状态码，否则优先保留原始错误并在外层统一包装。
- 不要为了“统一”而把所有错误都变成字符串；优先 `%w` 保留原始错误链。

## Non-goals

- 这份约定不要求所有业务错误都改成同一文本模板。
- 这份约定不覆盖 `statuserror.Wrap` 这类本来就带 HTTP 语义的错误类型。
