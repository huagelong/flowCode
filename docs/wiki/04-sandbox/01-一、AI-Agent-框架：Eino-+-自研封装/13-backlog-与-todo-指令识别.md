> 来源：`docs/plan/04-sandbox.md` 第 1109 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱 / Agent 基础设施](../README.md) -> [一、AI Agent 框架：Eino + 自研封装](README.md) -> `/backlog` 与 `/todo` 指令识别
> 相邻：[上一篇](12-anserAgent-Tool-系统（Skill-与系统通信）.md) · [下一篇](14-new-指令-—-会话上下文隔离.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### `/backlog` 与 `/todo` 指令识别

> **当前阶段入口**：Phase 1 仍使用显式命令 `/backlog`、`/todo`、`/new` 驱动。自然语言意图识别替代显式命令见 [06-agent.md](../../06-agent/README.md) §6.4，属于 [Phase 2](../../11-backlog/README.md) 能力。

anserAgent 在群聊中监听 WebSocket 消息，检测到 `/backlog` 或 `/todo` 指令时触发拆解流程：`/backlog` 创建需求 Issue，`/todo` 基于当前 backlog 需求创建一组 `parent_id` 指向该需求的任务 Issue。

```go

// internal/agent/command_handler.go

type CommandHandler struct {

    orchestrator *GroupOrchestrator

    parser       *BacklogParser

}

// HandleBacklog 创建 backlog 需求；HandleTodo 为指定 backlog 创建 todo 子 Issue

func (h *CommandHandler) HandleBacklog(msg *ws.Message) {

    // ① 解析指令文本 + 收集群聊上下文（最近 50 条）

    // ② 输入校验：非空 + 上下文 ≥3 条

    // ③ anserAgent 产出需求描述 → parser 创建 backlog Issue

    // ④ 写 DB + WS 广播结果

}

```

**`/todo` vs `/backlog` 对比**：

| 维度 | `/backlog` | `/todo` |

|------|-----------|--------|

| Agent 参与 | ✅ anserAgent 编排产出需求 | ✅ anserAgent 基于 backlog 分析任务 |

| Issue 状态 | 创建需求 Issue（`backlog`） | 创建子任务 Issue（`todo`） |

| 确认要求 | 无需确认 | 无需确认，直接可执行 |

| 适用场景 | 记录原始需求与讨论上下文 | 从需求分析出可执行任务列表 |
