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

### 可展开详情

`agent_log` 类型的时间线条目支持点击展开内联详情，渲染对应类型的 AgentOutputCard：

| action | 展开内容 | 组件 |
|--------|---------|------|
| `generate` | 文件列表 + 可展开 DiffViewer | AgentOutputCard (文件生成) |
| `commit` | 提交信息 + 哈希 + 文件统计 | AgentOutputCard (Commit) |
| `create_pr` | PR 标题 + 分支 + 外部链接 | AgentOutputCard (PR 创建) |
| `test_pass` | 通过用例列表 | AgentOutputCard (测试通过) |
| `test_fail` | 失败详情 + 堆栈 | AgentOutputCard (测试失败) |
| `error` | 错误信息 + 重试按钮 | AgentOutputCard (错误) |
| `fix` | 修复文件列表 | AgentOutputCard (文件生成) |
| 其他 | 不展开 | — |

**交互规则**:
- 手风琴模式：同时仅展开一个条目
- 展开图标：`ChevronDown` ↔ `ChevronUp`
- 展开动画：`max-height` 过渡 + `opacity`，200ms (Admin) / 150ms (Client)
- Admin 完整版：DiffViewer 内联展开
- Client 紧凑版：DiffViewer 通过 Sheet 弹出展示

各端具体展开规范见：
- Admin: [admin/06-issues.md](../admin/06-issues.md) 可展开详情部分
- Client: [client/04-issue-detail.md](../client/04-issue-detail.md) 可展开详情部分

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

> Client 端 Agent 消息 + Admin 群组消息只读视图

### 行内代码

```
使用 `npm install` 安装依赖
```

- `bg-muted rounded px-1.5 py-0.5 text-sm font-mono`

### 代码块

```
┌──────────────────────────────────────┐
│ src/components/Button.tsx     [📋]  │
│ ──────────────────────────────────── │
│  1 │ interface ButtonProps {         │
│  2 │   variant: 'default' | 'ghost'; │
│  3 │   size: 'sm' | 'default';       │
│  4 │ }                               │
│  5 │                                 │
│  6 │ export function Button() {      │
│  7 │   return <button />;            │
│  8 │ }                               │
└──────────────────────────────────────┘
```

| 属性 | 值 |
|------|------|
| 背景 | `bg-zinc-900 text-zinc-100` (暗色主题)，`bg-zinc-100` (亮色主题) |
| 字体 | `font-mono text-sm` |
| 圆角 | `rounded-lg` |
| 头部 | 文件路径或语言标签 + 复制按钮 |
| 行号 | `text-zinc-500 text-xs` 左侧 |
| 复制按钮 | `Clipboard` 图标，复制后变 `Check`，2s 恢复 |
| 最大高度 | `max-h-[400px] overflow-y-auto`，超出显示滚动条 |
| 折叠 | 超过 30 行时显示折叠按钮 "展开全部 (42 行)" |

### 语法高亮

使用 `shiki` 或 `prism-react-renderer`：
- 支持语言：typescript, javascript, python, go, sql, yaml, json, bash, css, html, markdown
- 主题：One Dark Pro (暗色模式)，GitHub Light (亮色模式)

---

## DiffViewer (代码差异展示)

> Client 聊天 + Admin/Client 时间线详情展开

展示 Agent 对文件的代码改动，支持 inline diff 渲染。

### Inline Diff 模式

```
┌──────────────────────────────────────────────────────────┐
│ 📄 src/components/LoginForm.tsx                   [📋]  │
│ +12 -3 行变更                                            │
│ ──────────────────────────────────────────────────────── │
│  1 │ import { useState } from 'react';                   │
│  2 │ import { z } from 'zod';                            │
│  3 │                                                     │
│  4 │ const schema = z.object({                           │
│  5 │   email: z.string().email(),                        │
│ - 6 │   password: z.string().min(6),                     │
│ + 6 │   password: z.string()                             │
│ + 7 │     .min(8, '至少8个字符')                          │
│ + 8 │     .regex(/[A-Z]/, '需包含大写字母')               │
│ + 9 │     .regex(/[0-9]/, '需包含数字'),                  │
│ 10 │ });                                                 │
│ ...                                                      │
│                                         [展开完整 diff]  │
└──────────────────────────────────────────────────────────┘
```

