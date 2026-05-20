# AnserFlow — CI/CD

## GitHub Flow 分支策略

AnserFlow 采用 GitHub Flow，保持主干可部署、分支短生命周期：

```
main ─────────────────────────●──────────────────●────  (始终可部署)
      \                      /                  /
       feature/xxx ──●──●──●       fix/yyy ──●
```

| 规则 | 说明 |
|------|------|
| `main` 保护 | 禁止直接 push，必须通过 PR 合并 |
| 功能分支 | `feature/<描述>` / `fix/<描述>` / `docs/<描述>` |
| PR 要求 | 至少 1 人 Review + CI 全绿 |
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

## GitHub Actions 工作流总览

```
┌──────────────────────────────────────────────────────────────┐
│  PR → main                                                    │
│  ┌──────────────────────────────────────────────────────┐    │
│  │  ci.yml (每次 push PR)                                │    │
│  │  ├── Go lint + test + build                          │    │
│  │  ├── Next.js lint + type-check + build (admin)       │    │
│  │  └── Next.js lint + type-check + build (client)      │    │
│  └──────────────────────────────────────────────────────┘    │
│                           ↓ 合并                              │
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

## ci.yml — Pull Request 检查

```yaml
# .github/workflows/ci.yml
name: CI
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  # ── Go 后端 ──
  go:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true
      - name: Lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          args: --timeout=3m
      - name: Test
        run: go test -race -coverprofile=coverage.out ./...
      - name: Build
        run: go build -o /dev/null ./...

  # ── Next.js admin ──
  admin:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: admin
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
          cache: 'npm'
      - run: npm ci
      - run: npm run lint
      - run: npm run type-check
      - run: npm run build

  # ── Next.js client ──
  client:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: client
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
          cache: 'npm'
      - run: npm ci
      - run: npm run lint
      - run: npm run type-check
      - run: npm run build
```

## sandbox-image.yml — Docker 沙箱镜像

```yaml
# .github/workflows/sandbox-image.yml
name: Sandbox Image
on:
  push:
    branches: [main]
    paths:
      - 'docker/sandbox/**'
      - '.github/workflows/sandbox-image.yml'
  workflow_dispatch:

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v4
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      - name: Login to GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - name: Build and push
        uses: docker/build-push-action@v6
        with:
          context: .
          file: docker/sandbox/Dockerfile
          push: true
          tags: |
            ghcr.io/${{ github.repository }}/sandbox:latest
            ghcr.io/${{ github.repository }}/sandbox:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

## go-release.yml — Go 后端发布

```yaml
# .github/workflows/go-release.yml
name: Go Release
on:
  push:
    tags: ['v*']

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - goos: linux
            goarch: amd64
          - goos: linux
            goarch: arm64
          - goos: windows
            goarch: amd64
            ext: .exe
          - goos: darwin
            goarch: amd64
          - goos: darwin
            goarch: arm64
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: '22', cache: 'npm' }
      - name: Build admin SPA
        run: |
          npm ci
          npm run build -w @anserflow/admin
      - uses: actions/setup-go@v5
        with: { go-version: '1.24' }
      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
          CGO_ENABLED: 0
        run: |
          go build -ldflags="-s -w -X main.Version=${GITHUB_REF_NAME}" \
            -o anserflow${{ matrix.ext }} .
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: anserflow-${{ matrix.goos }}-${{ matrix.goarch }}
          path: anserflow${{ matrix.ext }}

  release:
    needs: build
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/download-artifact@v4
      - name: Generate checksums
        run: |
          find . -type f -name 'anserflow*' -exec sha256sum {} \; > checksums.txt
      - name: Create Release
        uses: softprops/action-gh-release@v2
        with:
          name: 'AnserFlow ${{ github.ref_name }}'
          body: 'See [CHANGELOG.md](./CHANGELOG.md)'
          files: |
            */anserflow*
            checksums.txt
          generate_release_notes: true
```

## GitHub Secrets 清单

| Secret | 用途 | 触发工作流 |
|--------|------|----------|
| `GITHUB_TOKEN` | 自动提供，无需手动配置 | 所有 |

## 分支保护规则

| 规则 | 值 |
|------|-----|
| Require a pull request before merging | ✅ |
| Require approvals | 1 |
| Require status checks to pass | ✅ `ci.yml` (go / admin / client) |
| Require conversation resolution | ✅ |
| Do not allow bypassing | ✅ (包括 admins) |


