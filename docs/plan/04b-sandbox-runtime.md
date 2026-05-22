# AnserFlow - 沙箱执行运行时

> **职责边界**：本文档覆盖 Docker 沙箱运行时适配器模式。Agent 基础设施层（Eino 框架、状态机、通知、Git、Token）见 [04-sandbox.md](04-sandbox.md)。
>
> **架构说明**：保留 RuntimeAdapter 和 OutputParser 接口抽象，方便未来扩展新的运行时。当前仅实现 anserAgent 一个运行时。

---

## 一、运行时适配器架构

### 设计原则

保留适配器模式以便未来扩展新的运行时（如 Claude Code、Codex CLI 等）。当前仅实现 anserAgent 一个运行时。

- `RuntimeAdapter` 接口定义配置注入、环境变量映射等运行时差异点
- `OutputParser` 接口定义 stdout 解析、Token 采集等输出差异点
- 新运行时只需实现接口并替换 RuntimeManager 初始化即可
- Worker 不感知具体运行时，只调用接口

### 接口定义

```go
// internal/runtime/adapter.go

package runtime

// RuntimeAdapter 运行时适配器 — 封装不同 AI 工具的配置注入差异
type RuntimeAdapter interface {
    // Name 运行时标识
    Name() string
    
    // HomeDir 运行时在容器内的主目录（bind mount 目标）
    HomeDir() string
    
    // ConfigPath 配置文件写入路径（容器内）
    ConfigPath() string
    
    // RenderConfig 将通用配置渲染为该工具特有的配置格式
    RenderConfig(config map[string]interface{}) (string, error)
    
    // EnvMapping API Key → 环境变量映射
    EnvMapping(config map[string]interface{}, decryptedKey string) map[string]string
    
    // SkillsMountPath Skills 目录在容器内的路径
    SkillsMountPath() string
    
    // SessionPath 会话文件路径（用于事后 Token 汇总），空字符串表示不支持
    SessionPath() string
    
    // ExecuteCommand 生成执行命令
    ExecuteCommand(workdir, configPath, prompt string) string
}

// OutputParser 输出解析器 — 封装不同 AI 工具的 stdout 格式差异
type OutputParser interface {
    // ParseLine 解析单行 stdout 输出
    // 返回 nil 表示该行不是结构化事件（按纯文本处理）
    ParseLine(line []byte) *ParsedEvent
    
    // ParseSessionFile 解析会话文件（执行完成后调用）
    ParseSessionFile(content []byte) (*TokenSummary, error)
}
```

### 当前实现：anserAgent

```go
// internal/runtime/anseragent.go

// AnserAgentAdapter anserAgent 运行时适配器实现
type AnserAgentAdapter struct{}

func (a *AnserAgentAdapter) Name() string {
    return "anseragent"
}

func (a *AnserAgentAdapter) HomeDir() string {
    return "/home/sandbox/.anseragent"
}

func (a *AnserAgentAdapter) ConfigPath() string {
    return "/home/sandbox/.anseragent/config.yaml"
}

func (a *AnserAgentAdapter) SkillsMountPath() string {
    return "/home/sandbox/.anseragent/skills"
}

func (a *AnserAgentAdapter) SessionPath() string {
    return "/home/sandbox/.anseragent/sessions/*.jsonl"
}

// ExecuteCommand 生成执行命令
func (a *AnserAgentAdapter) ExecuteCommand(workdir, configPath, prompt string) string {
    return fmt.Sprintf(
        "/usr/local/bin/anserflow agent run --workdir %s --config %s --prompt %q --format json",
        workdir, configPath, prompt,
    )
}
```

### RuntimeManager — 简化管理器

