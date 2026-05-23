> 来源：`docs/plan/06-agent.md` 第 1393 行
> 位置：[总目录](../README.md) -> [AnserFlow - anserAgent 智能体系统](README.md) -> 十三、从 eino-* Skills 到 anserAgent 的迁移映射
> 相邻：[上一篇](12-十二、数据库表/README.md) · [下一篇](14-十四、当前阶段范围.md)
> 相关主题：[返回文档入口](README.md) · [十二、数据库表](12-十二、数据库表/README.md) · [十四、当前阶段范围](14-十四、当前阶段范围.md)

## 十三、从 eino-* Skills 到 anserAgent 的迁移映射

原 eino-* Skills 的功能全部由 anserAgent + Eino ADK 接管：

| 原 Skill | 迁移到 | 说明 |
|----------|--------|------|
| ~~eino-discuss~~ | **L0 元规则** | 讨论行为约束内化为元规则 |
| ~~eino-backlog~~ | **L3 Skill** | 方案拆解作为可自改进的 SOP（Agent 自动识别群聊意图触发） |
| ~~eino-optimizer~~ | **SkillImprover** | 提示词优化逻辑由自改进引擎接管 |
| ~~eino-planner~~ | **L3 Skill** | 任务编排 SOP，可自动生成 |
| ~~eino-* 硬编码编排~~ | **Workflow Agents + Runner** | 编排逻辑交由 Eino ADK Sequential/Parallel/Loop Agent + Runner 统一管理 |
---
