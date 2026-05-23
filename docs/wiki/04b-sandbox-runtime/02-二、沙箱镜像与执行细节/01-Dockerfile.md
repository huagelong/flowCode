> 来源：`docs/plan/04b-sandbox-runtime.md` 第 178 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱执行运行时](../README.md) -> [二、沙箱镜像与执行细节](README.md) -> Dockerfile
> 相邻：[上一篇](README.md) · [下一篇](02-entrypoint.sh.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### Dockerfile

位于 `docker/sandbox/Dockerfile`：

```bash
docker build -t anserflow/sandbox:latest -f docker/sandbox/Dockerfile .
```

```dockerfile
FROM alpine:3.21 AS builder

# 编译阶段（如果需要）
FROM golang:1.24-alpine AS compiler
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o anserflow ./cmd/anserflow

# 最终镜像
FROM alpine:3.21

RUN apk add --no-cache git bash ca-certificates

RUN adduser -D -u 1000 sandbox

# 直接复制编译好的二进制
COPY --from=compiler /build/anserflow /usr/local/bin/anserflow
RUN chmod +x /usr/local/bin/anserflow

RUN mkdir -p /workspace /home/sandbox/.anseragent
RUN chown sandbox:sandbox /workspace /home/sandbox/.anseragent

WORKDIR /workspace
USER sandbox

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
```

> 预估镜像大小约 50MB（移除 Node.js/npm/Python 后）。`.dockerignore` 排除 node_modules/ / .git/ / dist/ / .next/ / *.log。