### 样式

| 属性 | 值 |
|------|------|
| 添加行背景 | `bg-green-500/10 text-green-300` (暗色)，`bg-green-50 text-green-800` (亮色) |
| 添加行号标记 | `text-green-500` + 前缀 `+` |
| 删除行背景 | `bg-red-500/10 text-red-300` (暗色)，`bg-red-50 text-red-800` (亮色) |
| 删除行号标记 | `text-red-500` + 前缀 `-` |
| 未更改行 | 正常代码块样式，`text-zinc-500` |
| 折叠区域 | `text-zinc-500 text-xs bg-zinc-800/50 rounded px-2 py-0.5`，显示 "..." |
| 头部 | 文件路径 + 变更统计 `+N -M` + 复制按钮 |
| 字体 | `font-mono text-xs` (比 CodeBlock 更紧凑) |
| 最大高度 | `max-h-[300px]` |

### 变更统计 Badge

```
+12 -3 行变更
```

- `+N`：`text-green-600 font-mono text-xs font-medium`
- `-M`：`text-red-600 font-mono text-xs font-medium`
- 分隔：`text-muted-foreground text-xs`

### 折叠/展开

- 默认折叠超过 20 行的 diff，仅显示变更行 ± 3 行上下文
- "展开完整 diff" 按钮：`text-xs text-primary hover:underline`
- 展开后显示完整文件内容

---

## AgentOutputCard (Agent 成果卡片)

> Client 聊天消息 + Admin 群组消息 + 时间线展开详情

Agent 执行产生的结构化输出以卡片形式展示，区别于普通 Markdown 文本消息。

### 1. 文件生成卡片

Agent 生成/修改文件时：

```
┌──────────────────────────────────────────────────────────┐
│ 📄 生成了 3 个文件                                [展开] │
│ ──────────────────────────────────────────────────────── │
│ ✅ src/components/LoginForm.tsx     +42 -5              │
│ ✅ src/hooks/useAuth.ts             +28 new             │
│ ✅ src/__tests__/login.test.ts      +35 new             │
│ ──────────────────────────────────────────────────────── │
│ 合计: +105 -5 行                                         │
└──────────────────────────────────────────────────────────┘
```

展开后每个文件显示 DiffViewer：

```
┌──────────────────────────────────────────────────────────┐
│ 📄 生成了 3 个文件                                 [折叠]│
│ ──────────────────────────────────────────────────────── │
│ ▼ src/components/LoginForm.tsx             +42 -5       │
│ ┌────────────────────────────────────────────────────┐  │
│ │ (DiffViewer 内联)                                  │  │
│ └────────────────────────────────────────────────────┘  │
│ ▶ src/hooks/useAuth.ts                     +28 new     │
│ ▶ src/__tests__/login.test.ts              +35 new     │
└──────────────────────────────────────────────────────────┘
```

| 属性 | 值 |
|------|------|
| 背景 | `bg-card border rounded-lg` |
| 头部 | `FileCode2` 图标 + "生成了 N 个文件" + 折叠按钮 |
| 文件列表 | 每行 `text-sm font-mono`，点击展开 DiffViewer |
| 文件状态 | ✅ 新建 `text-green-500`，✅ 修改 `text-blue-500` |
| 变更统计 | `+N -M` 每文件右侧 |
| 底部合计 | `text-xs text-muted-foreground` |

### 2. Commit 卡片

```
┌──────────────────────────────────────────────────────────┐
│ 📦 提交并推送                                             │
│ ──────────────────────────────────────────────────────── │
│ feat: add login form validation with zod                  │
│                                                           │
│ a1b2c3d  3 files changed  +105 -5                        │
│                                                           │
│ branch: feat/issue-42                                     │
└──────────────────────────────────────────────────────────┘
```

