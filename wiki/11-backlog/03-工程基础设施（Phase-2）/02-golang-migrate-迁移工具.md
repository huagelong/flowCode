> 来源：`docs/plan/11-backlog.md` 第 161 行
> 位置：[总目录](../../README.md) -> [AnserFlow — 远期 Backlog（Phase 2+）](../README.md) -> [工程基础设施（Phase 2）](README.md) -> golang-migrate 迁移工具
> 相邻：[上一篇](01-Pact-合约测试.md) · [下一篇](03-Crowdin-Lokalise-翻译管理.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### golang-migrate 迁移工具

当前 L1-L4 采用 `AutoMigrate + 备份 SQL + 种子数据`。Phase 2 切换为 `golang-migrate/migrate` 以支持版本化迁移和回滚。
