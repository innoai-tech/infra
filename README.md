# Infra

`infra` 是一组面向 Go 应用的基础设施构件，重点放在 CLI 入口组织、配置注入、HTTP 服务承载、agent 生命周期与 OpenTelemetry 集成。

这个仓库同时提供库能力和一个可运行示例，用来验证这些基础构件如何组合成实际服务。

更多背景文档见 [docs/app-guide.md](./docs/app-guide.md)、[docs/core.md](./docs/core.md)、[docs/boundaries.md](./docs/boundaries.md)、[docs/mechanisms.md](./docs/mechanisms.md)、[docs/api-review.md](./docs/api-review.md)、[docs/error-style.md](./docs/error-style.md)。

## Responsibilities

仓库负责：

- 提供 `pkg/` 下可复用的基础能力，例如 `cli`、`configuration`、`http`、`otel`、`agent`。
- 提供 `internal/example` 作为示例应用与集成入口。
- 提供 `tool/internal/cmd` 作为仓库内生成、格式化与校验工具入口。
- 提供 `cuepkg/` 下的组件输出样例与相关工件。

仓库当前不负责：

- 在 root README 里维护详细实现原理或长期设计推导。
- 在 README 中承载协作约束或命令细节，这些分别下沉到 `AGENTS.md` 和 `justfile`。

## Boundary

- root `pkg/*` 承载可复用基础能力，目标是形成相对稳定的库级 API。
- `internal/*` 承载仓库内实现、示例应用和不承诺稳定性的组装代码。
- `internal/example` 是第一参考实现，用来表达推荐的应用分层，不视为 root `pkg/*` 的公共边界。

## Start Here

- 运行 `just` 查看仓库暴露的稳定入口。
- 运行 `just --list --list-submodules` 查看 root 与 toolchain 的聚合入口。
- 运行 `just go::gen-root`、`just go::gen-example` 或 `just example::gen` 使用约定好的生成入口。
- 阅读 [internal/example](./internal/example) 了解一个完整示例应用如何组装这些基础设施。
- 阅读 [pkg/cli](./pkg/cli)、[pkg/http](./pkg/http)、[pkg/otel](./pkg/otel)、[pkg/configuration](./pkg/configuration)
  查看核心库边界。

## Navigation

- [AGENTS.md](./AGENTS.md): root 协作约束与暂停门禁。
- [.agents/skills](./.agents/skills): 仓库内提供给 agent 使用的本地技能目录，包含 `infra-guideline` 等约定说明。
- [justfile](./justfile): root 级统一执行入口，聚合 toolchain 与示例运行命令。
- [tool/go/justfile](./tool/go/justfile): Go toolchain 执行面，承载依赖、测试、格式化与生成入口。
- [internal/example](./internal/example): 示例应用入口，覆盖 `cmd`、`pkg/apis`、`pkg/endpoints` 与 `domain` 的完整分层。
- [pkg](./pkg): 对外可复用的基础库。
- [cuepkg](./cuepkg): 组件与生成产物相关内容。
- [docs/app-guide.md](./docs/app-guide.md): 从零组装一个 app / command / server 的最小路径。
- [docs/core.md](./docs/core.md): 运行模型与生命周期时序。
- [docs/boundaries.md](./docs/boundaries.md): `pkg/*` 与 `internal/*` 的边界说明。
- [docs/mechanisms.md](./docs/mechanisms.md): 反射、生成与目录约定的交叉点说明。
- [docs/api-review.md](./docs/api-review.md): 当前公共 API 复盘与后续收敛方向。
- [docs/error-style.md](./docs/error-style.md): 初始化、运行与关闭阶段的错误表达约定。
