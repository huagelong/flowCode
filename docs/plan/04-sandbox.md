# AnserFlow - 沙箱 / Agent 基础设施

> **职责边界**：本文档覆盖 Agent 基础设施层（Eino 框架、Docker 沙箱、指令处理、Token 追踪、Skill 导入）。Agent 大脑设计（五层记忆、Skill 自改进、调度编排）见 [06-agent.md](06-agent.md)。
>
> 参考代码映射见 [07-architecture.md](07-architecture.md) §建议保留的模块 / 建议废弃的模块 / 建议做映射迁移的模块。

---

## 六、AI Agent 框架：Eino + 自研封装

### 架构分层

```

┌──────────────────────────────────────────┐

│  anserflow/internal/agent/  (自研业务层)  │

│  ├── orchestrator.go    讨论编排          │

│  ├── command_handler.go /backlog 指令识别 │

│  ├── executor.go        Docker沙箱调度    │

│  ├── skill_loader.go    Skills 加载       │

│  ├── group_discuss.go   群聊 Agent 调度   │

│  ├── backlog_parser.go   方案→Issue 拆解   │

│  ├── prompt_optimizer.go 人工提示词 Eino 优化│

│  ├── prompt_manager.go  提示词统一管理     │

│  └── tool/                                 │

│       ├── registry.go    Tool 注册中心     │

│       ├── dispatch.go    LLM 输出解析→执行 │

│       ├── create_issue.go  创建 Issue     │

│       ├── send_message.go  群聊发言       │

│       ├── read_timeline.go 读取时间线     │

│       ├── change_status.go 状态流转       │

├──────────────────────────────────────────┤

│  anserflow/prompts/  (提示词统一管理)      │

│  ├── agent_optimizer.go  anserAgent 优化器提示词│

│  ├── agent_mention.go    @Agent 任务注入模板  │

│  ├── agent_memory.go     五层记忆注入提示词模板│

│  ├── agent_new_session.go 新会话提示词       │

│  ├── system_messages.go  群聊系统消息模板     │

│  ├── notification.go     通知标题/正文模板    │

│  └── error_templates.go  错误消息模板        │

├──────────────────────────────────────────┤

│  anserflow/internal/status/  (状态机)      │

│  └── status_manager.go  Issue 状态流转管理  │

├──────────────────────────────────────────┤

│  anserflow/internal/sandbox/ (沙箱生命周期) │

│  └── sandbox_manager.go 创建/暂停/恢复/销毁 │

├──────────────────────────────────────────┤

│  anserflow/internal/runtime/ (运行时管理)   │

│  └── runtime_manager.go 配置构建/命令渲染   │

├──────────────────────────────────────────┤

│  anserflow/internal/notification/ (通知)   │

│  └── channel_manager.go 多渠道路由分发      │

├──────────────────────────────────────────┤

│  anserflow/internal/git/  (Git 管理)       │

│  └── manager.go        GitManager 统一入口 │

├──────────────────────────────────────────┤

│  anserflow/internal/token/  (Token 配额)   │

│  └── token_manager.go   用量追踪/配额/归档  │

├──────────────────────────────────────────┤

│  Eino (底层 LLM 引擎)                    │

│  ├── ChatModel         模型调用           │

│  ├── Tool              工具/Skills 抽象   │

│  ├── Graph             多Agent编排        │

│  ├── Callbacks         回调/日志/监控     │

│  └── Flow              流式处理           │

├──────────────────────────────────────────┤

│  ⚠️ anserAgent 职责边界 + 五层记忆驱动          │

│  anserAgent 负责：Agent 调度 / 群聊讨论编排    │

│               /backlog 方案拆解               │

│               人工提示词优化 / Skill 自改进    │

│  anserAgent 不负责：代码生成 / 编码执行        │

│              └→ 由运行时（opencode/hermes）在 Docker 沙箱完成│

│                                           │

│  调度行为由五层记忆 + L3 Skills/SOPs 驱动    │

│  详见 docs/plan/06-agent.md               │

└──────────────────────────────────────────┘

```

**Agent 编排通用规则（群聊 + 双人聊）**：

Hub 路由层统一根据 `HasAgentMember()` 决定是否调用 Agent 组件，不再区分 `group.type`：

| 场景 | GroupOrchestrator | CommandHandler | MentionResolver | 说明 |

|------|-------------------|----------------|-----------------|------|

| 群聊（含 Agent） | ✅ 调用 | ✅ /backlog /todo /new 可用 | ✅ 调用 | 原有群聊行为不变 |

| 群聊（无 Agent） | ❌ 不调用 | ✅ 仅 /new 可用 | ❌ 不调用 | 纯自然人聊天，不触发 Eino |

| 双人聊（人+Agent） | ✅ 调用 | ✅ /backlog /todo /new 可用 | ❌ 不调用 | Agent 自动参与，与群聊含 Agent 一致 |

| 双人聊（人+人） | ❌ 不调用 | ✅ 仅 /new 可用 | ❌ 不调用 | 纯自然人聊天，不触发 Eino |

以下组件**零修改**，Hub 层通过 `HasAgentMember()` 控制调用入口：

- `GroupOrchestrator.InvokeAgent` — 有 Agent 的群聊和双人聊直接复用

- `CommandHandler` — Hub 判断后调用；/new 全模式可用，/backlog 和 /todo 无 Agent 时返回提示

- `ToolRegistry` — Agent Tool（create_issue / send_message 等）在所有含 Agent 的会话中同样适用

### Eino 初始化与配置

Eino 使用 `config.yaml` 中 `llm` 配置段统一管理多 Agent 的模型连接：

```go

// internal/agent/agent_init.go

package agent

import (

    "github.com/cloudwego/eino/components/model"

    "github.com/cloudwego/eino/schema"

    "github.com/cloudwego/eino-ext/components/model/openai"

)

var chatModel model.ChatModel

func InitEino(cfg *config.EinoConfig) error {

    var err error

    chatModel, err = openai.NewChatModel(context.Background(), &openai.ChatModelConfig{

        APIKey:      cfg.APIKey,

        BaseURL:     cfg.BaseURL,

        Model:       cfg.Model,

        MaxTokens:   cfg.MaxTokens,

        Temperature: &cfg.Temperature,

        Timeout:     cfg.Timeout,

    })

    return err

}

// GetChatModel 返回初始化的模型实例（单例）

func GetChatModel() model.ChatModel { return chatModel }

```

### Agent 运行时配置

`agents.runtime_config` JSON 由绑定的运行时决定其 schema。`runtimes.config_schema` 定义了该运行时可配置的所有字段，前端根据 schema 动态生成表单：

**opencode 运行时配置示例**（`runtimes.config_schema` 驱动）：

```json

{

  "provider": "openai",

  "model": "gpt-4o",

  "agent": "build",

  "api_key_encrypted": "aes256:xxx",

  "max_iterations": 20,

  "thinking": true

}

```

**配置流转**：

```

Admin UI (Agent 编辑页)

│  ① 下拉选择运行时（opencode / hermes / ...）

│  ② 前端根据 runtimes.config_schema 动态渲染配置表单

│  ③ 保存 → agents.runtime_config JSON

│

▼

Worker (沙箱启动时)

│  ① 读取 agents.runtime_id → 确定运行时（opencode / hermes）

│  ② 读取 agents.runtime_config → 填充模板变量

│  ③ 通过 RuntimeClient 接口与沙箱内 opencode/hermes 建立双向流连接：

│     - SandboxClient: 启动 Docker 容器，通过 ContainerAttach 双向通讯

│     - LocalClient:   启动本地子进程，通过 stdin/stdout 双向通讯

│     - 两种模式使用同一套 JSON Lines 通讯协议

│  ④ 发送任务 → 流式接收日志、状态、结果事件

│  ⑤ 任务结束 → 关闭连接，销毁容器/终止子进程

```

### ChatModel 调用示例

```go

// internal/agent/group_discuss.go — Agent 参与群聊讨论

// 现由 anserAgent 统一编排，此处仅保留调用入口

func (o *GroupOrchestrator) InvokeAgent(

    ctx context.Context,

    agent *model.Agent,

    messages []*schema.Message,  // 群聊历史上下文

) (*schema.Message, error) {

    // 委托给 anserAgent（详见 06-agent.md）

    resp, err := o.anserAgent.Invoke(ctx, AgentInput{

        Context:  messages,

        Mode:     ModeOrchestrate,

        Tags:     []string{"group_discuss"},

    })

    if err != nil {

        return nil, fmt.Errorf("anserAgent invoke failed: %w", err)

    }

    return &schema.Message{Content: resp.Content}, nil

}

```

### Tool / Skill 抽象

```go

// internal/agent/skill_loader.go — Skills 加载为 anserAgent Tool

type SkillLoader struct {

    skillRepo *repository.SkillRepo

}

func (sl *SkillLoader) LoadAsTools(

    ctx context.Context,

    agentID uint,

) ([]tool.InvokableTool, error) {

    skills, err := sl.skillRepo.FindEnabledByAgent(ctx, agentID)

    if err != nil {

        return nil, err

    }

    tools := make([]tool.InvokableTool, 0, len(skills))

    for _, skill := range skills {

        // 每个 Skill 注册为一个 anserAgent Tool

        tools = append(tools, &SkillTool{

            name:        skill.Name,

            description: skill.Description,

            definition:  skill.Definition, // Markdown 正文

            handler: func(ctx context.Context, input string) (string, error) {

                // 将 Skill 定义注入 System Prompt 让 LLM 遵循

                return fmt.Sprintf(

                    "Skill '%s' loaded. Instructions:\n%s",

                    skill.Name, skill.Definition,

                ), nil

            },

        })

    }

    return tools, nil

}

```

