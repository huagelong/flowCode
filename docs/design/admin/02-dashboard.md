# Admin Dashboard 仪表盘

> 路由：`/admin/dashboard`
> API：`GET /api/orgs/:org_id/dashboard`

---

## 1. 页面概览

仪表盘为 Admin 首页，展示当前组织的关键指标和最近活动。

```
┌──────────────────────────────────────────────────────────────────┐
│ Dashboard                                              导出报告 │
│                                                                  │
│ ┌────────────────┐ ┌────────────────┐ ┌────────────────┐        │
│ │ Issue 分布      │ │ Agent 活跃度    │ │ 项目概览        │        │
│ │                │ │                │ │                │        │
│ │ [环形图]       │ │ [柱状图 7天]   │ │ 数字 + 进度条   │        │
│ │                │ │                │ │                │        │
│ │ done_rate: 68% │ │ 活跃 3/5      │ │ 12 项目        │        │
│ └────────────────┘ └────────────────┘ └────────────────┘        │
│                                                                  │
│ ┌────────────────────────────────────────────────────────────┐  │
│ │ 最近活动 (7天)                                      [查看全部] │  │
│ │                                                            │  │
│ │  12:30  🤖 Agent "前端" 完成了 Issue #42                   │  │
│ │  11:45  👤 张三 创建了 Issue #41                            │  │
│ │  10:20  🤖 Agent "后端" 开始执行 Issue #39                  │  │
│ │  09:00  👤 李四 加入了组织                                  │  │
│ │  ...                                                       │  │
│ └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│ ┌────────────────────────────────────────────────────────────┐  │
│ │ Token 用量趋势 (7天)                         [查看详细]     │  │
│ │ [面积图]                                                   │  │
│ └────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

---

## 2. 模块详细设计

### 2.1 Issue 分布卡片

**组件**: Card + Recharts PieChart (环形图)

```
┌────────────────────────┐
│ Issue 分布              │
│                        │
│      ╭───╮             │
│     │ 68% │            │
│      ╰───╯             │
│   ┌─ ─ ─ ─ ─ ┐        │
│                        │
│ ● backlog  3           │
│ ● todo     5           │
│ ● 进行中   2           │
│ ● 审核中   1           │
│ ● 已完成  12           │
│ ──────────────         │
│ 总计 23  完成率 68%    │
└────────────────────────┘
```

**数据字段**:
- `issue_distribution`: `{ backlog, todo, in_progress, in_review, done, total, done_rate }`

**样式**:
- 卡片：`bg-card rounded-lg border shadow-sm p-6`
- 环形图：内径 60%，外径 80%，中间显示完成率百分比
- 各状态色见 [01-design-tokens.md](../01-design-tokens.md) Issue 状态色
- 图例：`text-xs` + 圆点色块 `h-2 w-2 rounded-full`
- 完成率：`text-2xl font-bold`

### 2.2 Agent 活跃度卡片

**组件**: Card + Recharts BarChart

```
┌────────────────────────┐
│ Agent 活跃度            │
│                        │
│  ▂ ▄ ▅ ▆ ▇ █ ▅        │
│  M T W T F S S         │
│                        │
│ 活跃 3/5  │ 总执行 47次│
└────────────────────────┘
```

**数据字段**:
- `agent_activity`: `{ active_agents, total_agents, executions_by_day: [{date, count}] }`

**样式**:
- 柱状图：最近 7 天，柱体 `fill=hsl(var(--primary))`，`radius=[4,4,0,0]`
- 柱宽 `barSize=24`，间距 `barGap=4`
- X 轴：日期缩写 (Mon, Tue...)
- 底部统计：`text-sm text-muted-foreground`，活跃数/总数

### 2.3 项目概览卡片

**组件**: Card + 数字统计

```
┌────────────────────────┐
│ 项目概览                │
│                        │
│  12                    │
│  总项目数               │
│                        │
│  8 活跃  4 归档         │
│                        │
│  Issue 完成进度         │
│  ████████░░░░  68%     │
│                        │
│  本周新增 Issue: 15     │
│  本周完成 Issue: 9      │
└────────────────────────┘
```

**数据字段**:
- `project_overview`: `{ total_projects, active_projects, total_issues, completion_rate, weekly_new, weekly_done }`

**样式**:
- 主数字：`text-3xl font-bold`
- 标签：`text-sm text-muted-foreground`
- 进度条：shadcn/ui `Progress` 组件，高度 `h-2`
- 统计项：`flex justify-between text-sm`

### 2.4 最近活动列表

**组件**: Card + Timeline 式列表

```
┌────────────────────────────────────────────────────────────┐
│ 最近活动 (7天)                                   [查看全部] │
│                                                            │
│ ● 12:30  🤖 "前端Agent" 完成了 Issue #42           [查看→] │
│ ● 11:45  👤 张三 创建了 Issue #41 "修复登录Bug"     [查看→] │
│ ● 10:20  🤖 "后端Agent" 开始执行 Issue #39          [查看→] │
│ ● 09:00  👤 李四 加入了组织                           —    │
│ ● 昨天 22:15  🤖 "CTO" 生成了方案 (backlog)          [查看→] │
│ ...                                                        │
└────────────────────────────────────────────────────────────┘
```

**数据字段**:
- `recent_activity`: `Array<{ time, actor_type, actor_name, event_type, description, resource_id }>`

**样式**:
- 每项：`flex items-start gap-3 py-2.5 border-b last:border-0`
- 时间：`text-xs text-muted-foreground w-16 shrink-0`
- 图标：`h-4 w-4`，Agent 用 `Bot`，用户用 `User`
- 事件描述：`text-sm`
- 链接：`text-xs text-primary hover:underline`
- 最多显示 10 条，超出折叠

### 2.5 Token 用量趋势

**组件**: Card + Recharts AreaChart

```
┌────────────────────────────────────────────────────────────┐
│ Token 用量趋势 (7天)                          [查看详细]    │
│                                                            │
│  50k ┤                                          ╭──         │
│      │                              ╭──╮    ╭──╯           │
│  25k ┤           ╭──╮         ╭──╯    ╰──╯               │
│      │   ╭──╮ ╭╯    ╰──╮  ╭╯                                 │
│   0  └──╯  ╰─╯          ╰──╯                               │
│      Mon  Tue  Wed  Thu  Fri  Sat  Sun                      │
│                                                            │
│  本周总计: 156,420 tokens  估算费用: $2.34                  │
│  输入: 98,200  输出: 58,220                                 │
└────────────────────────────────────────────────────────────┘
```

**数据字段**:
- `token_summary`: `{ period, total_input, total_output, total_tokens, estimated_cost, daily: [{date, input, output}] }`

**样式**:
- 面积图：`fill=hsl(var(--primary)/0.1)`, `stroke=hsl(var(--primary))`
- X 轴：日期，Y 轴：Token 数
- 底部统计：`text-sm`，费用 `font-medium`

---

## 3. 交互说明

| 操作 | 行为 |
|------|------|
| 切换组织 | 顶部组织选择器切换 → 所有卡片数据刷新 (TanStack Query `invalidateQueries`) |
| 点击"查看全部" | 跳转到对应管理页面 (活动 → Issues 页) |
| 点击"查看详细" | 跳转到 Token 用量页 (`/admin/agents/:id`) |
| 点击活动项"查看→" | 跳转到对应 Issue/Agent 详情页 |
| 导出报告 | 触发 CSV 下载（当前组织 7 天数据） |

---

## 4. 数据刷新策略

- 初始加载：TanStack Query `staleTime: 30000` (30s)
- 后台刷新：`refetchInterval: 60000` (60s)
- WebSocket 推送触发：收到 `status_change` 事件时 `invalidateQueries`
- 手动刷新：页面无刷新按钮，依赖自动刷新

---

## 5. 空状态

新组织无数据时：

```
┌──────────────────────────────────────────────────────────┐
│                                                          │
│              [📊 图标 64px]                              │
│                                                          │
│           欢迎使用 AnserFlow                              │
│                                                          │
│    还没有数据。创建项目和 Agent 开始协作吧！                │
│                                                          │
│    [创建项目]  [创建 Agent]                               │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

---

## 6. 响应式

| 断点 | 布局 |
|------|------|
| lg (1024+) | 3 列网格统计卡 + 全宽列表 |
| xl (1280+) | 统计卡不变，Token 图表可并排 |
