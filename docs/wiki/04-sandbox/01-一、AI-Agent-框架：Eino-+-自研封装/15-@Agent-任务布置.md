> 来源：`docs/plan/04-sandbox.md` 第 1193 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱 / Agent 基础设施](../README.md) -> [一、AI Agent 框架：Eino + 自研封装](README.md) -> @Agent 任务布置
> 相邻：[上一篇](14-new-指令-—-会话上下文隔离.md) · [下一篇](16-Token-用量与成本追踪.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

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

Eino 在将人工提示词注入 anserAgent 之前，自动进行上下文增强与工程化改写：

```go

// internal/agent/prompt_optimizer.go — Eino 调用 LLM 改写用户提示词

// ① 收集 Issue 上下文 + 时间线

// ② 解析 @Issue #N 引用，合并关联 Issue 内容

// ③ LLM 将用户反馈改写为精确编码指令

func (o *PromptOptimizer) Enhance(ctx context.Context, rawPrompt string, issue *model.Issue) (string, error)

```

> **关键**：Eino 只做调度与提示词优化，不进入 Docker 沙箱。沙箱内的代码生成完全由 anserAgent 完成。
