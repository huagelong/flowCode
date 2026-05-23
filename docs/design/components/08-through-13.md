# Dropdown / Command / MessageBubble / Timeline / CodeBlock / Chart

---

## Dropdown (DropdownMenu)

> 基于 shadcn/ui `DropdownMenu`

### 触发器

```tsx
<Button variant="ghost" size="icon">
  <MoreHorizontal className="h-4 w-4" />
</Button>
```

### 下拉面板

```
┌──────────────────────────┐
│ ✏️ 编辑                   │
│ 📜 查看日志               │
│ 💰 Token 用量             │
│ ──────────────────────── │
│ 🗑️ 删除                 │
└──────────────────────────┘
```

| 属性 | 值 |
|------|------|
| 最小宽度 | `w-48` |
| 背景 | `bg-popover` |
| 圆角 | `rounded-md` |
| 阴影 | `shadow-md` |
| 对齐 | `align="end"` (右侧触发) |
| 项高度 | `h-8` |
| 项内边距 | `px-2` |
| Hover | `bg-accent` |
| 危险项 | `text-destructive focus:text-destructive` |
| 分隔线 | `Separator` 组件 |
| 图标 | `h-4 w-4 mr-2` |
| 快捷键 | `ml-auto text-xs text-muted-foreground` |

---

## Command (命令面板)

> 基于 shadcn/ui `Command`

### 全局搜索 (Ctrl/Cmd + K)

```
┌──────────────────────────────────────────────────┐
│ 🔍 搜索页面、会话、Agent...                       │
│ ──────────────────────────────────────────────── │
│                                                  │
│ 最近访问                                         │
│ 📁 anserFlow 项目                               │
│ 💬 项目讨论组                                    │
│ 🤖 前端Agent                                    │
│ ──────────────────────────────────────────────── │
│                                                  │
│ 快捷操作                                         │
│ +  创建新 Issue                                  │
│ +  创建新 Agent                                  │
│ +  创建新群组                                    │
│                                                  │
│ ↑↓ 导航  ↵ 选择  Esc 关闭                       │
└──────────────────────────────────────────────────┘
```

| 属性 | 值 |
|------|------|
| 宽度 | `w-[640px]` |
| 最大高度 | `max-h-[420px]` |
| 阴影 | `shadow-2xl` |
| 圆角 | `rounded-xl` |
| 分组标题 | `px-2 py-1.5 text-xs font-semibold text-muted-foreground` |
| 项高度 | `h-10` |
| 项 Hover | `bg-accent rounded-sm` |
| 选中态 | `bg-accent` |

### @提及面板 (聊天输入)

```
┌──────────────────────────┐
│ 🤖 CEO                   │
│ 🤖 前端Agent             │
│ 🤖 后端Agent             │
└──────────────────────────┘
```

- 仅显示群组中的 Agent 成员
- 浮于输入框上方
- 输入 `@` 后按关键词过滤

---

## MessageBubble (聊天气泡)

> 仅 Client 端使用

### 用户消息

```
┌────┐
│ 👤 │  张三                           14:30
└────┘
这是一条普通消息内容
```

### Agent 消息

```
┌────┐
│ 🤖 │  CTO · [CTO]                   14:31
└────┘
这是 Agent 的回复内容，支持 Markdown 渲染
```

### Markdown 渲染

Agent 消息内容使用 `prose` 类渲染 Markdown：
- `prose prose-sm max-w-none`
- 代码块：`prose-code:bg-muted prose-code:rounded prose-code:px-1`
- 列表：标准缩进
- 链接：`text-primary hover:underline`

### 消息分组规则
- 同一发送者 5 分钟内：仅第一条显示头像和名称
- 超过 5 分钟：重新显示头像和名称
- 日期变更：插入日期分隔线

---

## Timeline (时间线)

> Admin Issue 详情 (完整版) 和 Client Context Panel (紧凑版) 共用

### 完整版 (Admin)

```
┌─ 时间线 ────────────────────────────────────────┐
│ 筛选: [全部] [agent_log] [system] [human]        │
│                                                  │
│ 12:12  📦 系统                                   │
│        状态变更: 进行中 → 审核中                  │
│        ──────                                    │
│ 12:10  🤖 前端Agent                              │
│        📦 commit + push → PR #15                 │
│        ──────                                    │
│ 12:05  🤖 前端Agent                              │
│        📄 Generated: src/login.tsx               │
└──────────────────────────────────────────────────┘
```

- 事件间距：`py-2`
- 时间：`text-xs text-muted-foreground`
- 时间左侧连接线：`w-px bg-border` (可选)

### 紧凑版 (Client)

- 事件间距：`py-1`
- 时间：`text-[10px]`
- 图标：`h-3.5 w-3.5`
- 无连接线

### 实时追加动画

新事件追加到底部时：
```tsx
<motion.div
  initial={{ opacity: 0, y: 4 }}
  animate={{ opacity: 1, y: 0 }}
  transition={{ duration: 0.15 }}
>
```

---

## CodeBlock (代码展示)

> 仅 Client 端，Agent 消息中的代码输出

### 行内代码

```
使用 `npm install` 安装依赖
```

- `bg-muted rounded px-1.5 py-0.5 text-sm font-mono`

### 代码块

```
┌──────────────────────────────────────┐
│ typescript                    [📋]  │
│ ──────────────────────────────────── │
│ interface User {                     │
│   id: number;                        │
│   name: string;                      │
│   email: string;                     │
│ }                                    │
└──────────────────────────────────────┘
```

| 属性 | 值 |
|------|------|
| 背景 | `bg-zinc-900 text-zinc-100` (暗色主题)，`bg-zinc-100` (亮色主题) |
| 字体 | `font-mono text-sm` |
| 圆角 | `rounded-lg` |
| 头部 | 语言标签 + 复制按钮 |
| 行号 | `text-zinc-500 text-xs` 左侧 |
| 复制按钮 | `Clipboard` 图标，复制后变 `Check`，2s 恢复 |

---

## Chart (图表组件)

> 基于 Recharts 封装，仅 Admin Dashboard 使用

### 环形图 (PieChart)

Issue 状态分布。详见 [admin/02-dashboard.md](../admin/02-dashboard.md)

### 柱状图 (BarChart)

Agent 活跃度。详见 [admin/02-dashboard.md](../admin/02-dashboard.md)

### 面积图 (AreaChart)

Token 用量趋势。详见 [admin/02-dashboard.md](../admin/02-dashboard.md)

### 通用样式

| 属性 | 值 |
|------|------|
| 主色 | `hsl(var(--primary))` |
| 填充透明度 | `0.1` |
| Tooltip | `bg-popover border shadow-md rounded-md px-3 py-2 text-sm` |
| 网格线 | `stroke: hsl(var(--border))` |
| 动画 | `animationDuration={500}` |
