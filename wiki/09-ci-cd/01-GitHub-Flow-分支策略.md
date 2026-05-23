> ???`docs/plan/09-ci-cd.md` ? 3 ?
> ???[???](../README.md) -> [AnserFlow — CI/CD](README.md) -> GitHub Flow 分支策略
> ???[???](README.md) ? [???](02-GitHub-Actions-工作流总览.md)
> ?????[??????](README.md) ? [GitHub Actions 工作流总览](02-GitHub-Actions-工作流总览.md)

## GitHub Flow 分支策略

AnserFlow 采用 GitHub Flow，保持主干可部署、分支短生命周期：

```
main ─────────────────────────●──────────────────●────  (始终可部署)
      \                      /                  /
       feature/xxx ──●──●──●       fix/yyy ──●
```

| 规则 | 说明 |
|------|------|
| `main` 保护 | 禁止直接 push，由 Worker 在群聊审批通过后执行 squash merge |
| 功能分支 | `feat/issue-<id>`（AnserFlow 自动生成） / `feature/<描述>` / `fix/<描述>` |
| 审批方式 | 群聊审批：Agent 编码完成 → 发变更摘要到群聊 → 人工点「批准」→ 自动 merge |
| Commit 规范 | [Conventional Commits](https://www.conventionalcommits.org/zh-hans/)：`feat:` / `fix:` / `docs:` / `refactor:` / `ci:` |
| 发布标签 | `vX.Y.Z` 触发 CD 构建与发布 |
| 合并方式 | Squash & Merge（保持 main 线性历史） |

```bash
# 分支命名示例
git checkout -b feature/agent-orchestration
git checkout -b fix/issue-status-sync
git checkout -b docs/api-examples

# Commit 示例
feat: Agent 编排支持并行执行
docs: 补充 Docker 沙箱架构文档
fix: 修复 Issue 状态同步竞态条件
ci: 添加 Next.js lint 检查 workflow
```
