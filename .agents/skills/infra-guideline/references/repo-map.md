# Repo Map

这份文档是 `github.com/innoai-tech/infra` 的用户手册入口。目标是让 agent 在别的项目里也能直接判断“这个 pkg 提供什么能力，应该怎么起手”。

## 先建立的心智模型

`infra` 主要提供五类稳定公共能力：

1. `cli`
   负责 app / command 入口、flag 与 env 绑定。
2. `configuration`
   负责生命周期编排，把多个 configurator 串成统一初始化和运行流程。
3. `http`
   负责 HTTP server、本地 webapp 承载和相关 server 组件。
4. `otel`
   负责遥测接入。
5. `agent`
   负责后台 worker 或长生命周期任务。

调用方通常不是一次性把所有能力都接进来，而是按入口目标选择组合：

- 做 CLI 工具：至少用 `cli`，必要时接 `configuration`。
- 做 HTTP 服务：通常用 `cli + configuration + http`，按需补 `otel`、`agent`。
- 做带后台任务的服务：在 HTTP 或 CLI 组合上再接 `agent`。

## 推荐起手方式

先固定入口，再补生命周期字段，不要反过来从业务逻辑里回头拼装命令树。

推荐入口骨架：

```go
var App = cli.NewApp("example", "0.0.0")

func init() {
	cli.AddTo(App, &Serve{})
}
```

推荐命令骨架：

```go
type Serve struct {
	cli.C `name:"serve" component:"server" envprefix:"EXAMPLE_"`
}
```

如果这个命令还需要 HTTP、遥测或后台任务，再把对应 configurator 作为字段嵌进来，而不是在 `Run` 里手工创建它们。

## 推荐分层

在调用方项目中，建议按下面的职责切层：

- `cmd/<app>`
  只放 app 入口、命令注册、server 组装。
- `pkg/apis`
  放公开契约。
- `pkg/endpoints`
  放 endpoint 适配层。
- `domain/<name>`
  放业务逻辑和服务实现。

如果只有单个 app 自己使用的装配代码，不要因为它“看起来通用”就提前提到公共包。

## 什么时候参考示例

当你需要以下模式时，再参考 infra 提供的示例实现：

- `var App + func init()` 的入口组织方式。
- routes 生成和注册方式。
- `webapp.Server` 与普通 `http.Server` 的组合方式。
- domain / endpoints / apis 的分层。

示例是推荐装配方式，不是要求调用方复制整套目录树。

## 公开 API 核对入口

如果需要核对包职责、类型名和公开入口，优先使用 `go doc`：

- `go doc github.com/innoai-tech/infra/pkg/cli`
- `go doc github.com/innoai-tech/infra/pkg/configuration`
- `go doc github.com/innoai-tech/infra/pkg/http`
- `go doc github.com/innoai-tech/infra/pkg/http/webapp`
- `go doc github.com/innoai-tech/infra/pkg/otel`
- `go doc github.com/innoai-tech/infra/pkg/agent`

这些 `go doc` 信息应由各公共包的 `doc.go` 承载关键说明。对 pkg 用户重要的职责、边界、推荐入口和非目标，不应只写在仓库 `docs/` 里。

只有当 `go doc` 仍不足以解释某个示例装配模式时，才把示例实现当成补充材料，而不是主入口。
