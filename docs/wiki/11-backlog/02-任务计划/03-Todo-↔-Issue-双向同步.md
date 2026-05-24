> 来源：`docs/plan/11-backlog.md` 第 115 行
> 位置：[总目录](../../README.md) -> [AnserFlow — 远期 Backlog（Phase 2+）](../README.md) -> [任务计划](README.md) -> Todo ↔ Issue 双向同步
> 相邻：[上一篇](02-任务依赖图可视化.md) · [下一篇](04-执行策略引擎.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### Issue 层级任务视图

Todo 不再作为独立实体；任务视图直接读取 `issues` 表：

| 视图概念 | Issue 表达 | 说明 |
|----------|------------|------|
| 需求 | `status=backlog` 且 `parent_id IS NULL` | 承载原始需求与讨论上下文 |
| 任务列表 | `status=todo` 且 `parent_id=<backlog_id>` | 同一 backlog 下的任务放在一起 |
| 执行中 | `status=in_progress` | 正在运行的任务 Issue |
| 审核中 | `status=in_review` | 唯一人工确认节点，此时展示 PR |
| 已完成 | `status=done` | PR 已完成并合并 |
