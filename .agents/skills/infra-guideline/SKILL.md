---
name: infra-guideline
description: 说明本仓库中 `cli`、`configuration`、`http`、`otel`、`agent` 与 `internal/example` 的推荐用法、生命周期模型和公共边界；当任务涉及在本仓库内新增 app、命令、server、routes、singleton 组合，或需要判断应放在 `pkg/*` 还是 `internal/*` 时使用。
metadata:
  primary_pattern: Tool Wrapper
  execution_trait: sequential
---

# Infra Guideline

用于在本仓库内按既有约定组装 app、command、server、routes 和公共基础包，避免偏离 `infra` 当前推荐模式。

## 目标

本 skill 只负责：

1. 解释本仓库 `pkg/*` 与 `internal/*` 的边界。
2. 说明 `cli -> configuration -> http -> otel` 的组合方式。
3. 指向当前推荐示例和控制面入口。

## 何时使用

- 需要在本仓库内新增或调整 app / command / server 入口。
- 需要判断某个能力应放在 root `pkg/*` 还是 `internal/*`。
- 需要接入 `configuration` 生命周期、`http.Server`、`webapp.Server`、`otel.Otel`、`agent.Agent`。
- 需要参考 `internal/example` 的推荐分层。

## 不适用

- 只是在既有约定内补一小段局部业务逻辑，不涉及入口、分层或公共边界。
- 只做仓库控制面维护，应改用 `project-control-tidy`。
- 只维护某个 skill 目录，应改用 `skill-tidy`。

## 先看什么

1. 先看 [`references/repo-map.md`](references/repo-map.md) 了解仓库入口与推荐阅读顺序。
2. 需要理解生命周期时，再看 [`references/runtime.md`](references/runtime.md)。
3. 需要判断边界、公开 API 或隐式约束时，再看 [`references/boundaries.md`](references/boundaries.md)。

## 工作方式

1. 先确认任务落点是在 root `pkg/*`、`internal/example`，还是其他 `internal/*` 组装目录。
2. 涉及新入口时，优先复用 `var App + func init()` 方式注册命令。
3. 涉及 routes 时，优先沿用 `internal/example/cmd/example/routes` 的 `+gengo` 注册方式。
4. 涉及生命周期时，优先复用 `configuration` 提供的接口组合，而不是自行拼装启动顺序。
5. 涉及公共能力沉淀时，先按边界说明判断是否真的该进入 root `pkg/*`。

## 完成标准

1. 改动遵循当前推荐分层和生命周期模型。
2. 没有把示例装配细节直接扩散到不该公开的 root package。
3. 若任务涉及入口、路由或 lifecycle，已参照对应 reference 选择实现方式。
