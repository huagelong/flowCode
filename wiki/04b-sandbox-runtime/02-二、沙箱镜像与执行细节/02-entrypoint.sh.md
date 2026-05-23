> ???`docs/plan/04b-sandbox-runtime.md` ? 220 ?
> ???[???](../../README.md) -> [AnserFlow - 沙箱执行运行时](../README.md) -> [二、沙箱镜像与执行细节](README.md) -> entrypoint.sh
> ???[???](01-Dockerfile.md) ? [???](03-Go-Docker-SDK.md)
> ?????[??????](README.md) ? [??????](../README.md)

### entrypoint.sh

```bash
#!/bin/bash
set -e

# Git 仓库初始化
if [ -n "$GIT_REPO_URL" ]; then
    echo "📦 Cloning repository..."
    git clone --branch "${GIT_BRANCH:-main}" "$GIT_REPO_URL" /workspace/main
fi

# 执行编码任务
if [ -n "$TASK_PROMPT" ]; then
    echo "🤖 Starting anserAgent..."
    
    /usr/local/bin/anserflow agent run \
        --workdir "${WORKDIR:-/workspace/main}" \
        --config /home/sandbox/.anseragent/config.yaml \
        --prompt "$TASK_PROMPT" \
        --format json
    
    echo "✅ anserAgent completed"
    exit $?
fi

# 保持容器运行
exec tail -f /dev/null
```

> API Key AES-256 加密存储，Worker 解密后通过环境变量注入容器。