### Skill 两层继承（沙箱执行时）

Worker 通过 RuntimeClient 向沙箱注入 Skills 配置时，合并 Runtime 默认 + Agent 独立绑定，Agent 可覆盖关闭 Runtime 继承的 Skill。Skills 以 JSON Lines 消息随任务一并发送给沙箱内运行的工具：

```go

// internal/agent/skill_loader.go

func (sl *SkillLoader) LoadForSandbox(ctx context.Context, agent *model.Agent) ([]*model.Skill, error) {

    // ① Runtime 默认 Skills（如 opencode→anser-coder）

    runtimeSkills, _ := sl.skillRepo.FindEnabledByRuntime(ctx, agent.RuntimeID)

    // ② Agent 独立绑定的 Skills

    agentSkills, _ := sl.skillRepo.FindEnabledByAgent(ctx, agent.ID)

    // ③ 合并去重：Agent 级配置覆盖 Runtime 默认（以 skill_id 为 key）

    merged := make(map[uint]*model.Skill)

    for _, s := range runtimeSkills {

        merged[s.ID] = s

    }

    for _, s := range agentSkills {

        merged[s.ID] = s    // Agent 级覆盖（含 enabled=false 的情况）

    }

    // ④ 仅返回 enabled=true 的 Skill

    result := make([]*model.Skill, 0)

    for _, s := range merged {

        if s.Enabled {

            result = append(result, s)

        }

    }

    return result, nil

}

```

**Skill 注入规则**：

| Skill | 来源 | 能否关闭 | 说明 |

|-------|------|---------|------|

| `anser-coder` | Runtime 默认（opencode） | ❌ 不可关闭 | `is_builtin=1`，前端灰掉开关 |

| 用户创建的 Skill | Runtime 默认 / Agent 绑定 | ✅ 可开关 | 后台自由管理 |

| Agent 主动关闭 Runtime Skill | Agent 级覆盖 | ✅ | `agent_skills.enabled=false` 覆盖 Runtime 默认 |

### 提示词管理器（PromptManager）

系统中除 Agent System Prompt（DB 存储）和 Skill 定义（DB 存储）外，所有硬编码的提示词模板统一抽到 `prompts/` 目录，由 `PromptManager` 集中加载和管理。

**设计原则**：

- 提示词模板与业务逻辑分离，便于调优、国际化、版本管理

- 运行时通过 `PromptManager.Get("key")` 获取，不直接拼接字符串

- DB 存储的 Agent System Prompt 和 Skill 定义不动，仍走后台 CRUD

**提示词文件清单**：

| 文件 | key | 用途 | 原硬编码位置 |

|------|-----|------|------------|

| `agent_optimizer.go` | `agent.optimizer.system` | anserAgent 优化器系统提示词 | `prompt_optimizer.go` |

| `agent_optimizer.go` | `agent.optimizer.user` | anserAgent 优化器用户消息模板 | `prompt_optimizer.go` |

| `agent_mention.go` | `agent.mention.system` | @Agent 任务注入系统提示词 | `group_discuss.go` |

| `agent_memory.go` | `agent.memory.inject` | 五层记忆注入模板 | `memory/manager.go` |

| `agent_new_session.go` | `agent.new_session.hint` | 新会话 Agent 感知提示词 | `group_discuss.go` |

| `system_messages.go` | `system.issue.{status_change}` | 群聊系统消息模板（按状态变更） | `IssueService` |

| `notification.go` | `notify.issue_status_changed.title` | 通知标题模板 | `notification_service.go` |

| `error_templates.go` | `error.backlog_empty` | /backlog 错误提示 | `command_handler.go` |

| `error_templates.go` | `error.backlog_insufficient` | /backlog 上下文不足 | `command_handler.go` |

| `error_templates.go` | `error.backlog_no_plan` | /backlog 生成失败 | `command_handler.go` |

**实现**：

```go

// prompts/prompt_manager.go

package prompts

import (

    "embed"

    "sync"

)

//go:embed *.go

var promptsFS embed.FS

type PromptManager struct {

    mu       sync.RWMutex

    prompts  map[string]string

}

var defaultManager *PromptManager

func init() {

    defaultManager = &PromptManager{prompts: make(map[string]string)}

    defaultManager.loadAll()

}

// Get 获取提示词模板（支持占位符替换）

func Get(key string, args ...interface{}) string {

    defaultManager.mu.RLock()

    defer defaultManager.mu.RUnlock()

    tmpl, ok := defaultManager.prompts[key]

    if !ok {

        return key // fallback: 返回 key 本身

    }

    if len(args) > 0 {

        return fmt.Sprintf(tmpl, args...)

    }

    return tmpl

}

// MustGet 获取提示词模板（找不到则 panic）

func MustGet(key string, args ...interface{}) string {

    result := Get(key, args...)

    if result == key {

        panic(fmt.Sprintf("prompt not found: %s", key))

    }

    return result

}

// Reload 热重载（开发/运维调试用）

func Reload() {

    defaultManager.mu.Lock()

    defer defaultManager.mu.Unlock()

    defaultManager.loadAll()

}

```

```go

// prompts/agent_optimizer.go

package prompts

func init() {

    defaultManager.prompts["agent.optimizer.system"] = `你是提示词优化器。将用户反馈改写为精确的编码指令。

规则：保留用户原意；补充技术细节；如果用户提到具体文件/组件，添加文件路径。

如果提供了关联 Issue 的内容，确保生成代码时保持与关联 Issue 的技术方案一致。`

    defaultManager.prompts["agent.optimizer.user"] = `Issue 上下文：

%s

关联 Issue 内容：

%s

用户提示词：%s

输出优化后的编码指令：`

}

```

```go

// prompts/agent_mention.go

package prompts

func init() {

    defaultManager.prompts["agent.mention.system"] = `你是 %s。%s 正在群聊中向你布置任务。`

}

```

```go

// prompts/agent_memory.go

package prompts

func init() {

    defaultManager.prompts["agent.memory.inject"] = `[五层记忆] 以下是你的相关记忆，用于辅助决策：\n%s`

}

```

```go

// prompts/agent_new_session.go

package prompts

func init() {

    defaultManager.prompts["agent.new_session.hint"] = `这是一个新讨论主题的开端`

}

```

```go

// prompts/system_messages.go

package prompts

func init() {

    defaultManager.prompts["system.issue.backlog_created"] = `已生成 Issue #%d（backlog），请到 backlog Tab 确认细节并启动`

    defaultManager.prompts["system.issue.to_todo"] = `Issue #%d 已确认转为 todo，排队等待执行`

    defaultManager.prompts["system.issue.start"] = `Issue #%d 开始执行，Agent 已启动`

    defaultManager.prompts["system.issue.paused"] = `Issue #%d 执行已暂停`

    defaultManager.prompts["system.issue.stopped"] = `Issue #%d 执行已停止`

    defaultManager.prompts["system.issue.pr_submitted"] = `Issue #%d PR 已提交，等待审核`

    defaultManager.prompts["system.issue.failed"] = `Issue #%d 执行失败: %s`

    defaultManager.prompts["system.issue.done"] = `Issue #%d 已完成 ✅`

    defaultManager.prompts["system.issue.pr_rejected"] = `Issue #%d PR 被拒绝，已退回 todo`

}

```

```go

// prompts/notification.go

package prompts

func init() {

    defaultManager.prompts["notify.issue_status_changed.title"] = `Issue #%d 状态变更为 %s`

}

```

```go

// prompts/error_templates.go

package prompts

func init() {

    defaultManager.prompts["error.backlog_empty"] = `❌ 需要描述需求或先进行群聊讨论，不能为空。`

    defaultManager.prompts["error.backlog_insufficient"] = `❌ 群聊讨论内容不足，请先描述需求或补充更多讨论后再试。`

    defaultManager.prompts["error.backlog_no_plan"] = `❌ 未能产出有效方案，请补充更多需求细节后重试。`

}

```

**业务代码调用示例**（原硬编码替换为 PromptManager）：

```go

// 替换前（硬编码）：

schema.SystemMessage(`你是提示词优化器。将用户反馈改写为精确的编码指令。...`)

// 替换后（从 PromptManager 获取）：

schema.SystemMessage(prompts.Get("agent.optimizer.system"))

// 替换前：

h.ws.Reply(msg, "❌ 需要描述需求或先进行群聊讨论，不能为空。")

// 替换后：

h.ws.Reply(msg, prompts.Get("error.backlog_empty"))

// 替换前：

s.messageService.SendSystemMessage(ctx, issue.SourceGroupID, fmt.Sprintf(

    "Issue #%d %s → %s", issue.ID, oldStatus, newStatus))

// 替换后：

