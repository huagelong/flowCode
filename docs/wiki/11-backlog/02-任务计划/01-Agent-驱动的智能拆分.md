> 来源：`docs/plan/11-backlog.md` 第 66 行
> 位置：[总目录](../../README.md) -> [AnserFlow — 远期 Backlog（Phase 2+）](../README.md) -> [任务计划](README.md) -> Agent 驱动的智能拆分
> 相邻：[上一篇](README.md) · [下一篇](02-任务依赖图可视化.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### Agent 驱动的智能拆分

利用 anserAgent 对需求做语义级拆解，自动推断子任务、优先级和依赖关系：

```
需求: "做一个用户登录页"
        │
        ▼
  anserAgent 拆分（分析需求语义）
        │
        ▼
  ┌─────────────────────────────────────────┐
  │ T01  登录表单 UI        前端  2h  p1     │
  │ T02  JWT 认证 API       后端  3h  p0     │
  │ T03  bcrypt 密码加密    后端  1h  p0  ←── T02 依赖 T03
  │ T04  登录态持久化       前端  1h  p1  ←── T04 依赖 T01+T02
  └─────────────────────────────────────────┘
```

```go
// internal/agent/backlog_breakdown.go
type BreakdownResult struct {
    Tasks []BreakdownTask `json:"tasks"`
}
type BreakdownTask struct {
    Title          string   `json:"title"`
    Description    string   `json:"description"`
    EstimatedHours float64  `json:"estimated_hours"`
    RoleLabel      string   `json:"role_label"`
    Priority       string   `json:"priority"`
    DependsOn      []int    `json:"depends_on"`
    Acceptance     string   `json:"acceptance"`
}
```
