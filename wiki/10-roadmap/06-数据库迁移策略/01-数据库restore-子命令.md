> 来源：`docs/plan/10-roadmap.md` 第 87 行
> 位置：[总目录](../../README.md) -> [AnserFlow — 开发路线图](../README.md) -> [数据库迁移策略](README.md) -> 数据库`restore` 子命令
> 相邻：[上一篇](README.md) · [下一篇](../07-种子数据.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### 数据库`restore` 子命令

除 `migrate --backup` 自动备份外，提供独立的 `restore` 子命令用于灾难恢复：

```bash
anserflow restore --file data/migrations/20260514120000_before.sql
  --config  config.yaml 路径
  --yes     跳过二次确认（默认 false）
  --dry-run 预览将执行的 SQL（默认 false）
```
