> 来源：`docs/plan/11-backlog.md` 第 115 行
> 位置：[总目录](../../README.md) -> [AnserFlow — 远期 Backlog（Phase 2+）](../README.md) -> [任务计划](README.md) -> Todo ↔ Issue 双向同步
> 相邻：[上一篇](02-任务依赖图可视化.md) · [下一篇](04-执行策略引擎.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### Todo ↔ Issue 双向同步

| Todo 状态 | Issue 状态 | 同步方向 |
|-----------|-----------|----------|
| `[ ]` 未开始 | `backlog` | 创建时 Todo → Issue |
| 执行中 | `in_progress` | Issue → Todo（自动标记） |
| `[x]` 已完成 | `done` | 双向（任一侧完成即同步） |
| 验收退回 | `in_review` | Issue → Todo（取消勾选） |
