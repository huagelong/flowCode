> 来源：`docs/plan/04-sandbox.md` 第 243 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱 / Agent 基础设施](../README.md) -> [一、AI Agent 框架：Eino + 自研封装](README.md) -> Skill 两层继承（沙箱执行时）
> 相邻：[上一篇](05-Tool-Skill-抽象.md) · [下一篇](07-提示词管理器（PromptManager）.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### Skill 两层继承（沙箱执行时）

Worker 通过 RuntimeClient 向沙箱注入 Skills 配置时，合并 Runtime 默认 + Agent 独立绑定，Agent 可覆盖关闭 Runtime 继承的 Skill。Skills 以 JSON Lines 消息随任务一并发送给沙箱内运行的工具：

**实现代码**: [sandbox-code-examples.md §Skill 两层继承](../../../reference/sandbox-code-examples.md#skill-两层继承)

**Skill 注入规则**：

| Skill | 来源 | 能否关闭 | 说明 |

|-------|------|---------|------|

| `anser-coder` | Runtime 默认（anserAgent） | ❌ 不可关闭 | `is_builtin=1`，前端灰掉开关 |

| 用户创建的 Skill | Runtime 默认 / Agent 绑定 | ✅ 可开关 | 后台自由管理 |

| Agent 主动关闭 Runtime Skill | Agent 级覆盖 | ✅ | `agent_skills.enabled=false` 覆盖 Runtime 默认 |