```go
// internal/runtime/manager.go

type RuntimeManager struct {
    adapter RuntimeAdapter
    parser  OutputParser
}

// NewRuntimeManager 创建运行时管理器（直接初始化 anserAgent）
func NewRuntimeManager() *RuntimeManager {
    return &RuntimeManager{
        adapter: &AnserAgentAdapter{},
        parser:  &AnserAgentParser{},
    }
}

// ResolveConfig 解析运行时配置
func (m *RuntimeManager) ResolveConfig(
    config map[string]interface{},
    decryptedKey string,
) (renderedConfig string, envVars map[string]string, err error) {
    renderedConfig, err = m.adapter.RenderConfig(config)
    if err != nil {
        return "", nil, err
    }
    
    envVars = m.adapter.EnvMapping(config, decryptedKey)
    return renderedConfig, envVars, nil
}
```

### Worker 调用示例（运行时无关）

```go
// internal/worker/executor.go

// 初始化
runtimeMgr := runtime.NewRuntimeManager()
adapter := runtimeMgr.GetAdapter()
parser := runtimeMgr.GetParser()

// 解析配置
renderedConfig, envVars, _ := runtimeMgr.ResolveConfig(
    agentConfig,
    decryptedAPIKey,
)

// 写入配置文件到容器
configPath := adapter.ConfigPath()
writeToContainer(containerID, configPath, renderedConfig)

// 构建执行命令
cmd := adapter.ExecuteCommand(
    issue.Workdir(),
    configPath,
    issue.TaskPrompt(),
)

// 设置环境变量
cmd.Env = buildEnv(envVars)

// 执行并解析 stdout
stdout, _ := cmd.StdoutPipe()
go parseAgentOutput(stdout, parser, issue.ID)

cmd.Run()
```

---

## 二、沙箱镜像与执行细节

> 以下内容从 [04-sandbox.md](04-sandbox.md) §二、Docker 沙箱方案中提取，为沙箱执行的实现级细节。

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

### Go Docker SDK

```go
// internal/sandbox/ — Docker SDK 核心操作

func ensureProjectContainer(ctx, project, runtime) (string, error)
//   复用已有容器 / 重启停止容器 / 创建新容器（1GB/2CPU/AutoRemove:false）
//   首次 git clone → /workspace/main（worktree 基准）

func createWorktree(ctx, containerID, issue) error
func removeWorktree(ctx, containerID, issueID) error
func execAgent(ctx, containerID, issue, task) error
func destroyProjectSandbox(ctx, projectID, containerID) error
```

### 运行时数据目录

```
/var/lib/anserflow/               ← runtime_data_dir
├── runtimes/                     ← Layer 1: 全局模板
│   └── anseragent/
│       ├── skills/
│       └── config.yaml
├── projects/                     ← Layer 2: 项目实例
│   └── 42/
│       └── runtime/
│           ├── skills/
│           └── config.yaml
沙箱容器                         ← Layer 3: bind mount
└── /home/sandbox/.anseragent/
    ├── skills/   ← bind mount 自 projects/42/runtime/
    └── config.yaml
```

```go
// 项目创建时从全局模板递归复制到项目实例目录，幂等
func initProjectRuntime(ctx, projectID, runtimeName) (string, error)
```

---

## 三、扩展新运行时指南

当需要支持新的 AI 编码工具时（如 Claude Code）：

### 步骤 1: 实现接口

```go
// internal/runtime/claude_code.go

type ClaudeCodeAdapter struct{}

func (c *ClaudeCodeAdapter) Name() string { return "claude_code" }
func (c *ClaudeCodeAdapter) HomeDir() string { return "/home/sandbox/.claude" }
// ... 实现其他接口方法

type ClaudeCodeParser struct{}
// ... 实现 OutputParser 接口
```

### 步骤 2: 替换初始化

```go
// internal/runtime/manager.go

func NewRuntimeManager() *RuntimeManager {
    return &RuntimeManager{
        adapter: &ClaudeCodeAdapter{},  // 替换为新实现
        parser:  &ClaudeCodeParser{},
    }
}
```

### 步骤 3: 更新 Dockerfile

```dockerfile
# 添加新运行时的安装步骤
RUN curl -fsSL https://claude.ai/install.sh | sh
```

Worker 代码无需修改，完全通过接口隔离。