s.messageService.SendSystemMessage(ctx, issue.SourceGroupID,

    prompts.Get(fmt.Sprintf("system.issue.%s", statusKey), issue.ID))

```

### IssueStatusManager — Issue 状态机管理器

Issue 状态流转逻辑分散在 `IssueService`、`Worker`、`Scheduler`、`WebhookHandler` 四处，通过 `StatusManager` 集中管理状态机规则和副作用触发。

**设计原则**：

- 状态流转合法表集中定义，不散落在 if/switch 中

- 每次流转的副作用（群聊通知、时间线、通知推送）统一 Hook 回调

- 新增状态或修改流转规则只需改一处

**状态流转合法表**：

| from | to | 触发方 | 副作用 |

|------|----|--------|--------|

| `backlog` | `todo` | 人工确认 | 群聊通知 + 时间线 |

| `todo` | `in_progress` | Scheduler 分配 | 群聊通知 + 时间线 + 通知被分配人 |

| `in_progress` | `in_review` | Worker PR 提交 | 群聊通知 + 时间线 + 通知 Reviewer |

| `in_progress` | `paused` | 人工暂停 | 群聊通知 + 时间线 |

| `in_progress` | `todo` | Worker 执行失败 | 群聊通知 + 时间线 + retry_count++ |

| `in_review` | `done` | GitHub Webhook merge | 群聊通知 + 时间线 |

| `in_review` | `todo` | GitHub Webhook reject | 群聊通知 + 时间线（沙箱保留） |

| `todo` | `backlog` | 人工退回 | 时间线 |

**实现**：

```go

// internal/status/status_manager.go

package status

import (

    "context"

    "fmt"

)

// Transition 表示一次状态流转

// type Transition struct {

//     From string

//     To   string

// }

type TransitionHook func(ctx context.Context, issueID uint, from, to string) error

type StatusManager struct {

    transitions map[string]map[string]bool   // from -> to -> allowed

    hooks       map[string][]TransitionHook   // "from->to" -> hooks

}

func NewStatusManager() *StatusManager {

    m := &StatusManager{

        transitions: make(map[string]map[string]bool),

        hooks:       make(map[string][]TransitionHook),

    }

    // 注册合法流转

    m.allow("backlog", "todo")

    m.allow("todo", "in_progress")

    m.allow("todo", "backlog")

    m.allow("in_progress", "in_review")

    m.allow("in_progress", "paused")

    m.allow("in_progress", "todo")

    m.allow("in_review", "done")

    m.allow("in_review", "todo")

    m.allow("paused", "in_progress")

    return m

}

func (m *StatusManager) allow(from, to string) {

    if m.transitions[from] == nil {

        m.transitions[from] = make(map[string]bool)

    }

    m.transitions[from][to] = true

}

// OnTransition 注册流转副作用 Hook

func (m *StatusManager) OnTransition(from, to string, hook TransitionHook) {

    key := from + "->" + to

    m.hooks[key] = append(m.hooks[key], hook)

}

// CanTransition 检查流转是否合法

func (m *StatusManager) CanTransition(from, to string) bool {

    return m.transitions[from][to]

}

// Transition 执行状态流转（校验 + 触发 hooks）

func (m *StatusManager) Transition(ctx context.Context, issueID uint, from, to string) error {

    if !m.CanTransition(from, to) {

        return fmt.Errorf("invalid transition: %s -> %s", from, to)

    }

    key := from + "->" + to

    for _, hook := range m.hooks[key] {

        if err := hook(ctx, issueID, from, to); err != nil {

            return fmt.Errorf("hook failed for %s: %w", key, err)

        }

    }

    return nil

}

```

**业务代码接入**：

```go

// 初始化时注册 hooks

statusMgr := status.NewStatusManager()

// 所有含群聊通知的流转都注册群聊 Hook

statusMgr.OnTransition("backlog", "todo",

    func(ctx context.Context, issueID uint, from, to string) error {

        issue, _ := issueRepo.FindByID(ctx, issueID)

        if issue.SourceGroupID != 0 {

            msgService.SendSystemMessage(ctx, issue.SourceGroupID,

                prompts.Get("system.issue.to_todo", issueID))

        }

        return nil

    },

)

// 业务代码调用：一行代替多处 if/通知/时间线

statusMgr.Transition(ctx, issueID, "backlog", "todo")

```

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

### NotificationChannelManager — 通知渠道管理器

`NotificationService` 中 WS 推送、浏览器通知、群聊系统消息、邮件通知四种渠道的触发逻辑硬编码在多个方法里，通过 `ChannelManager` 统一分发。

**设计原则**：

- 统一入口：`manager.Notify(event, payload)` → 自动分发到用户开通的渠道

- 新增渠道只需注册新 Channel，不改业务代码

- 用户通知偏好统一查询

**核心接口**：

```go

type Event string

const (

    EventIssueAssigned     Event = "issue_assigned"

    EventIssueStatusChange Event = "issue_status_changed"

    EventAgentCompleted    Event = "agent_completed"

    EventMention           Event = "mention"

    EventNewDM             Event = "new_dm"

)

type Channel interface {

    Name() string

    Send(ctx context.Context, userID uint, payload *NotifyPayload) error

}

type ChannelManager struct {

    channels  []Channel

    userPrefs UserPreferenceService

}

// Register 注册渠道 → Notify 查询偏好后按渠道分发

// NotifyGroup 直接写入 messages 表（不受用户偏好控制）

```

**渠道注册**：`WebSocketChannel`、`EmailChannel`、`BrowserChannel` 三种渠道在初始化时注册。

### GitManager — Git 管理器

统一管理 Git 平台 API 和仓库操作，分为两个子接口：

| 子接口 | 职责 | 运行位置 | 平台相关性 |

|--------|------|----------|-----------|

| **GitPlatform** | 平台 REST API（Issue/PR/Repo） | Go 后端 Service 层 | 平台相关 |

| **GitOps** | 仓库操作（Clone/Fetch/Push/Commit） | Worker 沙箱内 | 平台无关 |

**设计原则**：

- `manager.Platform(platform)` → 返回对应平台 API 实现

- `manager.NewOps(containerID, workdir)` → 返回仓库操作实例（绑定容器）

- 业务代码不感知具体平台和底层 Git 实现

- Phase 2 可替换为 go-git 库实现，上层无感知

**接口定义**：

```go

// internal/git/manager.go

package git

import (

    "context"

    "fmt"

)

// GitPlatform 平台 REST API 接口（平台相关，运行在 Go 后端 Service 层）

type GitPlatform interface {

    CreateIssue(ctx context.Context, repo, title, body string, labels []string) (issueID string, err error)

    CreatePR(ctx context.Context, repo, title, head, base, body string) (prURL string, err error)

    GetRepoInfo(ctx context.Context, repo string) (*RepoInfo, error)

    ListBranches(ctx context.Context, repo string) ([]string, error)

}

// GitOps 仓库操作接口（平台无关，运行在 Worker 沙箱内）

type GitOps interface {

    IsRepo(ctx context.Context) bool

    Clone(ctx context.Context, repoURL, branch, dest string) error

    FetchAll(ctx context.Context) error

    Checkout(ctx context.Context, branch string) error

    Pull(ctx context.Context, branch string) error

    Commit(ctx context.Context, message string, author Author) (string, error)

    Push(ctx context.Context) error

}

// Author 提交者信息

type Author struct {

    Name  string

    Email string

}

// GitManager Git 管理器（统一入口）

type GitManager struct {

    platforms       map[string]GitPlatform

    defaultPlatform string

}

func NewGitManager() *GitManager {

    return &GitManager{

        platforms:       make(map[string]GitPlatform),

        defaultPlatform: "github",

    }

}

// Register 注册平台 API 实现

func (m *GitManager) Register(platform string, p GitPlatform) {

    m.platforms[platform] = p

}

// Platform 获取平台 API 实现（默认返回 github）

func (m *GitManager) Platform(platform string) (GitPlatform, error) {

    if platform == "" {

        platform = m.defaultPlatform

    }

    p, ok := m.platforms[platform]

    if !ok {

        return nil, fmt.Errorf("unsupported git platform: %s", platform)

    }

    return p, nil

}

// NewOps 为指定容器创建仓库操作实例

func (m *GitManager) NewOps(containerID, workdir string) GitOps {

    return NewContainerGitOps(containerID, workdir)

}

```

**ContainerGitOps — 容器内 Shell 实现**：通过 `docker exec` 映射 GitOps 方法到 Shell 命令：

| GitOps 方法 | 容器内 Shell 命令 |

|------------|------------------|

| `IsRepo` | `test -d {workdir}/.git` |

| `Clone(repoURL,branch,dest)` | `git clone --branch {branch} {repoURL} {dest}` |

| `FetchAll` | `cd {workdir} && git fetch --all` |

| `Checkout(branch)` | `cd {workdir} && git checkout {branch}` |

| `Pull(branch)` | `cd {workdir} && git checkout {branch} && git pull` |

| `Commit(msg,author)` | `cd {workdir} && git add . && git commit -m "{msg}" --author="{name} <{email}>"` |

| `Push` | `cd {workdir} && git push` |

> Phase 2 可选替换为 `GoGitOps`（go-git 库实现），上层通过 `GitOps` 接口无感知切换。

**初始化**：

```go

