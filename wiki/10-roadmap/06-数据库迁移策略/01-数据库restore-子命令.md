> ???`docs/plan/10-roadmap.md` ? 87 ?
> ???[???](../../README.md) -> [AnserFlow — 开发路线图](../README.md) -> [数据库迁移策略](README.md) -> 数据库`restore` 子命令
> ???[???](README.md) ? [???](../07-种子数据.md)
> ?????[??????](README.md) ? [??????](../README.md)

### 数据库`restore` 子命令

除 `migrate --backup` 自动备份外，提供独立的 `restore` 子命令用于灾难恢复：

```bash
anserflow restore --file data/migrations/20260514120000_before.sql
  --config  config.yaml 路径
  --yes     跳过二次确认（默认 false）
  --dry-run 预览将执行的 SQL（默认 false）
```
