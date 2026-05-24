> 来源：`docs/plan/04-sandbox.md` 第 471 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱 / Agent 基础设施](../README.md) -> [一、AI Agent 框架：Eino + 自研封装](README.md) -> IssueStatusManager — Issue 状态机管理器
> 相邻：[上一篇](07-提示词管理器（PromptManager）.md) · [下一篇](09-NotificationChannelManager-—-通知渠道管理器.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

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

| `todo` | `in_progress` | Scheduler 分配 | 群聊通知 + 时间线 + 通知被分配人 |

| `in_progress` | `in_review` | 编码完成，Agent 推送代码并创建 PR | 群聊通知 + 时间线 + PR 链接 |

| `in_progress` | `paused` | 人工暂停 | 群聊通知 + 时间线 |

| `in_progress` | `todo` | Worker 执行失败 | 群聊通知 + 时间线 + retry_count++ |

| `in_review` | `done` | GitHub Webhook 通知 PR 已 merge | 群聊通知 + 时间线 + worktree 清理 |

| `in_review` | `todo` | PR 被拒绝/关闭未合并 | 群聊通知 + 时间线（worktree 保留待重试） |

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

    m.allow("todo", "in_progress")

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

// 任务列表生成不是 backlog → todo 状态流转，而是创建 parent_id 指向 backlog 的 todo 子 Issue。
// 所有含群聊通知的运行态流转都注册群聊 Hook。

statusMgr.OnTransition("todo", "in_progress",

    func(ctx context.Context, issueID uint, from, to string) error {

        issue, _ := issueRepo.FindByID(ctx, issueID)

        if issue.SourceGroupID != 0 {

            msgService.SendSystemMessage(ctx, issue.SourceGroupID,

                prompts.Get("system.issue.start", issueID))

        }

        return nil

    },

)

// 业务代码调用：一行代替多处 if/通知/时间线

statusMgr.Transition(ctx, issueID, "todo", "in_progress")

```
