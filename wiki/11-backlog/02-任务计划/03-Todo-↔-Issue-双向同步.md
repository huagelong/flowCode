> ???`docs/plan/11-backlog.md` ? 115 ?
> ???[???](../../README.md) -> [AnserFlow — 远期 Backlog（Phase 2+）](../README.md) -> [任务计划](README.md) -> Todo ↔ Issue 双向同步
> ???[???](02-任务依赖图可视化.md) ? [???](04-执行策略引擎.md)
> ?????[??????](README.md) ? [??????](../README.md)

### Todo ↔ Issue 双向同步

| Todo 状态 | Issue 状态 | 同步方向 |
|-----------|-----------|----------|
| `[ ]` 未开始 | `backlog` | 创建时 Todo → Issue |
| 执行中 | `in_progress` | Issue → Todo（自动标记） |
| `[x]` 已完成 | `done` | 双向（任一侧完成即同步） |
| 验收退回 | `in_review` | Issue → Todo（取消勾选） |
