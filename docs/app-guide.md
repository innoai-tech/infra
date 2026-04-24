# Build an App

这份文档给出一个从零开始组装 `app / command / server` 的最小路径，目标是让调用链清楚，而不是罗列所有能力。

## 1. 定义入口

入口文件推荐只保留全局 `App` 和 `init` 注册：

```go
var App = cli.NewApp("example", "0.0.0")

func init() {
	cli.AddTo(App, &Serve{})
}
```

这样可以把命令树的注册关系固定在入口面，避免在运行时散落拼装。

## 2. 声明命令

命令类型嵌入 `cli.C`，用 tag 声明名字、组件、环境变量前缀等元信息：

```go
type Serve struct {
	cli.C `name:"serve" component:"server" envprefix:"EXAMPLE_"`
}
```

`cli` 会负责：

- 从 struct tag 收集 args / flags。
- 按应用名、命令路径和 `envprefix` 推导环境变量。
- 在执行前串起 `configuration` 生命周期。

## 3. 组合 singleton / configurator

在命令 struct 上直接声明需要参与生命周期的字段：

```go
type Serve struct {
	cli.C `component:"server"`

	appinfo.Info
	otel.Otel
	webapp.Server
}
```

这些字段如果实现了 `SetDefaults`、`Init`、`InjectContext`、`Run`、`Serve`、`Shutdown` 中的任一接口，就会被 `configuration.SingletonsFromStruct` 提取出来。

## 4. 把业务逻辑放到明确层次

推荐分层：

- `cmd/<app>` 或 `internal/example/cmd/example`
  只放入口、routes、server 组装。
- `pkg/apis`
  放契约定义。
- `pkg/endpoints`
  放 endpoint 适配层。
- `domain/<name>`
  放业务逻辑与服务实现。

不要把 courier 契约、HTTP routes 和 domain service 混在同一个包里。

## 5. 运行和验证

- `just` 查看 root 暴露入口。
- `just go::test` 或 `go test ./...` 运行测试。
- `just go::gen` 运行生成。
- `just go::cover-core` 检查公共核心包覆盖率。

## 6. 什么时候放进 `pkg/*`

只有同时满足下面条件，才适合进入 root `pkg/*`：

- 这个能力可以被多个 app 复用。
- API 形态足够稳定，后续不希望随示例演进频繁调整。
- 不依赖某个单独业务域的命名和目录结构。

否则应留在 `internal/*`。
