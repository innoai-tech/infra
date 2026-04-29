---
name: infra-guideline
description: 说明 `github.com/innoai-tech/infra` pkg 中 `cli`、`configuration`、`http`、`otel`、`agent` 与示例实现的推荐用法、生命周期模型和公共边界；当任务涉及基于该 pkg 组装 app、命令、server、routes、singleton，或判断能力应留在调用方还是沉淀到 infra 公共包时使用。
metadata:
  primary_pattern: Tool Wrapper
  execution_trait: sequential
---

# Infra Guideline

用于按 `github.com/innoai-tech/infra` 的既有约定组装 app、command、server、routes 和公共基础包，避免调用方或贡献者偏离 `infra` 当前推荐模式。

## 目标

本 skill 只负责：

1. 解释 `infra` pkg 公共 API，以及其公共包与示例/内部实现的边界。
2. 说明 `cli -> configuration -> http -> otel` 的组合方式。
3. 给出可直接复用的用户手册，并在需要时补充 `go doc` 级别的公开 API 核对入口。

## 何时使用

- 需要基于 `github.com/innoai-tech/infra` 新增或调整 app / command / server 入口。
- 需要判断某个能力应留在调用方，还是应作为 infra 公共能力沉淀。
- 需要接入 `configuration` 生命周期、`http.Server`、`webapp.Server`、`otel.Otel`、`agent.Agent`。
- 需要参考 infra 推荐的示例分层和 routes 组装方式。

## 不适用

- 只是在既有约定内补一小段局部业务逻辑，不涉及入口、分层或公共边界。
- 只维护 infra 仓库控制面、skill 目录或发布配置。
- 只需要解释调用方自己的业务分层，且不涉及 infra pkg 的 API 或生命周期。

## 先看什么

1. 先看 [`references/repo-map.md`](references/repo-map.md) 了解这个 pkg 暴露哪些能力，以及一个 app 应该怎样分层。
2. 需要理解生命周期、命令注册和错误包装时，再看 [`references/runtime.md`](references/runtime.md)。
3. 需要判断代码该留在调用方还是沉淀到 infra 公共包时，再看 [`references/boundaries.md`](references/boundaries.md)。

## 工作方式

1. 先确认任务落点是在调用方 app、infra 公共 pkg，还是 infra 提供的示例/内部组装模式。
2. 涉及新入口时，优先复用 `var App + func init()` 方式注册命令。
3. 涉及 routes 时，优先沿用示例中的 `+gengo` 注册方式，而不是自行发明另一套入口组织。
4. 涉及生命周期时，优先复用 `configuration` 提供的接口组合，而不是自行拼装启动顺序。
5. 涉及公共能力沉淀时，先按边界说明判断是否真的该进入 infra 公共 pkg。
6. 优先根据本 skill 内的手册直接完成判断；只有需要核对公开 API、类型名或包职责时，才使用 `go doc` 查看对应 import path。
7. 涉及对外公开包的关键说明、使用边界或推荐入口时，优先更新对应包的 `doc.go`，不要只把信息留在 `docs/`。

## 完成标准

1. 改动遵循当前推荐分层和生命周期模型。
2. 没有把示例装配细节直接扩散到调用方公共 API 或 infra 公共 pkg。
3. 若任务涉及入口、路由或 lifecycle，已参照对应 reference 选择实现方式。
4. 若新增或修改了公共包的关键使用约定，`go doc` 可见层已同步反映，而不是只改 `docs/`。
