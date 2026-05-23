> 来源：`docs/plan/11-backlog.md` 第 152 行
> 位置：[总目录](../../README.md) -> [AnserFlow — 远期 Backlog（Phase 2+）](../README.md) -> [工程基础设施（Phase 2）](README.md) -> Pact 合约测试
> 相邻：[上一篇](README.md) · [下一篇](02-golang-migrate-迁移工具.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### Pact 合约测试

前后端 API 契约一致性校验，当前用 TypeScript interface + Go struct 手工对齐。

| 项目 | 说明 |
|------|------|
| 触发时机 | 前后端 API 变更时 CI 自动运行 |
| 参考方案 | Pact Broker + `@pact-foundation/pact` (JS) + `pact-go` |
