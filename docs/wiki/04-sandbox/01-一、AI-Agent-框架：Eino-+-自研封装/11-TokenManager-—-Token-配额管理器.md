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
