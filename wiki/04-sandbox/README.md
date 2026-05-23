> 来源：`docs/plan/04-sandbox.md` 第 1 行
> 位置：[总目录](../README.md) -> AnserFlow - 沙箱 / Agent 基础设施
> 相邻：[上一篇](../03-client/README.md) · [下一篇](../04b-sandbox-runtime/README.md)
> 相关主题：[上一份源文档](../03-client/README.md) · [下一份源文档](../04b-sandbox-runtime/README.md)

﻿# AnserFlow - 沙箱 / Agent 基础设施

> **职责边界**：本文档覆盖 Agent 基础设施层（Eino 框架、状态机、通知、Git、Token 追踪、Skill 导入）。Agent 大脑设计（五层记忆、Skill 自改进、调度编排）见 [06-agent.md](../06-agent/README.md)。沙箱执行运行时（SandboxManager / RuntimeManager）见 [04b-sandbox-runtime.md](../04b-sandbox-runtime/README.md)。
>
> 参考代码映射见 [07-architecture.md](../07-architecture/README.md) §建议保留的模块 / 建议废弃的模块 / 建议做映射迁移的模块。
>
> **实现代码**：本文档中的实现级代码示例已外提至 [sandbox-code-examples.md](../../reference/sandbox-code-examples.md)，文中通过链接引用。

---

## 章节导航

- [一、AI Agent 框架：Eino + 自研封装](01-一、AI-Agent-框架：Eino-+-自研封装/README.md)
- [二、Docker 沙箱方案](02-二、Docker-沙箱方案/README.md)
- [三、Skills 技能系统](03-三、Skills-技能系统/README.md)
