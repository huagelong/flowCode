# Notification / EmptyState / Sidebar / FormBuilder / AgentLog / TokenUsage / Pagination

---

## Notification (通知)

> 基于 Sonner toast + 自定义 WebSocket 推送

### Sonner 配置

```tsx
<Toaster
  position="bottom-right"
  toastOptions={{
    duration: 5000,
    className: "bg-card border shadow-lg",
  }}
/>
```

### Toast 类型

| 类型 | 图标 | 持续时间 | 用途 |
|------|------|----------|------|
| `toast.success()` | `CheckCircle2` | 3s | 操作成功 |
| `toast.error()` | `XCircle` | 5s | 操作失败 |
| `toast.info()` | `Info` | 4s | 信息通知 |
| `toast.warning()` | `AlertTriangle` | 5s | 警告 |
| `toast.promise()` | Spinner → 上述 | 直到 resolve | 异步操作 |

### Toast 样式

```
┌──────────────────────────────────────────┐
│ ✅ 操作成功                               │
│ Agent "前端Agent" 已创建                  │
│                              [查看 →]    │
└──────────────────────────────────────────┘
```

- 最大宽度：`max-w-sm` (384px)
- 圆角：`rounded-lg`
- 阴影：`shadow-lg`
- 关闭按钮：右上角
- 可选 action 按钮

### 浏览器推送通知

- `Notification.requestPermission()` 获取权限
- 收到 WebSocket `native_notification` 事件时触发
- 图标使用 AnserFlow logo

---

## EmptyState (空状态)

### 标准空状态

```
        [图标 64px]
     标题文字
 描述文字说明当前状态
 [可选操作按钮]
```

### 样式

| 元素 | 样式 |
|------|------|
| 容器 | `flex flex-col items-center justify-center py-12` |
| 图标 | `h-16 w-16 text-muted-foreground/50` |
| 标题 | `text-lg font-medium mt-4` |
| 描述 | `text-sm text-muted-foreground mt-1 text-center max-w-[300px]` |
| 按钮 | `mt-4` |

### 各页面空状态

| 页面 | 图标 | 标题 | 描述 | 按钮 |
|------|------|------|------|------|
| Dashboard | `LayoutDashboard` | 欢迎使用 AnserFlow | 创建项目和 Agent 开始协作 | [创建项目] [创建 Agent] |
| Agent 列表 | `Bot` | 还没有 Agent | 创建你的第一个 AI Agent | [创建 Agent] |
| 项目列表 | `FolderKanban` | 还没有项目 | 创建项目开始管理 Issue | [创建项目] |
| Issue (某状态) | `CircleDot` | 暂无 backlog Issue | 通过 /backlog 命令或手动创建 | [创建 Issue] |
| Skill 列表 | `Zap` | 还没有技能 | 创建或导入 Agent 技能 | [创建 Skill] |
| 会话列表 | `MessageSquare` | 开始新对话 | 搜索用户或 Agent 开始对话 | — |
| 聊天窗口 | `MessageSquare` | 开始对话 | 输入消息或使用命令 | — |
| 通知 | `Bell` | 没有新通知 | 有新消息时这里会显示 | — |

---

## Sidebar (侧边导航)

> 仅 Admin 端使用

详见 [admin/00-navigation.md](../admin/00-navigation.md) Sidebar 部分。

### 组件接口

```tsx
interface SidebarProps {
  collapsed: boolean;
  onToggle: () => void;
  items: NavItem[];
  activeItem: string;
}

interface NavItem {
  label: string;
  icon: LucideIcon;
  href: string;
  badge?: number;
  permission?: "super_admin";
}
```

### 折叠/展开

- 展开：`w-64`，显示图标 + 文字
- 折叠：`w-16`，仅显示图标 + Tooltip
- 过渡：`transition-all duration-200 ease-in-out`
- 偏好存 `localStorage`

---

## FormBuilder (动态表单)

> 仅 Admin 端 Agent Runtime 配置使用
> 将 `runtimes.config_schema` JSON Schema 渲染为表单

### 支持的 Schema 类型

