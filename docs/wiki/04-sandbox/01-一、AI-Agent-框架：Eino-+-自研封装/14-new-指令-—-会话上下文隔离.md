> 来源：`docs/plan/04-sandbox.md` 第 1157 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱 / Agent 基础设施](../README.md) -> [一、AI Agent 框架：Eino + 自研封装](README.md) -> `/new` 指令 — 会话上下文隔离
> 相邻：[上一篇](13-backlog-与-todo-指令识别.md) · [下一篇](15-@Agent-任务布置.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

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
