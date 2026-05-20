# AnserFlow - 沙箱执行运行时

> **职责边界**：本文档覆盖 Docker 沙箱生命周期管理和运行时适配器模式。Agent 基础设施层（Eino 框架、状态机、通知、Git、Token）见 [04-sandbox.md](04-sandbox.md)。

---

### SandboxManager — 沙箱生命周期管理器

沙箱创建、暂停、恢复、销毁、复用的逻辑分散在 Worker 和 executor.go 中，通过 `SandboxManager` 统一 Docker API 调用。

**设计原则**：

- Worker 只调 `SandboxManager.Create/Resume/Destroy`，不直接操作 Docker SDK

- 资源限额、网络隔离、超时清理等策略集中管理

- 便于替换沙箱方案（如 Firecracker）

**核心接口**：

```go

type SandboxConfig struct {

    IssueID        uint

    Image          string

    AllowedDomains []string

    MaxMemoryMB    int

    TimeoutMinutes int

    EnvVars        map[string]string

    ExistingID     string // 复用已有容器

}

type SandboxManager struct {

    cli    *client.Client

    config Config

}

func (m *SandboxManager) Create(ctx context.Context, cfg SandboxConfig) (*SandboxHandle, error)

func (m *SandboxManager) Pause(ctx context.Context, containerID string) error

func (m *SandboxManager) Resume(ctx context.Context, containerID string) error

func (m *SandboxManager) Destroy(ctx context.Context, containerID string) error

func (m *SandboxManager) IsAlive(ctx context.Context, containerID string) bool

```

> 实现细节：容器名 `anserflow-issue-{id}`，`AutoRemove=false`（保障 Worker 重启后可恢复），`Memory` 和网络白名单来自 `SandboxConfig`。

### RuntimeManager — 运行时管理器（适配器模式）

运行时配置构建（解密 API Key、拼接 execute_template、写 config.json）当前直接写在 Worker 中，通过 `RuntimeAdapter` 接口抽象运行时差异，使切换 AI 工具只需实现接口 + 在 `runtimes` 表插入一行。

**设计原则**：

- `RuntimeAdapter` 接口定义配置注入、环境变量映射等运行时差异点

- `OutputParser` 接口定义 stdout 解析、Token 采集等输出差异点

- 新运行时注册只需：实现接口 + 插入 `runtimes` 表

- Worker 不感知具体运行时，只调用接口

**接口定义**：

```go

// internal/runtime/adapter.go

package runtime

// RuntimeAdapter 运行时适配器 — 封装不同 AI 工具的配置注入差异

// 新运行时只需实现此接口，注册到 Registry 即可

type RuntimeAdapter interface {

    // Name 运行时标识（对应 runtimes.name）

    Name() string

    // HomeDir 运行时在容器内的主目录（bind mount 目标）

    // 项目级 runtime 目录会整体 bind mount 到此路径（读写）

    // opencode → /home/sandbox/.opencode

    // hermes → /home/sandbox/.hermes

    HomeDir() string

    // ConfigPath 配置文件写入路径（容器内，HomeDir 的子路径）

    // opencode → /home/sandbox/.opencode/config.json

    // hermes → /home/sandbox/.hermes/config.yaml

    ConfigPath() string

    // RenderConfig 将通用配置渲染为该工具特有的配置 JSON

    // 不同工具的配置结构差异封装在此方法内

    RenderConfig(config map[string]interface{}) (string, error)

    // EnvMapping API Key → 环境变量映射

    // opencode + openai → OPENAI_API_KEY

    // hermes + openrouter → OPENROUTER_API_KEY

    EnvMapping(config map[string]interface{}, decryptedKey string) map[string]string

    // SkillsMountPath Skills 目录在容器内的路径（HomeDir 子路径）

    // opencode → /home/sandbox/.opencode/skills

    // hermes → /home/sandbox/.hermes/skills

    SkillsMountPath() string

    // SessionPath 会话文件路径（用于事后 Token 汇总），空字符串表示不支持

    // opencode → /home/sandbox/.local/share/opencode/sessions/*.jsonl

    // hermes → /home/sandbox/.hermes/sessions/*.jsonl

    SessionPath() string

}

// OutputParser 输出解析器 — 封装不同 AI 工具的 stdout 格式差异

type OutputParser interface {

    // ParseLine 解析单行 stdout 输出

    // 返回 nil 表示该行不是结构化事件（忽略或按纯文本处理）

    ParseLine(line []byte) *ParsedEvent

    // ParseSessionFile 解析会话文件（执行完成后调用）

    // 返回 nil 表示该运行时不支持 session 文件

    ParseSessionFile(content []byte) (*TokenSummary, error)

}

// ParsedEvent 解析后的单行事件

type ParsedEvent struct {

    Type       string            // "agent_log" / "token_usage" / "error"

    Content    string            // 文本内容（写入 issue_timeline）

    TokenUsage *TokenUsageDetail // Token 用量（非空时写入 token_tracker）

}

// TokenUsageDetail Token 用量明细

type TokenUsageDetail struct {

    InputTokens      int64

    OutputTokens     int64

    CacheReadTokens  int64

    CacheWriteTokens int64

}

// TokenSummary 会话文件 Token 汇总

type TokenSummary struct {

    TotalInput  int64

    TotalOutput int64

}

```

**Registry — 运行时注册表**：