| JSON Schema 类型 | 渲染组件 |
|------------------|----------|
| `string` | Input |
| `string + enum` | Select |
| `string + format=password` | Password Input |
| `string + format=textarea` | Textarea |
| `number / integer` | Number Input |
| `boolean` | Switch |
| `array` | 多选列表 |

### 布局

```
┌──────────────────────────────────────┐
│ Runtime 自定义配置                    │
│ ──────────────────────────────────── │
│                                      │
│ 自定义字段 1                         │
│ ┌──────────────────────────────────┐ │
│ │ 值                               │ │
│ └──────────────────────────────────┘ │
│                                      │
│ 自定义字段 2                         │
│ ┌──────────────┐                     │
│ │ 选项       ▾ │                     │
│ └──────────────┘                     │
│                                      │
│ 自定义开关                           │
│ [● 开启]                            │
└──────────────────────────────────────┘
```

### 验证

- 使用 `ajv` 或自定义验证器
- `required` 字段显示必填标记
- `minimum/maximum` 数字范围验证
- `minLength/maxLength` 字符串长度验证

---

## AgentLog (执行日志条目)

### 列表行 (Admin 日志表)

| 列 | 字段 | 说明 |
|------|------|------|
| 时间 | `started_at` | `HH:mm` 格式 |
| 类型 | `type` Badge | discuss/execute/system |
| 动作 | `action` | 带 action 图标 |
| 状态 | `status` | 带状态色图标 |
| 耗时 | `duration_ms` | 格式化为 `Xm Ys` |
| Issue | `issue_id` | 可点击链接 |

### 时间线条目 (Issue Timeline)

详见 [admin/06-issues.md](../admin/06-issues.md) 时间线部分。

### 动作图标映射

| action | 图标 | 颜色 |
|--------|------|------|
| generate | `FileCode2` | `text-blue-500` |
| test_pass | `CheckCircle2` | `text-green-500` |
| test_fail | `XCircle` | `text-red-500` |
| fix | `Wrench` | `text-amber-500` |
| commit | `GitCommit` | `text-violet-500` |
| create_pr | `GitPullRequest` | `text-purple-500` |
| error | `AlertTriangle` | `text-red-500` |
| paused | `Pause` | `text-orange-500` |
| plan | `Lightbulb` | `text-yellow-500` |
| discuss | `MessageSquare` | `text-sky-500` |
| default | `FileText` | `text-muted-foreground` |

---

## TokenUsage (用量展示)

### 紧凑版 (Issue 详情 / Context Panel)

```
Token 用量
────────────
输入   12,450
输出    3,200
────────────
总计   15,650
费用   $0.15
```

- 数字：`font-mono text-sm`
- 费用：`text-sm font-medium`
- 分隔线：`border-t border-dashed`

### 完整版 (Agent 详情 Token Tab)

包含按天柱状图 + 汇总统计卡片。详见 [admin/04-agents.md](../admin/04-agents.md) Token 用量 Tab。

### 费用估算色

| 费用范围 | 颜色 |
|----------|------|
| < $1 | `text-foreground` |
| $1 - $10 | `text-amber-600` |
| > $10 | `text-red-600` |

---

## Pagination (分页)

> 基于 shadcn/ui `Pagination`

### 标准分页

```
← 1 2 3 ... 5 →
```

### 带信息分页 (DataTable 底部)

```
显示 1-20 / 共 50 条          ← 1 2 3 →
```

### 简单加载更多 (消息列表)

```
↑ 加载更多消息
```

- 点击或滚动到顶部触发
- 加载中显示 Spinner
- 无更多数据显示 "没有更多消息"

### 元素样式

| 元素 | 样式 |
|------|------|
| 页码按钮 | `h-8 w-8 text-sm rounded-md` |
| 当前页 | `bg-primary text-primary-foreground` |
| 其他页 | `hover:bg-accent` |
| 箭头 | `variant="outline" size="icon" h-8 w-8` |
| 省略号 | `text-muted-foreground` |
| 信息文字 | `text-sm text-muted-foreground` |
