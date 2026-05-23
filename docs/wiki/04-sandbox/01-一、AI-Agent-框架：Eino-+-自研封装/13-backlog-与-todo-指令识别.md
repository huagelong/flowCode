> 来源：`docs/plan/04-sandbox.md` 第 1109 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱 / Agent 基础设施](../README.md) -> [一、AI Agent 框架：Eino + 自研封装](README.md) -> `/backlog` 与 `/todo` 指令识别
> 相邻：[上一篇](12-anserAgent-Tool-系统（Skill-与系统通信）.md) · [下一篇](14-new-指令-—-会话上下文隔离.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### `/backlog` 与 `/todo` 指令识别

> **当前阶段入口**：Phase 1 仍使用显式命令 `/backlog`、`/todo`、`/new` 驱动。自然语言意图识别替代显式命令见 [06-agent.md](../../06-agent/README.md) §6.4，属于 [Phase 2](../../11-backlog/README.md) 能力。

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