gitMgr := git.NewGitManager()

gitMgr.Register("github", &GitHubPlatform{token: cfg.GitHubToken})

// Phase 2: gitMgr.Register("gitea", &GiteaPlatform{...})

// Phase 2: gitMgr.Register("gitlab", &GitLabPlatform{...})

```

### TokenManager — Token 配额管理器

在现有 `TokenTracker` 基础上升级，增加配额检查、周期归档、用量报告能力。

**设计原则**：

- 配额检查：`manager.CheckQuota(orgID)` → 超额时暂停 Agent 执行

- 周期归档：每天凌晨 Redis → MySQL 持久化

- 用量报告：按组织/项目/Agent/日期多维度查询

**实现**：

```go

type Usage struct {

    PromptTokens     int64

    CompletionTokens int64

    Source           string // "agent" | "opencode"

}

type TokenManager struct {

    redis  *redis.Client

    quota  QuotaService

}

func (m *TokenManager) Record(ctx context.Context, agentID uint, usage *Usage)          // Redis Hash 按 Agent+日期聚合

func (m *TokenManager) CheckQuota(ctx context.Context, orgID uint) bool                  // 检查组织月度配额

func (m *TokenManager) GetDailyUsage(ctx context.Context, agentID uint, date string) (*DailyReport, error)

func (m *TokenManager) Archive(ctx context.Context, date string) error                  // 定时归档到 MySQL

```

> 用量以 Redis Hash 聚合（key=`tokens:agent:{id}:date:{date}`），TTL 30天；定期 Archive 写入 `token_usage` 表。

### anserAgent Tool 系统（Skill 与系统通信）

Skill 不只是 Markdown 文档，每个 Skill 声明一组可调用 Tool。anserAgent 调度 LLM 时，LLM 决定调用哪个 Tool → 执行对应的 Go handler → 操作数据库/群聊/通知。

**Skill 定义格式**（YAML frontmatter + Markdown）：

```markdown

---

name: issue-backlog

description: /backlog 方案拆解规范，将群聊讨论产出为一个 Issue

tools:

  - create_issue     # 创建 Issue（调用 IssueService）

  - read_issues      # 读取已有 Issue（防重复）

  - send_message     # 向群聊发送消息

is_builtin: true

---

# 方案拆解规范

## 触发条件

收到 /backlog 指令时调用 create_issue 工具。

## 创建规则

- title: 简洁的功能描述（<50字）

- description: 技术方案概述 + 验收标准

- priority: P0=核心路径 P1=重要功能 P2=增强

- 调用 read_issues 检查是否已存在相同 Issue

- 创建成功后调用 send_message 通知群聊

```

**Tool 注册与调度**：

```go

// internal/agent/tool/registry.go

type ToolRegistry struct {

    tools map[string]ToolHandler

}

type ToolHandler func(ctx context.Context, params json.RawMessage) (string, error)

func NewRegistry(services *Services) *ToolRegistry {

    r := &ToolRegistry{tools: make(map[string]ToolHandler)}

    // 注册 anserAgent 可调用的所有 Tool

    r.Register("create_issue",   services.IssueService.CreateFromAgent)

    r.Register("read_issues",    services.IssueService.ListByProject)

    r.Register("send_message",   services.WS.SendToGroup)

    r.Register("read_timeline",  services.TimelineRepo.FindByIssue)

    r.Register("change_status",  services.IssueService.UpdateStatus)

    r.Register("find_agent",     services.AgentRepo.FindByID)

    return r

}

// GetToolsSchema 生成 OpenAI Function Calling 格式的 tools 定义

func (r *ToolRegistry) GetToolsSchema(skillNames []string) []ToolDef {

    // 根据 Skill 声明的 tools 列表，返回对应的 Function 定义

}

```

```go

// internal/agent/tool/dispatch.go

func (d *Dispatcher) Execute(ctx context.Context, llmOutput string, agent *model.Agent) error {

    // ① 解析 LLM 输出的 JSON: {"tool": "create_issue", "params": {...}}

    var call ToolCall

    json.Unmarshal([]byte(llmOutput), &call)

    // ② Casbin 校验（Agent 是否有权限调用此 Tool）

    if !d.enforcer.Enforce(agent, call.Tool) {

        return fmt.Errorf("Agent %s 无权调用 %s", agent.Name, call.Tool)

    }

    // ③ 执行 Tool → 写入 DB / 发送 WS

    result, err := d.registry.Invoke(ctx, call.Tool, call.Params)

    // ④ 注入 agent_logs 记录

    d.logRepo.Create(ctx, &model.AgentLog{

        AgentID: agent.ID,

        Action:  call.Tool,

        Input:   call.Params,

        Output:  json.RawMessage(result),

        Status:  "success",

    })

    // ⑤ 返回结果给 LLM 继续上下文

    return err

}

```

**anserAgent Tool 与 opencode Tool 对比**：

| | anserAgent Tool | opencode Tool |

|------|---------|------------|

| 运行位置 | Go 后端进程 | Docker 沙箱内 |

| 操作对象 | 系统数据（Issue/消息/时间线） | 代码文件（read/write/bash） |

| 权限 | Casbin RBAC | 沙箱隔离 |

| 注册方式 | `registry.Register(name, handler)` | opencode 内置 |

| 典型调用 | `create_issue` / `send_message` | `read` / `write` / `bash` |

### `/backlog` 与 `/todo` 指令识别

anserAgent 在群聊中监听 WebSocket 消息，检测到 `/backlog` 或 `/todo` 指令时触发方案拆解流程。两者共享 anserAgent 编排逻辑，区别仅在于 Issue 创建时的初始状态：

```go

// internal/agent/command_handler.go

type CommandHandler struct {

    orchestrator *GroupOrchestrator

    parser       *BacklogParser

}

// HandleBacklog 统一处理 /backlog 和 /todo 指令，initialStatus 决定 Issue 初始状态

func (h *CommandHandler) HandleBacklog(msg *ws.Message, initialStatus string) {

    // ① 解析指令文本 + 收集群聊上下文（最近 50 条）

    // ② 输入校验：非空 + 上下文 ≥3 条

    // ③ anserAgent 产出方案 → parser 拆解为 Issue

    // ④ 写 DB + WS 广播结果

}

```

**`/todo` vs `/backlog` 对比**：

| 维度 | `/backlog` | `/todo` |

|------|-----------|--------|

| Agent 参与 | ✅ anserAgent 编排产出方案 | ✅ anserAgent 编排产出方案 |

| Issue 状态 | `backlog` | `todo`（跳过 backlog） |

| 人工确认 | 需确认后转为 todo | 无需确认，直接可执行 |

| 适用场景 | 需求模糊，需讨论出方案后再审 | 需求明确，希望快速推进到执行 |

### `/new` 指令 — 会话上下文隔离

自然人可在群聊中发送 `/new` 开启新会话。系统生成新的 `session_id`，后续 Agent 讨论仅感知该会话之后的消息历史：

```go

// internal/agent/command_handler.go — /new 指令：生成新 session_id → 广播 new_session 消息 → 后续消息自动携带

func (h *CommandHandler) HandleNewSession(msg *ws.Message) { /* uuid → broadcast → SetCurrent */ }

```

**上下文窗口规则**：

- `GroupOrchestrator.InvokeAgent` 收集群聊历史时，以当前 `session_id` 为过滤条件，仅加载该会话内的消息

- 历史会话消息仍在客户端可见（滚动加载），但不进入 Agent 的 LLM 上下文窗口

- `/new` 不携带额外参数时，仅重置上下文；后跟文本时（如 `/new 下一阶段：部署上线`），该文本作为新会话的首条消息

```go

// internal/agent/group_discuss.go — 上下文加载适配 session 隔离

func (o *GroupOrchestrator) getRecentMessages(groupID uint, limit int) []*schema.Message {

    sessionID := o.sessionManager.GetCurrent(groupID)

    // 按 session_id 过滤，只取当前会话的消息

    return o.msgRepo.FindByGroupAndSession(groupID, sessionID, limit)

}

```

### @Agent 任务布置

群内 Agent 在讨论过程中，可通过 `@AgentName` 语法向群内其他 Agent 布置任务。anserAgent 在调度时解析消息中的 `@AgentName`，匹配群内成员 Agent，将任务注入被 @ Agent 的上下文：

```go

// internal/agent/mention_resolver.go — 正则 `@(\S+)` 匹配消息中的 @AgentName，查询群成员匹配

func (r *MentionResolver) Resolve(ctx context.Context, groupID uint, text string) ([]*AgentMention, error)

```

**调度流程**：

1. Agent A 发言含 `@前端Agent 请实现登录表单组件`

2. `MentionResolver` 解析出 `@前端Agent`，匹配群内 Agent 成员

3. anserAgent 将任务描述 + Agent A 的上下文 + 被提及 Agent 的角色定义（`system_prompt` + 记忆系统）注入被提及 Agent 的调用链

4. 被提及 Agent 根据自身角色和 Skill 决定响应方式：直接回复方案、创建 Issue、或执行操作

```go

// internal/agent/group_discuss.go — 遍历 mentions，为每个被 @ Agent 构建上下文（角色+任务+讨论）→ 异步调度