| 属性 | 值 |
|------|------|
| 背景 | `bg-card border rounded-lg` |
| 头部 | `GitCommit` 图标 `text-violet-500` + "提交并推送" |
| 提交消息 | `text-sm font-medium`，首行 |
| 哈希 | `font-mono text-xs text-muted-foreground`，前 7 位 |
| 文件变更 | `text-xs text-muted-foreground` |
| 分支名 | `bg-muted rounded px-1.5 py-0.5 text-xs font-mono` |

### 3. PR 创建卡片

```
┌──────────────────────────────────────────────────────────┐
│ 🔀 创建了 Pull Request                                   │
│ ──────────────────────────────────────────────────────── │
│ #15 feat: add login form validation                      │
│                                                           │
│ feat/issue-42 → main  +105 -5  3 files                   │
│                                                           │
│                              [在 GitHub 中查看 →]        │
└──────────────────────────────────────────────────────────┘
```

| 属性 | 值 |
|------|------|
| 背景 | `bg-card border rounded-lg` |
| 头部 | `GitPullRequest` 图标 `text-purple-500` + "创建了 Pull Request" |
| PR 标题 | `text-sm font-medium`，可点击打开 `pr_url` |
| 分支信息 | `text-xs text-muted-foreground`，`source → target` 格式 |
| 查看按钮 | `variant="outline" size="sm"`，打开外部链接 |

### 4. 测试结果卡片

```
┌──────────────────────────────────────────────────────────┐
│ ✅ 测试通过                                    4/4 通过   │
│ ──────────────────────────────────────────────────────── │
│ ✓ LoginForm > validates email                           │
│ ✓ LoginForm > validates password strength               │
│ ✓ LoginForm > submits on valid input                    │
│ ✓ LoginForm > shows error on invalid input              │
└──────────────────────────────────────────────────────────┘
```

失败时：

```
┌──────────────────────────────────────────────────────────┐
│ ❌ 测试失败                                    3/4 通过   │
│ ──────────────────────────────────────────────────────── │
│ ✓ LoginForm > validates email                           │
│ ✓ LoginForm > validates password strength               │
│ ✓ LoginForm > submits on valid input                    │
│ ──────────────────────────────────────────────────────── │
│ ✗ LoginForm > shows error on invalid input              │
│   Expected: "密码至少8个字符"                             │
│   Received: "密码不能为空"                                │
│                                             [展开堆栈]   │
└──────────────────────────────────────────────────────────┘
```

| 属性 | 值 |
|------|------|
| 全部通过背景 | `bg-green-50/50 border-green-200 rounded-lg` |
| 有失败背景 | `bg-red-50/50 border-red-200 rounded-lg` |
| 通过项 | `text-sm` + `CheckCircle2 h-3.5 w-3.5 text-green-500` |
| 失败项 | `text-sm font-medium` + `XCircle h-3.5 w-3.5 text-red-500` |
| 失败详情 | `text-xs font-mono text-red-700 bg-red-50 rounded px-2 py-1` |
| 堆栈跟踪 | 默认折叠，"展开堆栈" 展开后显示 `font-mono text-xs bg-zinc-900 text-zinc-100` |

### 5. 错误卡片

```
┌──────────────────────────────────────────────────────────┐
│ ⚠️ 执行错误                                              │
│ ──────────────────────────────────────────────────────── │
│ Command timed out after 30 minutes                        │
│                                                           │
│ The agent process exceeded the configured timeout.        │
│                                                           │
│                                    [展开详情]  [重试]     │
└──────────────────────────────────────────────────────────┘
```

| 属性 | 值 |
|------|------|
| 背景 | `bg-destructive/5 border-destructive/30 rounded-lg` |
| 头部 | `AlertTriangle h-4 w-4 text-destructive` + "执行错误" |
| 错误消息 | `text-sm font-medium text-destructive` |
| 描述 | `text-sm text-muted-foreground` |
| 展开详情 | 显示完整堆栈跟踪，`font-mono text-xs bg-zinc-900 rounded p-3` |
| 重试按钮 | `variant="outline" size="sm"` |

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
