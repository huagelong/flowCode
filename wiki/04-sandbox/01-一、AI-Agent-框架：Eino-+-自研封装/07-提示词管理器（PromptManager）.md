> 来源：`docs/plan/04-sandbox.md` 第 261 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱 / Agent 基础设施](../README.md) -> [一、AI Agent 框架：Eino + 自研封装](README.md) -> 提示词管理器（PromptManager）
> 相邻：[上一篇](06-Skill-两层继承（沙箱执行时）.md) · [下一篇](08-IssueStatusManager-—-Issue-状态机管理器.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

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