func (o *GroupOrchestrator) InvokeWithMentions(ctx, agent, messages, mentions) { /* 异步 InvokeAgent */ }

```

Eino 在将人工提示词注入 opencode 之前，自动进行上下文增强与工程化改写：

```go

// internal/agent/prompt_optimizer.go — Eino 调用 LLM 改写用户提示词

// ① 收集 Issue 上下文 + 时间线

// ② 解析 @Issue #N 引用，合并关联 Issue 内容

// ③ LLM 将用户反馈改写为精确编码指令

func (o *PromptOptimizer) Enhance(ctx context.Context, rawPrompt string, issue *model.Issue) (string, error)

```

> **关键**：Eino 只做调度与提示词优化，不进入 Docker 沙箱。沙箱内的代码生成完全由 opencode 完成。

### Token 用量与成本追踪

系统 Token 消耗来自两个阶段，需要分别追踪后汇总：

| 阶段 | 谁调 LLM | 追踪方式 | 占比（典型） |

|------|---------|---------|------------|

| **anserAgent 调度** | Go 后端进程 | `TokenTracker.Record(agentID, usage, "agent")` | ~10-20% |

| **opencode 执行** | Docker 沙箱内 | 解析 stdout JSON + session 文件 | ~80-90% |

#### TokenTracker — 统一记录（区分来源）

```go

// internal/agent/token_tracker.go

type TokenTracker struct { redis *redis.Client }

// Record(agentID, usage, source) → 按 Agent+日期聚合 Redis Hash（key=tokens:agent:{id}:date:{date}, TTL 30天）

//   source="agent"  → agent_prompt_tokens / agent_completion_tokens

//   source="opencode" → opencode_prompt_tokens / opencode_completion_tokens

// GetDailyUsage(agentID, date) → prompt, completion 汇总

```

#### anserAgent 调度阶段 — 回调记录（已有）

```go

// anserAgent 阶段：Eino WithCallbacks → OnEnd 回调中 tokenTracker.Record(agentID, usage, "agent")

```

#### opencode 执行阶段 — 双通道采集

**通道 ① 实时：`opencode run --format json` stdout 解析**

```go

// internal/worker/executor.go — opencode run --format json → stdout JSON Lines 解析

//   解析 "token_usage" 字段 → tokenTracker.Record(agentID, ..., "opencode")

//   解析 "content" 字段 → timelineRepo.Append(issueID, ...)

//   非 JSON 行 → parseStdoutLine 兼容处理

```

**通道 ② 事后汇总：opencode session 文件解析**

opencode 在 `~/.local/share/opencode/sessions/` 下保存 JSONL 格式的会话文件，每条消息包含 `token_usage` 字段。Worker 在执行完成后从容器中提取：

```go

// internal/worker/session_parser.go — 事后汇总：读取 opencode session JSONL → 累加 token_usage → RecordFinal

//   去重：取 max(实时, 事后) 作为最终值，弥补实时 JSON 解析遗漏

func (w *Worker) collectSessionTokens(ctx, containerID, agentID, issueID) error

```

> **双通道去重策略**：实时通道在 opencode 执行过程中持续累加 token，事后通道在执行完成后读取 session 文件获得最终精确值。取 `max(实时, 事后)` 作为最终值，覆盖 Redis 中的 opencode 部分（通过 `RecordFinal` 实现）。这样即使实时通道部分 JSON 解析失败，事后通道也能兜底。

#### Token 总量公式

```

Agent 总 Token = anserAgent 调度 Token + opencode 执行 Token

              = (讨论 + /backlog + 提示词优化) + (编码 + 测试 + 修复 + PR)

```

| 来源 | 包含的 LLM 调用 | 触发位置 |

|------|---------------|---------|

| `agent` | 群聊 Agent 讨论、`/backlog` 方案拆解、`PromptOptimizer.Enhance()` | `anserAgent.Invoke` / `CommandHandler.HandleBacklog` |

| `opencode` | `opencode run` 全过程（读取代码、生成代码、运行测试、修复错误、commit） | `Worker.executeWithTokenTracking` |

> **LLM API Key 安全模型**：API Key 在 `agents.runtime_config.llm.api_key_encrypted` 中以 AES-256 加密存储；Agent 执行时 Worker 解密后通过环境变量注入 Docker 沙箱，不写入容器文件系统。

#### Token 用量 API 暴露

提供按 Agent/按组织维度的 Token 用量查询接口，用于 Dashboard 成本展示：

```go

// internal/handler/token_handler.go

// GET /api/orgs/:org_id/agents/:agent_id/token-usage?from=&to=  → TokenUsageResponse（含来源明细 + 成本估算）

// GET /api/orgs/:org_id/token-summary?period=7d                → OrgTokenSummary（按 Agent + 按日期聚合）

```

**成本估算函数**：

```go

// internal/agent/cost.go — 按 provider 单价估算（per 1M tokens）：

//   gpt-4o: $2.50/$10.00 | gpt-4o-mini: $0.15/$0.60 | claude-3.5: $3.00/$15.00 | deepseek-v3: $0.14/$0.28

func estimateCost(providerKey string, promptTokens, completionTokens int64) float64

```

---

---

## 七、Docker 沙箱方案

### 7.0 容器与代码隔离策略

本项目采用 **一个项目一个常驻容器 + 一个 Issue 一个 git worktree** 的隔离模型。

**设计决策**：

| 维度 | 旧方案（1 Issue = 1 容器） | 新方案（1 Project = 1 容器 + worktree） |

|------|---------------------------|----------------------------------------|

| 容器粒度 | 每个 Issue 创建新容器 | 每个 Project 一个常驻容器 |

| 代码隔离 | 共享 Named Volume，多 Issue 踩踏 | git worktree 物理隔离，零干扰 |

| 容器启动 | 每个 Issue 启动一次（秒级） | 项目启动一次，后续 exec 零开销 |

| 资源占用 | N 个容器 × 512MB | 1 个容器 × 1GB（共享） |

| 崩溃影响 | 单 Issue 容器死不影响别的 | 容器死 → 全部 Issue 受影响 |

| 崩溃恢复 | 需要逐个容器恢复 | 仅恢复一个容器 + 重建 worktree |

| 磁盘占用 | 1 份代码（Volume） | N+1 份 worktree（共享 .git/objects） |

**架构图**：

```

Project "my-app" (project_id=1)

│

├── 容器 anserflow-project-1 (常驻，项目创建时初始化)

│   ├── opencode / hermes / git / node / python / npm

│   ├── AutoRemove: false

│   ├── 资源: 1GB Memory, 2 CPU

│   │

│   ├── /workspace/

│   │   ├── main/                         ← git clone --bare 或初始 worktree（基准）

│   │   │   └── ...完整代码...

│   │   ├── issue-42/                     ← git worktree add -b feat/issue-42

│   │   │   └── src/login.tsx  (Worker A 独占，不受干扰)

│   │   └── issue-43/                     ← git worktree add -b feat/issue-43

│   │       └── src/api.ts    (Worker B 独占，不受干扰)

│   │

│   └── /home/sandbox/.opencode/          ← bind mount: 项目级运行时数据（Skills/配置）

│

├── Issue #42 (in_progress) ──► docker exec opencode run --workdir /workspace/issue-42

├── Issue #43 (in_progress) ──► docker exec opencode run --workdir /workspace/issue-43

└── Issue #44 (todo) ──────────► 等待分配

```

**git worktree 生命周期**：

```bash

# Issue 开始执行时创建 worktree

docker exec anserflow-project-1 git worktree add /workspace/issue-42 -b feat/issue-42 main

# Isssue 执行过程中，opencode 在 worktree 内开发

docker exec anserflow-project-1 opencode run --workdir /workspace/issue-42 "实现登录页"

# Issue 完成后清理（先 push 分支，再 remove worktree）

docker exec anserflow-project-1 git -C /workspace/issue-42 push origin feat/issue-42

# PR merge 后

docker exec anserflow-project-1 git worktree remove /workspace/issue-42

docker exec anserflow-project-1 git branch -D feat/issue-42

```

**合并冲突**：由于每个 Issue 在独立 worktree 中开发，代码层面不会出现文件级写冲突。

当两个 Issue 修改了同一文件的同一区域时，由 GitHub PR 流程在合并时提示冲突，

属于 Git 正常行为，不归 Worker 处理。

---

### 7.1 执行流程

每个 Issue 在**项目常驻容器**内通过**独立 git worktree** 执行，多 Issue 并发互不干扰：

```

┌──────────────────────────────────────────────────────────┐

│  Asynq Worker                                             │

│  ┌────────────────────────────────────────────────────┐  │

│  │  Step 1: 准备                                      │  │

│  │  ├── 读取 Issue 上下文                             │  │

│  │  ├── 加载 Agent 配置                               │  │

│  │  └── 加载绑定的 Skills                             │  │

│  │                                                    │  │

│  │  Step 2: 确保项目容器存活（常驻）                     │  │

│  │  ├── 查询 projects.sandbox_container_id             │  │

│  │  ├── 容器存在且运行中 → 直接复用                    │  │

│  │  ├── 容器不存在 → ContainerCreate + Start           │  │

