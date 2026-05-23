> ???`docs/plan/06-agent.md` ? 53 ?
> ???[???](../README.md) -> [AnserFlow - anserAgent 智能体系统](README.md) -> 二、设计基础 — Agent 统一接口
> ???[???](01-一、系统定位.md) ? [???](03-三、目录结构.md)
> ?????[??????](README.md) ? [一、系统定位](01-一、系统定位.md) ? [三、目录结构](03-三、目录结构.md)

## 二、设计基础 — Agent 统一接口

参照 Eino ADK 的 `Agent` 接口设计，anserFlow 中所有 Agent 类型（ChatModel Agent / Workflow Agent / Custom Agent）都实现同一个接口：

```go
// core/interface.go

type Agent interface {
    // Name 返回 Agent 名称，用于日志和 Agent 间发现
    Name(ctx context.Context) string

    // Description 返回 Agent 描述，用于 Supervisor 模式下的任务路由
    Description(ctx context.Context) string

    // Run 执行 Agent，返回异步事件迭代器
    // 每个事件可以是推理输出（Output）、工具调用（Action）或中断（Interrupted）
    Run(ctx context.Context, input *AgentInput) *AsyncIterator[*AgentEvent]
}
```

**核心优势**：
- 编排器（Runner / Orchestrator）统一调度任意 Agent，无需类型断言
- Workflow Agent 的子 Agent 可以是任意类型，支持嵌套组合
- 中断/恢复机制在接口层统一注入，所有 Agent 自动获得可中断能力

---
