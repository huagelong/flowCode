# =============================================================================
# AnserFlow Agent Sandbox Dockerfile
# 目标：最小体积的 AI 代理代码执行沙箱（~50MB）
#
# 构建:  docker build -t anserflow/sandbox:latest -f docker/sandbox/Dockerfile .
# 运行:  docker run --rm -v $(pwd):/workspace anserflow/sandbox:latest
# =============================================================================

# ── Stage 1: 编译 anserflow CLI（Go 二进制） ─────────────────────────────────
FROM golang:1.24-alpine AS compiler

WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o anserflow ./cmd/anserflow

# ── Stage 2: 最小运行时 ─────────────────────────────────────────────────────
FROM alpine:3.21

LABEL org.opencontainers.image.title="AnserFlow Agent Sandbox"
LABEL org.opencontainers.image.description="Minimal sandbox for AI agent code execution"
LABEL org.opencontainers.image.source="https://github.com/anserflow/anserflow"

RUN apk add --no-cache \
    bash \
    curl \
    ca-certificates \
    git \
    git-lfs \
    openssh-client \
    && rm -rf /var/cache/apk/*

# ── 从 compiler 复制 anserflow CLI ──
COPY --from=compiler /build/anserflow /usr/local/bin/anserflow
RUN chmod +x /usr/local/bin/anserflow

# ── 安全加固 ──
RUN addgroup -g 1000 sandbox && \
    adduser -D -u 1000 -G sandbox sandbox

# ── 工作目录 ──
RUN mkdir -p /workspace /home/sandbox/.anseragent
RUN chown sandbox:sandbox /workspace /home/sandbox/.anseragent

WORKDIR /workspace

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

USER sandbox

# ── 环境变量 ──
ENV LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    TZ=Asia/Shanghai \
    GIT_CONFIG_GLOBAL=/tmp/gitconfig

# ── Git 初始化配置 ──
RUN git config --global --add safe.directory /workspace && \
    git config --global init.defaultBranch main

# ── 入口：anserflow agent run ──
ENTRYPOINT ["/entrypoint.sh"]
CMD ["--help"]