│  │  │   ├── Image: anserflow/sandbox                  │  │

│  │  │   ├── Memory: 1GB, CPU: 2 cores                │  │

│  │  │   ├── NamedVol: workspace-{projectID}:/workspace│  │

│  │  │   ├── BindMount: project-runtime→home (rw)     │  │

│  │  │   └── AutoRemove: false (常驻)                  │  │

│  │  └── 首次: git clone 到 /workspace/main            │  │

│  │                                                    │  │

│  │  Step 3: 创建 Issue 专属 worktree                  │  │

│  │  ├── docker exec git worktree add \               │  │

│  │  │     /workspace/issue-{id} -b feat/issue-{id}    │  │

│  │  ├── 分支名: feat/issue-{id}                       │  │

│  │  └── 工作目录: /workspace/issue-{id}（物理隔离）    │  │

│  │                                                    │  │

│  │  Step 4: 注入运行时配置                             │  │

│  │  ├── 读取 Agent runtime_config                      │  │

│  │  ├── 解密 API Key → 环境变量                        │  │

│  │  ├── 写入 config.json 到容器                        │  │

│  │  ├── 写入 Runtime 默认 Skills                       │  │

│  │  └── 写入 Agent 绑定 Skills                         │  │

│  │                                                    │  │

│  │  Step 5: 在 worktree 内执行 opencode               │  │

│  │  │  docker exec \                                  │  │

│  │  │    --workdir /workspace/issue-{id} \            │  │

│  │  │    opencode run "实现登录页"                     │  │

│  │  │                                                  │  │

│  │  │  ┌─ opencode exec ←→ Worker 消息环 ──────────┐  │  │

│  │  │  │  opencode 通过 stdout JSON Lines 实时输出  │  │  │

│  │  │  │  Worker 解析 → DB + WebSocket → 前端       │  │  │

│  │  │  └──────────────────────────────────────────┘  │  │

│  │  │                                                  │  │

│  │  └── 退出码 + 工作区检查 → 判定通过/失败            │  │

│  │                                                    │  │

│  │  Step 6: 收尾                                      │  │

│  │  ├── opencode 检查通过:                            │  │

│  │  │   ├── git add → commit → push（在 worktree 内）│  │

│  │  │   ├── 创建 PR                                   │  │

│  │  │   ├── 更新 Issue → in_review                    │  │

│  │  │   └── git worktree remove + branch -D（清理）   │  │

│  │  ├── opencode 检查失败:                            │  │

│  │  │   ├── 写入失败原因到时间线                       │  │

│  │  │   ├── Issue → todo（保留 worktree 待重试）      │  │

│  │  │   └── 等待人工提示词 → 重试                     │  │

│  │  ├── Issue done（PR merge）:                       │  │

│  │  │   └── git worktree remove + branch -D            │  │

│  │  └── 项目容器不销毁（常驻，给其他 Issue 复用）      │  │

│  └────────────────────────────────────────────────────┘  │

└──────────────────────────────────────────────────────────┘

```

### 7.1.1 沙箱 ↔ 系统消息互通闭环

opencode 通过 **docker exec 的 stdout 流 + 退出码** 与 Worker 通信，Worker 负责解析、存储、推送：

```

沙箱内 opencode（docker exec）    宿主机 Worker                   前端 Issue 时间线

══════════════════════════        ════════════                   ════════════════

                                    ┌─ docker exec, attach stdout ─┐

                                    │                               │

"正在分析 Issue 描述..."  ──stdout──→ 解析 → agent_logs              │

                                    │         issue_timeline         │

                                    │         WS push ──────────────→ "Agent 开始分析需求"

                                    │                               │

"生成文件: src/login.tsx" ──stdout──→ 解析 → agent_logs              │

                                    │         WS push ──────────────→ "正在生成 src/login.tsx"

                                    │                               │

"运行测试: 3/4 passed"    ──stdout──→ 解析 → agent_logs              │

"FAIL: login.test.tsx"             │         WS push ──────────────→ "测试 3/4 通过，1 个失败"

                                    │                               │

"正在修复 login.test.tsx" ─stdout──→ 解析 → agent_logs              │

                                    │         WS push ──────────────→ "正在修复测试"

                                    │                               │

"测试全部通过"            ──stdout──→ 解析 → agent_logs              │

                                    │         WS push ──────────────→ "测试全部通过"

                                    │                               │

exit 0                              │ 检查: 退出码=0                 │

                                    │       worktree 有变更          │

                                    │                               │

                                    │ git commit + push + PR        │

                                    │ git worktree remove           │

                                    │ issue → in_review ────────────→ "PR 已提交，等待审核"

                                    └───────────────────────────────┘

```

**Worker 侧 stdout 流式捕获**：与旧方案一致，通过 `docker exec` 的 stdout pipe 实时捕获日志行，解析后写入 `agent_logs`、`issue_timeline`，并通过 WebSocket 推送给前端。

| 通信方向 | 通道 | 内容 | 频率 |

|---------|------|------|------|

| opencode → Worker | docker exec stdout | 实时日志行（生成/测试/修复） | 每秒数次 |

| Worker → MySQL | GORM Insert | `agent_logs`（结构存储）+ `issue_timeline`（展示存储） | 每条 stdout |

| Worker → Redis | ZADD | 最近 N 条日志缓存（可选，加速前端加载历史） | 每条 |

| Worker → 前端 | WebSocket | JSON 事件 `{type:"agent_log", text, ts}` | 每条 stdout |

| opencode → Worker | Exit Code | 0=成功, 非0=失败 | 结束时 1 次 |

### 7.1.2 执行控制（暂停 / 恢复 / 停止）

Worker 监听 Issue 控制命令，通过 Docker API 直接操作沙箱容器：

```go

// internal/sandbox/control.go — 进程级信号控制（非 ContainerPause，避免影响同项目其他 Issue）

func (w *Worker) PauseIssue(issueID)  // kill -STOP <opencode_pid> + 状态 in_progress→paused

func (w *Worker) ResumeIssue(issueID) // kill -CONT <opencode_pid> + 状态 paused→in_progress

func (w *Worker) StopIssue(issueID)   // kill -TERM <opencode_pid> + 状态 →backlog（worktree 保留可重试）

```

**前端控制按钮**（仅在 `in_progress` / `paused` 状态下显示）：

```

┌─ Issue #1 (in_progress) ────────────────────────────────────────┐

│  [⏸ 暂停]  [⏹ 停止]                                             │

│                                                                  │

│  时间线:                                                         │

│   12:02  agent   正在生成 src/login.tsx                           │

│   12:03  agent   正在生成 src/auth.ts                             │

│   ───────────────────────────────────────────                    │

│   [追加提示词: ________________________] [发送并重新执行]          │

└──────────────────────────────────────────────────────────────────┘

```

暂停后按钮变为：

```

┌─ Issue #1 (paused) ────────────────────────────────────────────┐

│  [▶ 恢复]  [⏹ 停止]                                             │

│                                                                  │

│  时间线:                                                         │

│   12:02  agent   正在生成 src/login.tsx                           │

│   12:05  system  执行已暂停（沙箱冻结中）                           │

└──────────────────────────────────────────────────────────────────┘

```

| 操作 | Docker API | Issue 状态 | 影响范围 | 数据保留 |

|------|-----------|-----------|---------|----------|

| **暂停** | `kill -STOP <opencode_pid>`（docker exec） | `paused` | 仅该 Issue 的 opencode 进程 | ✅ worktree 保留，内存保留 |

| **恢复** | `kill -CONT <opencode_pid>`（docker exec） | `in_progress` | 仅该 Issue | ✅ 从断点继续 |

| **停止** | `kill -TERM <opencode_pid>`（docker exec） | `backlog` | 仅该 Issue | ❌ 进程终止，worktree 保留（待重试复用） |

> **为什么不用 ContainerPause？** 容器为项目级常驻，`ContainerPause` 会冻结同项目内所有 Issue。改用进程级信号控制，只能暂停/恢复/停止单个 Issue 的 opencode 进程。

#### 暂停/恢复与 Asynq 任务生命周期

暂停/恢复操作直接操作 Docker 容器，但 Asynq 中的任务**已经出队**（dequeue），不再受 Redis 队列管理。为保证状态一致性，Worker 采用 **Redis Pub/Sub 事件驱动** 机制（替代 DB 轮询）：

```go

// internal/sandbox/control.go — 暂停时发布 Redis Pub/Sub 事件到 issue:control:{id} 频道

func (w *Worker) PauseIssue(issueID) { /* ... ContainerPause + Redis Publish("pause") */ }

```

**Worker 事件驱动状态监听**：Redis Pub/Sub 替代 DB 轮询，订阅 `issue:control:{id}` 频道，<100ms 响应（vs 旧方案 3-10s 延迟）：

```go

// internal/worker/executor.go

func (w *Worker) executeWithControl(ctx, issueID, runFn)  // select { Redis msg | done }

func (w *Worker) waitForResume(ctx, sub, issueID)         // 阻塞等待 "resume" 或 "stop" 事件

