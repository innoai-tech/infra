# Package Boundaries

`infra` 同时维护公共基础库和仓库内示例。为了避免示例装配细节反向污染公共 API，目录边界需要明确。

## `pkg/*`

放稳定、可复用、希望对外承诺的基础能力，例如：

- `pkg/cli`
- `pkg/configuration`
- `pkg/http`
- `pkg/otel`
- `pkg/agent`

这些 root package 可以进一步分成两类：

- 稳定公共 API：
  `pkg/cli`、`pkg/configuration`、`pkg/http`、`pkg/otel`、`pkg/agent`、`pkg/appinfo`、`pkg/cron`
- 仅供公共包内部装配或能力细分的子包：
  例如 `pkg/http/middleware`、`pkg/http/webapp/appconfig`、`pkg/otel/metric`

判断原则不是“路径在 pkg 下就都等价公开”，而是看它是否作为 root package 的直接入口被推荐使用。

进入 `pkg/*` 的前提：

- 具备跨应用复用价值。
- 公开 API 和命名已经稳定。
- 不依赖示例 app 的目录、协议或业务概念。

## `internal/*`

放仓库内部实现和推荐组装方式，例如：

- `internal/example`
- `internal/cmd/...`
- 只服务于当前仓库生成/开发流程的适配逻辑

这里的代码可以更快演进，不承诺稳定导入路径。

## Decision Rule

遇到新代码时，先问三个问题：

1. 这是库能力，还是某个 app 的装配代码？
2. 其它 app 是否真的会复用，而不是“可能以后会”？
3. 暴露出去后，未来变更是否愿意承担兼容成本？

只要有一个答案偏向否，优先放在 `internal/*`。
