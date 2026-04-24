# Reflection, Generation, and Convention

这个仓库的运行体验来自三套机制的叠加：

- 反射：
  `cli` 和 `configuration` 通过 struct 遍历提取 flags、args、singletons。
- 生成：
  `gengo` 负责 injectable、runtime doc、operator registration 等样板代码。
- 约定：
  入口采用 `var App + func init()`，routes 采用 `+gengo` 注册，目录采用 `cmd / pkg / domain` 分层。

如果这三套机制的交叉点不明确，调用方就必须同时理解源码、生成结果和目录约定，维护成本会快速上升。

## Current Cross Points

### CLI + Reflection

- `cli.C` 不直接承载所有业务字段。
- 只有被识别为 singleton / configurator 的嵌入字段，才会继续被反射收集 flags。
- 这意味着“字段放在命令 struct 上”和“字段放在嵌入 configurator 上”语义不同。

### Configuration + Reflection

- `configuration.SingletonsFromStruct` 只提取导出字段。
- 匿名嵌入字段如果未导出，会被直接跳过。
- 一个字段即使是 struct，也只有实现了生命周期/注入接口后才会被纳入运行面。

### Gengo + Routes

- routes 层不是手写注册，而是通过 `+gengo:operator:register=R` 生成。
- 因此目录结构和生成入口必须稳定，否则运行时路由树会和源码定义脱节。

## Recommended Reading Order

1. 先看 `docs/app-guide.md` 了解推荐目录和入口写法。
2. 再看 `docs/core.md` 理解生命周期时序。
3. 最后在需要时查看生成文件，确认某个 injectable / operator 是如何落到运行时的。

## Practical Rule

当一个行为看起来“不像源码里写的那样”时，先检查：

1. 这个字段是否真的会被反射扫到。
2. 这个类型是否实现了被框架识别的接口。
3. 是否有对应的生成文件在补全真实运行逻辑。

这三步能覆盖当前仓库里大多数“为什么这里会生效 / 为什么这里没生效”的问题。