```

> **优势对比**：

>

> | 方案 | 响应延迟 | DB 压力 | Worker 资源占用 |

> |------|---------|---------|----------------|

> | DB 轮询（旧） | 3-10s | 每 Issue 3-10次/s | CPU 占用（循环轮询） |

> | Redis Pub/Sub（新） | <100ms | 0 | 零开销（channel 阻塞等待） |

**恢复/停止同样发布事件**：`ResumeIssue` → Publish("resume") | `StopIssue` → Publish("stop")，Worker goroutine 通过 `waitForResume` 接收。

**Worker 重启后恢复**：Worker 进程重启时，需恢复 in_progress 和 paused 状态的 Issue：

```

Worker 启动

    │

    ├── 1. RecoverRunningIssues — 恢复 in_progress

    │       │

    │       ▼

    │   扫描 issues WHERE status='in_progress'

    │       │

    │       ├── 查询 projects.sandbox_container_id  → 检查容器存活

    │       │   ├── 容器 running → 检查 /workspace/issue-{id} worktree 是否存在

    │       │   │   ├── 存在 → 检查 opencode 进程是否存活，是则重新 Attach stdout

    │       │   │   └── 不存在 → 重建 worktree 并重新入队

    │       │   ├── 容器 stopped → 重启容器（docker start），重新 Attach

    │       │   └── 容器不存在 → 重创项目容器 + git clone + 重建所有 worktree

    │       │

    │       └── 若 opencode 已退出 → 检查 exit code

    │           ├── exit 0 → commit + PR → in_review

    │           └── exit ≠0 → Issue → todo + 记录异常

    │

    └── 2. RecoverPausedIssues — 恢复 paused

            │

            ▼

        扫描 issues WHERE status='paused'

            │

            ├── opencode 进程存活（paused by SIGSTOP）→ 重新 Attach + 订阅控制频道

            └── 进程已不存在 → Issue → backlog + 记录异常

```

```go

// internal/worker/recovery.go

// RecoverRunningIssues: 遍历 in_progress → 检查项目容器存活 → worktree 存在则 Attach，否则重建入队

// RecoverPausedIssues:  遍历 paused → 检查容器存活 → Attach + 订阅控制频道 / 回退 backlog

//   恢复优先级：in_progress → paused；容器 ID 从 projects.sandbox_container_id 获取

```

> **设计要点**：

> - **恢复优先级**：先恢复 `in_progress`（可能需要继续执行），再恢复 `paused`（等待人工操作）

> - **容器不再按 Issue 追踪**：容器 ID 存在 `projects.sandbox_container_id`，不再存在 `issues.sandbox_container_id`

> - **容器保护**：`AutoRemove: false` 确保容器不会随 Docker 重启被自动清理

> - **暂停/恢复**依赖 opencode 进程 PID + SIGSTOP/SIGCONT 信号，容器重启后 PID 变化，需要重建 worktree 上下文

> - **issuess 表不再有 sandbox_container_id 字段**：改为通过 `project_id → projects.sandbox_container_id` 间接获取

### 7.2 安全策略

| 策略 | 配置 |

|------|------|

| 资源限制 | CPU 2核 / 内存 1GB / 磁盘 10GB |

| 执行超时 | 单任务最长 30 分钟，超时强制 SIGKILL |

| 网络隔离 | 仅允许出站到 GitHub API + LLM API 白名单 |

| 凭证注入 | GitHub Token 通过环境变量注入，不落盘 |

| 自动清理 | Issue done → git worktree remove + branch -D；容器常驻不销毁 |

| 非 root 运行 | User `sandbox` (uid 1000)，无 sudo 权限 |

| 镜像最小化 | Alpine 3.21 基础，预装 Node/Python/Git/opencode，不含 Go 运行时 |

### 7.2.1 数据持久化保障

服务器重启、Worker 进程崩溃等异常场景下的数据保护：

| 数据类型 | 存储位置 | 重启后 | 保障机制 |

|---------|---------|--------|----------|

| Issue / Agent / 配置 / 日志 | MySQL | ✅ 不丢失 | InnoDB WAL + binlog，ACID 事务保证 |

| 代码仓库（/workspace） | Docker Named Volume | ✅ 不丢失 | 生命周期独立于容器，Docker 重启后卷自动恢复 |

| 运行时数据（Skills/插件/配置） | 宿主机 bind mount | ✅ 不丢失 | 本地文件系统，与容器/进程生命周期解耦 |

| Asynq 未消费任务 | Redis（AOF） | ✅ 不丢失 | `appendonly yes` + `fsync everysec`，重启后自动重放 |

| Asynq 已出队未完成任务 | MySQL（Issue 状态） | ✅ 不丢失 | 任务出队时 Issue 已标记 `in_progress`，Worker 重启时 `RecoverRunningIssues` 兜底 |

| 容器内进程内存状态 | 容器内存 | ❌ 丢失 | 容器停止 → opencode 进程上下文丢失，回退到 `todo` 重试 |

**部署要求**（config.yaml 或部署文档注明）：

```yaml

# Redis 必须开启 AOF 持久化，否则 Asynq 未消费任务在 Redis 重启后丢失

redis:

  host: 127.0.0.1

  port: 6379

  # ↓ 部署时需在 redis.conf 中配置：

  # appendonly yes

  # appendfsync everysec

```

> **容器重启策略**：沙箱容器设置 `AutoRemove: false`，**不设置** `RestartPolicy`。原因：容器停止后由 Worker 的 `RecoverRunningIssues` 统一处理，避免容器自启动后无 Worker 监听导致 stdout 丢失。

**网络白名单实现**：

Docker 沙箱通过自定义网络 + 容器内 iptables 规则实现出站域名白名单：

```go

// internal/sandbox/network.go — 自定义 bridge 网络(anserflow-sandbox) + CapDrop NET_RAW + ALLOWED_DOMAINS 环境变量注入

func createSandboxNetwork(ctx, cli) (string, error)

func createSandboxContainer(ctx, cli, cfg) (string, error)  // entrypoint.sh 根据 ALLOWED_DOMAINS 配置 iptables

```

**容器内 iptables 脚本**（`entrypoint.sh`）：默认 DROP → 允许 lo/DNS/HTTPS → 解析 ALLOWED_DOMAINS 逐域名 ACCEPT

> **备选方案**：如果 iptables 方式在某些 CI 环境中不可用（如 Docker-in-Docker 受限），可退化为**无网络过滤**（仅依赖 Docker 默认 bridge 网络隔离 + 容器内不暴露任何端口）。生产环境建议在 K8s NetworkPolicy 层统一管理。

**Dockerfile** 位于 `docker/sandbox/Dockerfile`：

```bash

# 构建沙箱镜像

docker build -t anserflow/sandbox:latest -f docker/sandbox/Dockerfile .

```

> `docker/sandbox/.dockerignore` 已排除 `node_modules/`、`.git/`、`dist/` 等，确保构建上下文最小。

**Dockerfile**（`docker/sandbox/Dockerfile`）：

```dockerfile

FROM alpine:3.21

