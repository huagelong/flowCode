> 来源：`docs/plan/04-sandbox.md` 第 1361 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱 / Agent 基础设施](../README.md) -> 二、Docker 沙箱方案
> 相邻：[上一篇](../01-一、AI-Agent-框架：Eino-+-自研封装/README.md) · [下一篇](../03-三、Skills-技能系统/README.md)
> 相关主题：[返回文档入口](../README.md) · [一、AI Agent 框架：Eino + 自研封装](../01-一、AI-Agent-框架：Eino-+-自研封装/README.md) · [三、Skills 技能系统](../03-三、Skills-技能系统/README.md)

## 二、Docker 沙箱方案

> 📎 沙箱运行时接口定义（SandboxManager / RuntimeManager 适配器模式）已迁至 [04b-sandbox-runtime.md](../../04b-sandbox-runtime/README.md)。本节保留架构设计与执行流程。

## 子章节导航

- [2.0 容器与代码隔离策略](01-2.0-容器与代码隔离策略.md)
- [2.1 执行流程](02-2.1-执行流程.md)
- [2.1.1 沙箱 ↔ 系统消息互通闭环](03-2.1.1-沙箱-↔-系统消息互通闭环.md)
- [2.1.2 执行控制（暂停 / 恢复 / 停止）](04-2.1.2-执行控制（暂停-恢复-停止）.md)
- [2.2 安全策略](05-2.2-安全策略.md)
- [2.2.1 数据持久化保障](06-2.2.1-数据持久化保障.md)
- [2.3.1 运行时数据目录 — 三层架构](07-2.3.1-运行时数据目录-—-三层架构.md)
- [2.4 GitHub SDK 集成](08-2.4-GitHub-SDK-集成.md)
