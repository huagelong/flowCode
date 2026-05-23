> ???`docs/plan/11-backlog.md` ? 161 ?
> ???[???](../../README.md) -> [AnserFlow — 远期 Backlog（Phase 2+）](../README.md) -> [工程基础设施（Phase 2）](README.md) -> golang-migrate 迁移工具
> ???[???](01-Pact-合约测试.md) ? [???](03-Crowdin-Lokalise-翻译管理.md)
> ?????[??????](README.md) ? [??????](../README.md)

### golang-migrate 迁移工具

当前 L1-L4 采用 `AutoMigrate + 备份 SQL + 种子数据`。Phase 2 切换为 `golang-migrate/migrate` 以支持版本化迁移和回滚。