RUN apk add --no-cache \

    nodejs npm \

    python3 py3-pip \

    git \

    curl \

    bash \

    && rm -rf /var/cache/apk/*

RUN adduser -D -u 1000 sandbox

# opencode — 开源 AI 编码代理（npm 全局安装）

# https://github.com/anomalyco/opencode

RUN npm install -g opencode-ai@latest \

    && npm cache clean --force

RUN mkdir -p /workspace && chown sandbox:sandbox /workspace

WORKDIR /workspace

COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

USER sandbox

ENTRYPOINT ["/entrypoint.sh"]

```

```bash

# entrypoint.sh — 沙箱入口脚本

# 由 Worker 传入环境变量和配置注入文件控制行为：

#   GIT_REPO_URL / GIT_BRANCH / GITHUB_TOKEN → git clone

#   TASK_PROMPT                              → opencode run 的 prompt 参数

#

# opencode 配置由 Worker 在容器启动后、执行编码前通过以下方式注入（每次覆盖）：

#   ① 写入 ~/.config/opencode/config.json   → opencode 读取 provider / model 配置

#   ② 注入环境变量                           → API Key（如 OPENAI_API_KEY）不落盘

#   ③ opencode run --model provider/model    → 运行时指定模型

```

```

> **opencode 配置注入**：已由 `RuntimeAdapter` 接口替代（见 §六 RuntimeManager）。API Key AES-256 加密存储，Worker 解密后通过环境变量注入容器，不落盘。

> 预估镜像大小约 400MB。`.dockerignore` 排除 node_modules/ / .git/ / dist/ / .next/ / *.log。

### 7.3 Go Docker SDK

使用官方 `github.com/docker/docker/client`：

```go

// internal/sandbox/ — Docker SDK 核心操作

func ensureProjectContainer(ctx, project, runtime) (string, error)

//   复用已有容器 / 重启停止容器 / 创建新容器（1GB/2CPU/AutoRemove:false）

//   首次 git clone → /workspace/main（worktree 基准）

func createWorktree(ctx, containerID, issue) error   // git worktree add /workspace/issue-{id} -b feat/issue-{id}

func removeWorktree(ctx, containerID, issueID) error // git worktree remove + branch -D

func execOpenCode(ctx, containerID, issue, task) error // opencode run --workdir /workspace/issue-{id}

func destroyProjectSandbox(ctx, projectID, containerID) error // ContainerRemove + VolumeRemove

```

### 7.3.1 运行时数据目录 — 三层架构

```

宿主机目录结构（由 config.yaml sandbox.runtime_data_dir 指定根目录）

/var/lib/anserflow/                                ← runtime_data_dir

├── runtimes/                                      ← Layer 1: 全局模板（管理员维护）

│   └── opencode/                                  ← 对应 runtimes.name

│       ├── skills/                                ← 默认 Skills（如 anser-coder）

│       │   └── anser-coder/SKILL.md

│       ├── config.json                            ← 默认配置模板

│       └── plugins/                               ← 预装插件（如 opencode-agent-memory）

│

├── projects/                                      ← Layer 2: 项目实例

│   └── 42/                                        ← projects.id

│       └── runtime/                               ← projects.runtime_data_dir

│           ├── skills/                            ← 项目级 Skills（可增删）

│           ├── config.json                        ← 项目级配置（可自定义）

│           └── plugins/                           ← 项目级插件（可安装）

│

沙箱容器                                          ← Layer 3: bind mount

└── /home/sandbox/.opencode/                       ← RuntimeAdapter.HomeDir()

    ├── skills/                                    ← ← bind mount 自 projects/42/runtime/

    ├── config.json

    └── plugins/

```

```go

// internal/sandbox/runtime_init.go — 项目创建时从全局模板(runtimes/{name}) 递归复制到项目实例目录(projects/{id}/runtime)

//   幂等：已存在跳过；chown sandbox(uid=1000)；回写 projects.runtime_data_dir

func initProjectRuntime(ctx, projectID, runtimeName) (string, error)

```

**调用时机**：项目创建时（ProjectService.Create）或首次创建沙箱时（懒初始化）

```go

// 项目创建时调用

project, _ := projectRepo.Create(ctx, ...)

projectRuntimeDir, _ := initProjectRuntime(ctx, project.ID, defaultRuntimeName)

// 后续 createSandbox 使用 projectRuntimeDir

```

**清理**：

```go

// 在项目删除时清理 Named Volume + 运行时数据目录

func destroyWorkspaceVolume(ctx context.Context, projectID uint) error {

    cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

    cli.VolumeRemove(ctx, fmt.Sprintf("anserflow-workspace-%d", projectID), true)

    // 同时清理项目运行时数据目录

    dataDir := viper.GetString("sandbox.runtime_data_dir")

    projectDir := filepath.Join(dataDir, "projects", strconv.Itoa(int(projectID)))

    return os.RemoveAll(projectDir)

}

```

### 7.4 GitHub SDK 集成

Git 操作与 GitHub API 分别使用两套 Go SDK，职责分离：

| SDK | 用途 | 运行位置 | 认证 |

|------|------|----------|------|

| **go-git** (`github.com/go-git/go-git/v5`) | `clone` / `commit` / `push` / `checkout` | Docker 沙箱内（Worker） | HTTP Token / SSH 私钥 |

| **go-github** (`github.com/google/go-github/v68`) | 创建 PR / Issue / Review Comment / 读取仓库信息 | Go 后端（Service 层） | HTTP Personal Access Token |

```

┌─────────────────────────────────────────────────┐

│  go-github（GitHub REST API）                    │

│  ├── PullRequests.Create → 创建 PR              │

│  ├── Issues.Create / Edit → 操作 Issue          │

│  ├── PullRequests.CreateReview → 添加 Review    │

│  └── Repositories.GetContents → 读取文件树      │

│  运行位置：internal/service/（Gin HTTP 层）      │

├─────────────────────────────────────────────────┤

│  go-git（纯 Go Git 实现，无需系统 git 二进制）     │

│  ├── PlainClone → git clone                     │

│  ├── Worktree.Add + Commit → git add / commit   │

│  ├── Push → git push origin                     │

│  └── Checkout → git checkout -b                 │

│  运行位置：Docker 沙箱内（Asynq Worker）          │

└─────────────────────────────────────────────────┘

```

#### go-github 示例

```go

// go-github/v68 典型调用

client := github.NewClient(nil).WithAuthToken(token)

client.Issues.Create(ctx, owner, repo, &github.IssueRequest{...})

client.PullRequests.Create(ctx, owner, repo, &github.NewPullRequest{...})

```

#### go-git 认证示例

**go-git 认证**：HTTP Token（`http.BasicAuth{Username:"x-access-token", Password:ghp_xxx}`）或 SSH 私钥（`ssh.NewPublicKeys`）。commit→push：`Worktree.Add → Commit → Push`。

#### GitHub Token 权限要求

| 权限 | 用途 |

|------|------|

| `repo`（全部） | 读写仓库代码 |

| `issues:write` | 创建/更新 Issue |

| `pull_requests:write` | 创建 PR |

> Classic Token 授权全部仓库即可；Fine-grained Token 需额外指定具体仓库。

#### Go 模块依赖

```

# go.mod

github.com/google/go-github/v68 v68.0.0

github.com/go-git/go-git/v5 v5.15.0

```

#### 多平台扩展设计

后期支持 Gitea / GitLab / Gitee 等平台，通过 `GitPlatform` + `GitOps` 双接口抽象：

```

┌──────────────────────────────────────────────────────────┐

│  internal/git/                                            │

│  ├── manager.go           GitManager 统一入口             │

│  ├── platform.go          GitPlatform 接口定义             │

│  ├── ops.go               GitOps 接口 + Author 定义       │

│  ├── container_ops.go     容器 Shell 实现（当前）          │

│  ├── gogit_ops.go         go-git 库实现（Phase 2 可选）    │

│  ├── github.go            GitHub GitPlatform 实现          │

│  ├── gitea.go             Gitea 实现（Phase 2）            │

│  └── gitlab.go            GitLab 实现（Phase 2）           │

└──────────────────────────────────────────────────────────┘

```

```go

// internal/git/platform.go — 平台无关接口

type GitPlatform interface {

    CreateIssue(ctx, repo, title, body string, labels []string) (issueID string, err error)

    CreatePR(ctx, repo, title, head, base, body string) (prURL string, err error)

    GetRepoInfo(ctx, repo string) (*RepoInfo, error)

    ListRepos(ctx) ([]RepoInfo, error)

}

```

扩展策略：

| 层 | 扩展方式 | 说明 |

|----|---------|------|

| **GitPlatform** | 新增平台实现 | 每个 REST API SDK 封装为一个 `GitPlatform` 实现 |

| **GitOps** | 替换底层实现 | `ContainerGitOps`（Shell）→ `GoGitOps`（go-git 库），上层无需改动 |

| 平台 | Go SDK | 备注 |

|------|--------|------|

| **GitHub** | `go-github/v68` | 已实现 |

| **Gitea** | `code.gitea.io/sdk/gitea` | Phase 2 |

| **GitLab** | `github.com/xanzy/go-gitlab` | Phase 2 |

| **Gitee** | REST API | Phase 2 |

> 当前仅闭环 GitHub。仓库操作使用 `ContainerGitOps`（Shell），Phase 2 可选 `GoGitOps`（go-git）。

---

## 八、Skills 技能系统

### 8.1 两种导入方式

```

Skills 数据表：

┌──────────────────────────────────────────────┐

│ skills                                        │

├──────────────────────────────────────────────┤

│ source_type:  'manual' | 'zip'                │

│                                             │

│ 手动模式 (source_type='manual'):              │

│   definition: TEXT  ← 直接在 UI 编辑 Markdown │

│                                             │

│ ZIP 模式 (source_type='zip'):                 │

│   zip_hash: VARCHAR(64)   ← ZIP 包 SHA256    │

│   file_tree: JSON          ← 文件树快照       │

│   definition: TEXT         ← 解压后的 SKILL.md│

└──────────────────────────────────────────────┘

```

### 8.2 ZIP 包格式

```

my-skill.zip

├── SKILL.md              # 必须：Skill 定义

│                         #   ---

│                         #   name: my-skill

│                         #   description: ...

│                         #   ---

│                         #   # Skill 正文

├── agents/               # 可选：Agent 配置

│   └── openai.yaml

├── tools/                # 可选：辅助脚本

│   └── lint.sh

└── examples/             # 可选：示例

    └── sample.md

```

### 8.4 ZIP 导入完整流程

```

POST /api/orgs/:org_id/skills/import/zip    Content-Type: multipart/form-data

  → 内存解压 ZIP（≤10MB, MaxBytesReader）

  → 校验必须有 SKILL.md

  → 解析 frontmatter → 写入 skills 表（source_type="zip", zip_hash=SHA256）

```

> ZIP 全程内存处理，≤10MB。SHA256 去重可选（当前不自动去重）。

### 8.3 启用控制

```

                 ┌─────────────────┐

                 │  Skill A (全局)  │

                 │  enabled: true   │────────── 全局开关

                 └────────┬────────┘

                          │

          ┌───────────────┼───────────────┐

          │               │               │

    ┌─────┴─────┐   ┌─────┴─────┐   ┌─────┴─────┐

    │ Agent CEO │   │ Agent CTO │   │ Agent Dev │

    │ Skill A ✅ │   │ Skill A ✅ │   │ Skill A ❌ │── 单Agent开关

    └───────────┘   └───────────┘   └───────────┘

```

