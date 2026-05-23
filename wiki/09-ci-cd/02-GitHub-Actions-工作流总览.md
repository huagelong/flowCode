> ???`docs/plan/09-ci-cd.md` ? 35 ?
> ???[???](../README.md) -> [AnserFlow — CI/CD](README.md) -> GitHub Actions 工作流总览
> ???[???](01-GitHub-Flow-分支策略.md) ? [???](03-ci.yml-—-Pull-Request-检查.md)
> ?????[??????](README.md) ? [GitHub Flow 分支策略](01-GitHub-Flow-分支策略.md) ? [ci.yml — Pull Request 检查](03-ci.yml-—-Pull-Request-检查.md)

## GitHub Actions 工作流总览

```
┌──────────────────────────────────────────────────────────────┐
│  分支 push → main                                             │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  ci.yml (push 分支 / push main)                       │    │
│  │  ├── Go lint + test + build                          │    │
│  │  ├── Next.js lint + type-check + build (admin)       │    │
│  │  └── Next.js lint + type-check + build (client)      │    │
│  └──────────────────────────────────────────────────────┘    │
│                           ↓ 群聊审批 → squash merge            │
├──────────────────────────────────────────────────────────────┤
│  main → 发布                                                  │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  sandbox-image.yml (push main / Dockerfile 变更)      │    │
│  │  ├── Build sandbox Docker image                      │    │
│  │  ├── Push to ghcr.io/anserflow/sandbox               │    │
│  │  └── Tag: latest + commit-sha                        │    │
│  ├──────────────────────────────────────────────────────┤    │
│  │  go-release.yml (push tag v*)                        │    │
│  │  ├── Cross-compile Go backend                        │    │
│  │  ├── Upload anserflow binary (linux/windows/macos)   │    │
│  │  └── Create GitHub Release                           │    │
│  └──────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────┘
```

| 工作流 | 触发条件 | 耗时 | 产物 |
|--------|---------|------|------|
| `ci.yml` | PR / push main | ~3min | 无（仅检查） |
| `sandbox-image.yml` | push main (Dockerfile) | ~5min | `ghcr.io/anserflow/sandbox` |
| `go-release.yml` | tag `v*` | ~8min | 多平台二进制 + Release |
