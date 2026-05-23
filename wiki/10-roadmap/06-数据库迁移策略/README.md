> 来源：`docs/plan/10-roadmap.md` 第 68 行
> 位置：[总目录](../../README.md) -> [AnserFlow — 开发路线图](../README.md) -> 数据库迁移策略
> 相邻：[上一篇](../05-测试策略.md) · [下一篇](../07-种子数据.md)
> 相关主题：[返回文档入口](../README.md) · [测试策略](../05-测试策略.md) · [种子数据](../07-种子数据.md)

## 数据库迁移策略

GORM AutoMigrate 仅处理正向迁移（创建表/添加列）。需要回滚时采用：

| 场景 | 方案 |
|------|------|
| **本地开发** | `anserflow migrate --dry-run` 预览 SQL → 手动执行回滚 DDL |
| **生产发布** | 每次 `anserflow migrate` 前自动生成备份 SQL（`data/migrations/YYYYMMDDHHMMSS_before.sql`） |
| **紧急回滚** | 执行对应时间的备份 SQL 恢复表结构 |
| **当前收口** | 本轮统一采用 `AutoMigrate + 备份 SQL + 种子数据`；`golang-migrate/migrate` 放入 [Phase 2](../../11-backlog/README.md) 独立任务 |

```bash
# 迁移前自动备份
anserflow migrate --backup    # → data/migrations/20260514120000_before.sql

# 跳过种子数据（仅 DDL）
anserflow migrate --seed=false
```

## 子章节导航

- [数据库`restore` 子命令](01-数据库restore-子命令.md)
