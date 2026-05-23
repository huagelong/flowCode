> ???`docs/plan/04-sandbox.md` ? 261 ?
> ???[???](../../README.md) -> [AnserFlow - 沙箱 / Agent 基础设施](../README.md) -> [一、AI Agent 框架：Eino + 自研封装](README.md) -> 提示词管理器（PromptManager）
> ???[???](06-Skill-两层继承（沙箱执行时）.md) ? [???](../02-触发条件.md)
> ?????[??????](README.md) ? [??????](../README.md)

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

**实现代码**: [sandbox-code-examples.md §PromptManager](../../../reference/sandbox-code-examples.md#tool--skill-抽象)

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

    defaultManager.prompts["system.issue.pr_submitted"] = `Issue #%d 编码完成，等待审批`

    defaultManager.prompts["system.issue.failed"] = `Issue #%d 执行失败: %s`

    defaultManager.prompts["system.issue.done"] = `Issue #%d 已完成 ✅`

    defaultManager.prompts["system.issue.pr_rejected"] = `Issue #%d 审批未通过，已退回 todo`

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

> **Phase 1 基线**：当前状态机以 [02-api.md](../../02-api/README.md) 定义的显式命令流 + GitHub PR 审核流为准。群聊审批/HITL 仅作为 [Phase 2](../../11-backlog/README.md) 预研能力，不纳入本节主状态流转。

**设计原则**：

- 状态流转合法表集中定义，不散落在 if/switch 中

- 每次流转的副作用（群聊通知、时间线、通知推送）统一 Hook 回调

- 新增状态或修改流转规则只需改一处

**状态流转合法表**：

| from | to | 触发方 | 副作用 |

|------|----|--------|--------|

| `backlog` | `todo` | 人工确认 | 群聊通知 + 时间线 |

| `todo` | `in_progress` | Scheduler 分配 | 群聊通知 + 时间线 + 通知被分配人 |

| `in_progress` | `in_review` | 编码完成，Agent 推送代码并创建 PR | 群聊通知 + 时间线 + PR 链接 |

| `in_progress` | `paused` | 人工暂停 | 群聊通知 + 时间线 |

| `in_progress` | `todo` | Worker 执行失败 | 群聊通知 + 时间线 + retry_count++ |

| `in_review` | `done` | GitHub Webhook 通知 PR 已 merge | 群聊通知 + 时间线 + worktree 清理 |

| `in_review` | `todo` | PR 被拒绝/关闭未合并 | 群聊通知 + 时间线（worktree 保留待重试） |

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

- Phase 2 可替换为 go-git 库实现（详见 [11-backlog.md](../../11-backlog/README.md)），上层无感知

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

> Phase 2 可选替换为 `GoGitOps`（go-git 库实现，详见 [11-backlog.md](../../11-backlog/README.md)），上层通过 `GitOps` 接口无感知切换。

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

    Source           string // "agent" | "anseragent"

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
