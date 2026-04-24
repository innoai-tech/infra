# Repo Map

本仓库的稳定入口按下面顺序阅读：

1. [README.md](../../../README.md)
   先了解仓库职责、边界和主导航。
2. [AGENTS.md](../../../AGENTS.md)
   再确认仓库级协作约束、暂停条件和变更边界。
3. [justfile](../../../justfile) 与 [tool/go/justfile](../../../tool/go/justfile)
   查看稳定执行入口、覆盖率检查、生成入口和文档检查入口。
4. [internal/example/README.md](../../../internal/example/README.md)
   了解仓库内第一参考实现的目录分层。

推荐阅读顺序：

1. 若任务是理解仓库，先看 root `README.md`。
2. 若任务是写代码，先看 `docs/app-guide.md` 和 `internal/example`。
3. 若任务涉及控制面或执行入口，再回到 `AGENTS.md` 和 `justfile`。