```go

// internal/runtime/registry.go

package runtime

// Registry 全局运行时注册表，初始化时注册内置运行时

type Registry struct {

    adapters map[string]RuntimeAdapter

    parsers  map[string]OutputParser

}

func NewRegistry() *Registry {

    r := &Registry{

        adapters: make(map[string]RuntimeAdapter),

        parsers:  make(map[string]OutputParser),

    }

    // 注册内置运行时

    r.Register(&OpenCodeAdapter{}, &OpenCodeParser{})

    r.Register(&HermesAdapter{}, &HermesParser{})

    return r

}

func (r *Registry) Register(a RuntimeAdapter, p OutputParser) {

    r.adapters[a.Name()] = a

    r.parsers[a.Name()] = p

}

func (r *Registry) GetAdapter(name string) RuntimeAdapter { return r.adapters[name] }

func (r *Registry) GetParser(name string) OutputParser    { return r.parsers[name] }

```

**内置运行时路径映射**：

| 属性 | opencode | hermes |

|------|----------|--------|

| HomeDir | `/home/sandbox/.opencode` | `/home/sandbox/.hermes` |

| ConfigPath | `.opencode/config.json` | `.hermes/config.yaml` |

| SkillsMountPath | `.opencode/skills` | `.hermes/skills` |

| SessionPath | `.local/share/opencode/sessions/*.jsonl` | `.hermes/sessions/*.jsonl` |

| EnvKey(openai) | `OPENAI_API_KEY` | `OPENAI_API_KEY` |

| EnvKey(openrouter) | — | `OPENROUTER_API_KEY` |

**RuntimeManager 核心方法**：

```go

type RuntimeManager struct {

    registry *Registry

    crypto   CryptoService

}

// ResolveConfig 通过适配器解析运行时配置：

//   1. 解密 API Key → 2. adapter.RenderConfig → 3. 渲染 execute_template → 4. adapter.EnvMapping

func (m *RuntimeManager) ResolveConfig(ctx context.Context, runtimeName string, ...) (*ResolvedSandboxConfig, error)

// GetParser 获取输出解析器

func (m *RuntimeManager) GetParser(runtimeName string) OutputParser

```

**Worker 调用（运行时无关）**：

```go

resolved, _ := runtimeMgr.ResolveConfig(ctx, runtime.Name, runtime.ExecuteTemplate, agent.RuntimeConfig, issue.Prompt)

parser := runtimeMgr.GetParser(runtime.Name)

// stdout 解析通过 parser.ParseLine，事件统一处理

```

---

## 附录：沙箱镜像与执行细节

> 以下内容从 [04-sandbox.md](04-sandbox.md) §七中提取，为沙箱执行的实现级细节。

### Dockerfile

位于 `docker/sandbox/Dockerfile`：

```bash
docker build -t anserflow/sandbox:latest -f docker/sandbox/Dockerfile .
```

```dockerfile
FROM alpine:3.21

RUN apk add --no-cache \
    nodejs npm \
    python3 py3-pip \
    git curl bash \
    && rm -rf /var/cache/apk/*

RUN adduser -D -u 1000 sandbox

RUN npm install -g opencode-ai@latest \
    && npm cache clean --force

RUN mkdir -p /workspace && chown sandbox:sandbox /workspace

WORKDIR /workspace

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

USER sandbox
ENTRYPOINT ["/entrypoint.sh"]
```

> 预估镜像大小约 400MB。`.dockerignore` 排除 node_modules/ / .git/ / dist/ / .next/ / *.log。

### entrypoint.sh

```bash
# 由 Worker 传入环境变量和配置注入文件控制行为：
#   GIT_REPO_URL / GIT_BRANCH / GITHUB_TOKEN → git clone
#   TASK_PROMPT                              → opencode run 的 prompt 参数
#
# opencode 配置由 Worker 在容器启动后通过以下方式注入（每次覆盖）：
#   ① 写入 ~/.config/opencode/config.json   → opencode 读取 provider / model 配置
#   ② 注入环境变量                           → API Key 不落盘
#   ③ opencode run --model provider/model    → 运行时指定模型
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
func execOpenCode(ctx, containerID, issue, task) error
func destroyProjectSandbox(ctx, projectID, containerID) error
```

### 运行时数据目录

```
/var/lib/anserflow/               ← runtime_data_dir
├── runtimes/                     ← Layer 1: 全局模板
│   └── opencode/
│       ├── skills/
│       ├── config.json
│       └── plugins/
├── projects/                     ← Layer 2: 项目实例
│   └── 42/
│       └── runtime/
│           ├── skills/
│           ├── config.json
│           └── plugins/
沙箱容器                         ← Layer 3: bind mount
└── /home/sandbox/.opencode/
    ├── skills/   ← bind mount 自 projects/42/runtime/
    ├── config.json
    └── plugins/
```

```go
// 项目创建时从全局模板递归复制到项目实例目录，幂等
func initProjectRuntime(ctx, projectID, runtimeName) (string, error)
```

### GitHub SDK 集成

| SDK | 用途 | 运行位置 |
|------|------|----------|
| **go-git** | `clone`/`commit`/`push`/`checkout` | Docker 沙箱内 |
| **go-github/v68** | 创建 PR/Issue/Review/读取仓库 | Go 后端 Service 层 |

```go
client := github.NewClient(nil).WithAuthToken(token)
client.Issues.Create(ctx, owner, repo, &github.IssueRequest{...})
client.PullRequests.Create(ctx, owner, repo, &github.NewPullRequest{...})
```

**Token 权限**：`repo` + `issues:write` + `pull_requests:write`。

**多平台扩展**（Phase 2）：通过 `GitPlatform` + `GitOps` 双接口抽象，支持 Gitea / GitLab / Gitee。
