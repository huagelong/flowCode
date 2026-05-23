> ???`docs/plan/11-backlog.md` ? 152 ?
> ???[???](../../README.md) -> [AnserFlow — 远期 Backlog（Phase 2+）](../README.md) -> [工程基础设施（Phase 2）](README.md) -> Pact 合约测试
> ???[???](README.md) ? [???](02-golang-migrate-迁移工具.md)
> ?????[??????](README.md) ? [??????](../README.md)

### Pact 合约测试

前后端 API 契约一致性校验，当前用 TypeScript interface + Go struct 手工对齐。

| 项目 | 说明 |
|------|------|
| 触发时机 | 前后端 API 变更时 CI 自动运行 |
| 参考方案 | Pact Broker + `@pact-foundation/pact` (JS) + `pact-go` |
