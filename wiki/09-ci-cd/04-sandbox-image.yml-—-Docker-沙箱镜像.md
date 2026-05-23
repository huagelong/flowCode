> ???`docs/plan/09-ci-cd.md` ? 135 ?
> ???[???](../README.md) -> [AnserFlow — CI/CD](README.md) -> sandbox-image.yml — Docker 沙箱镜像
> ???[???](03-ci.yml-—-Pull-Request-检查.md) ? [???](05-go-release.yml-—-Go-后端发布.md)
> ?????[??????](README.md) ? [ci.yml — Pull Request 检查](03-ci.yml-—-Pull-Request-检查.md) ? [go-release.yml — Go 后端发布](05-go-release.yml-—-Go-后端发布.md)

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
