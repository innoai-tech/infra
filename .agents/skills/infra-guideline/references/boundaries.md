# Boundaries

这份文档回答一个核心问题：代码应该留在调用方项目，还是应该沉淀到 `github.com/innoai-tech/infra` 的公共包。

## 先做的判断

看到一段新能力时，先问三个问题：

1. 它是公共基础设施能力，还是某个 app 的装配代码？
2. 其它 app 现在是否真的会复用，而不是“以后也许会”？
3. 如果公开出去，后续演进时是否愿意承担兼容成本？

只要有一个答案偏向否，就先留在调用方项目。

## 应该留在调用方的内容

下面这些内容默认不应进入 infra 公共包：

- 业务域模型。
- 只服务某个 app 的 routes 组装。
- 单个项目特有的协议、命名、目录约定。
- 一次性 glue code。
- 依赖某个业务背景才能理解的 helper。

这类代码可以放在调用方自己的 `internal/*`、`domain/*`、`pkg/*` 中，由项目自己承担演进成本。

## 可以考虑沉淀到 infra 公共包的内容

只有同时满足下面条件，才考虑进入 infra 公共包：

- 跨多个 app 复用。
- API 形态和命名已经稳定。
- 不依赖某个具体业务域。
- 调用方拿来即用，不需要复制示例目录结构。

典型例子是：

- 通用 CLI 入口组织能力。
- 生命周期编排能力。
- 通用 HTTP / webapp / otel / agent 组件。

## 关于 `pkg/*` 和 `internal/*`

在 infra 的设计分层里：

- `pkg/*` 表示稳定、可复用、愿意公开承诺的能力。
- `internal/*` 表示示例实现、内部组装和仓库内部适配逻辑。

但对于调用方 agent，真正重要的判断不是目录名，而是“是否应该成为稳定公共 API”。

## 常见误判

### 误判 1：示例里有，所以公共包里也应该有

不成立。示例的职责是展示推荐装配方式，不是扩大公共 API 面积。

### 误判 2：放进 `pkg/*` 以后更容易复用

这常常只是把调用方自己的演进成本转嫁成公共兼容成本。只有已经形成稳定复用需求时，才值得上提。

### 误判 3：子包存在就等于公开入口

不成立。公共入口取决于是否被推荐为稳定 API，而不是路径是否位于 `pkg/` 之下。

## 公开 API 核对入口

如果需要确认某个能力是否已经被公开承诺，优先看对应 root package 的 `go doc` 是否把它当成直接使用入口：

- `go doc github.com/innoai-tech/infra/pkg/cli`
- `go doc github.com/innoai-tech/infra/pkg/configuration`
- `go doc github.com/innoai-tech/infra/pkg/http`
- `go doc github.com/innoai-tech/infra/pkg/http/webapp`
- `go doc github.com/innoai-tech/infra/pkg/otel`
- `go doc github.com/innoai-tech/infra/pkg/agent`

因此，涉及公共边界的关键说明应优先进入相应包的 `doc.go`。`docs/` 可以补充设计背景，但不应成为 pkg 用户理解公开边界的唯一入口。

判断标准不是“是否存在某个实现子包”，而是“公开包文档是否把它表达成稳定入口”。
