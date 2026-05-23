> 来源：`docs/plan/04-sandbox.md` 第 907 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱 / Agent 基础设施](../README.md) -> [一、AI Agent 框架：Eino + 自研封装](README.md) -> TokenManager — Token 配额管理器
> 相邻：[上一篇](10-GitManager-—-Git-管理器.md) · [下一篇](12-anserAgent-Tool-系统（Skill-与系统通信）.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### TokenManager — Token 配额管理器

在现有 `TokenTracker` 基础上升级，增加配额检查、周期归档、用量报告能力。

**设计原则**：

- 配额检查：`manager.CheckQuota(orgID)` → 超额时暂停 Agent 执行

- 周期归档：每天凌晨 Redis → MySQL 持久化

- 用量报告：按组织/项目/Agent/日期多维度查询

**实现**：

```go

type Usage struct {

    PromptTokens     int64

    CompletionTokens int64

    Source           string // "agent" | "anseragent"

}

type TokenManager struct {

    redis  *redis.Client

    quota  QuotaService

}

func (m *TokenManager) Record(ctx context.Context, agentID uint, usage *Usage)          // Redis Hash 按 Agent+日期聚合

func (m *TokenManager) CheckQuota(ctx context.Context, orgID uint) bool                  // 检查组织月度配额

func (m *TokenManager) GetDailyUsage(ctx context.Context, agentID uint, date string) (*DailyReport, error)

func (m *TokenManager) Archive(ctx context.Context, date string) error                  // 定时归档到 MySQL

```

> 用量以 Redis Hash 聚合（key=`tokens:agent:{id}:date:{date}`），TTL 30天；定期 Archive 写入 `token_usage` 表。

**配额策略**：

| 级别 | 周期 | 默认配额 | 超限行为 |
|------|------|---------|---------|
| 组织 | 月度 | 500M tokens | 所有 Agent 暂停执行，群聊通知管理员 |
| 项目 | 月度 | 100M tokens | 该项目 Agent 暂停，群聊通知项目成员 |
| Agent | 单次执行 | 2M tokens | 当前执行中断，Issue 退回 todo |

**超额告警与拦截**：

```
     配额消耗
        │
  ┌─────┼─────┐
  │     │     │
  ▼     ▼     ▼
 80%   90%   100%
  │     │     │
  ▼     ▼     ▼
群聊   邮件   硬阻断
告警   告警   Agent 停止
(黄色) (橙色) (红色)
```

| 阈值 | 动作 | 通知渠道 |
|------|------|---------|
| 月度用量达 80% | 发送用量提醒 | 群聊系统消息 |
| 月度用量达 90% | 发送紧急提醒 + 建议升级配额 | 群聊 + 邮件 |
| 月度用量达 100% | 暂停所有 Agent 执行，等待管理员确认 | 群聊 + 邮件 + 管理后台红色告警 |
| 单次执行超 2M | 中断当前执行，记录超限日志 | 群聊 + agent_logs |

**实现**：

```go
type QuotaLevel int

const (
    QuotaOrg   QuotaLevel = iota  // 组织级
    QuotaProject                   // 项目级
    QuotaAgent                     // Agent 执行级
)

type QuotaConfig struct {
    OrgMonthlyLimit    int64  // 默认 500_000_000
    ProjectMonthlyLimit int64 // 默认 100_000_000
    AgentExecutionLimit int64 // 默认 2_000_000
}

func (m *TokenManager) CheckQuota(ctx context.Context, orgID, projectID uint, estimatedTokens int64) (allowed bool, reason string) {
    // ① 检查组织月度配额
    if orgUsage := m.getOrgMonthlyUsage(orgID); orgUsage >= m.config.OrgMonthlyLimit {
        return false, "组织月度配额已耗尽"
    }
    // ② 检查项目月度配额
    if projUsage := m.getProjectMonthlyUsage(projectID); projUsage >= m.config.ProjectMonthlyLimit {
        return false, "项目月度配额已耗尽"
    }
    // ③ 估算单次执行是否超限
    if estimatedTokens >= m.config.AgentExecutionLimit {
        return false, "预估用量超过单次执行上限"
    }
    return true, ""
}

func (m *TokenManager) NotifyQuotaAlert(ctx context.Context, level string, usage, limit int64, channels []string) {
    // level: "warning"(80%) | "critical"(90%) | "blocked"(100%)
    // channels: ["group_chat", "email", "admin_banner"]
}
```
