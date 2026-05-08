---
name: infra-guideline
description: 说明 `github.com/innoai-tech/infra` pkg 中 `cli`、`configuration`、`http`、`otel`、`agent` 与示例实现的推荐用法、生命周期模型和公共边界；当任务涉及基于该 pkg 组装 app、命令、server、routes 或判断能力应留在调用方还是沉淀到 infra 公共包时使用。
---

# Infra Guideline

用于按 `github.com/innoai-tech/infra` 的约定组装 app、command、server、routes 和公共基础包。

## 架构总览

```
cli ──→ configuration ──→ http / agent
入口      生命周期编排       运行时承载
                              │
                              ├─ http.Server
                              ├─ webapp.Server
                              ├─ otel.Otel
                              └─ agent.Agent
```

1. **cli** — app / command 入口，flag 与 env 绑定。起手先固定入口。
2. **configuration** — 生命周期编排，按 `SetDefaults → Init → InjectContext → Run → Serve → Shutdown` 串联。
3. **http / agent** — 运行时承载。HTTP 服务、WebApp、遥测、后台任务。

## 使用范围

- 基于 infra 新增或调整 app / command / server 入口。
- 判断某个能力应留在调用方还是沉淀到 infra 公共包。
- 接入 `configuration` 生命周期、`http.Server`、`webapp.Server`、`otel.Otel`、`agent.Agent`。

不适用：只补局部业务逻辑且不涉及入口分层，或只维护 infra 仓库控制面/skill。

## 读取导航

- 了解 infra 提供哪些能力、如何起手 → [references/repo-map.md](references/repo-map.md)
- 理解生命周期顺序和组合方式 → [references/runtime.md](references/runtime.md)
- 判断代码该留在调用方还是沉淀到 infra → [references/boundaries.md](references/boundaries.md)

需要核对 API 签名时，优先 `go doc` 对应包，不在 skill 中复制手册。

## 完成标准

- 改动遵循当前推荐分层和生命周期模型。
- 没有把示例装配细节扩散到调用方公共 API 或 infra 公共包。
- 若新增或修改了公共包的关键使用约定，`go doc` 可见层已同步反映。
