> 来源：`docs/plan/09-ci-cd.md` 第 177 行
> 位置：[总目录](../README.md) -> [AnserFlow — CI/CD](README.md) -> go-release.yml — Go 后端发布
> 相邻：[上一篇](04-sandbox-image.yml-—-Docker-沙箱镜像.md) · [下一篇](06-GitHub-Secrets-清单.md)
> 相关主题：[返回文档入口](README.md) · [sandbox-image.yml — Docker 沙箱镜像](04-sandbox-image.yml-—-Docker-沙箱镜像.md) · [GitHub Secrets 清单](06-GitHub-Secrets-清单.md)

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
          body: 'Release artifacts for this tag.'
          files: |
            */anserflow*
            checksums.txt
          generate_release_notes: true
```
