---
name: infra-guideline
description: 说明 `github.com/innoai-tech/infra` pkg 中 `cli`、`configuration`、`http`、`otel`、`agent` 与示例实现的推荐用法、生命周期模型和公共边界；当任务涉及基于该 pkg 组装 app、command、server、routes 或判断能力应留在调用方还是沉淀到 infra 公共包时使用。
---

# Infra Guideline

按 `github.com/innoai-tech/infra` 约定组装 CLI、server 和生命周期。

## 启动一个 HTTP 服务

```
cli ──→ configuration ──→ http (承 courier Router)
```

```go
var App = cli.NewApp("myapp", "0.1.0")

func init() {
    cli.AddTo(App, &ServeCmd{})
}

type ServeCmd struct {
    // courierhttp 的 routers 通过 infra/pkg/http.Server 挂载
}
```

**关键约定**：
- 入口通过 `var App + func init()` 注册命令
- 生命周期由 `configuration` 驱动：`SetDefaults → Init → InjectContext → Run → Serve → Shutdown`
- HTTP server 通过 `infra/pkg/http` 封装 courier Router
- 基础设施单例在 `cmd/{app}` 层声明

API 细节以 `go doc` 为准：`go doc github.com/innoai-tech/infra/pkg/cli` / `configuration` / `http`。

## 更多

- 包能力地图 → [references/repo-map.md](references/repo-map.md)
- 生命周期详情 → [references/runtime.md](references/runtime.md)
- 代码应沉淀到 infra 还是留在调用方 → [references/boundaries.md](references/boundaries.md)
