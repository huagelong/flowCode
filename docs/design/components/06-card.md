# Card 卡片容器

> 基于 shadcn/ui `Card`

---

## 1. 标准 Card

```
┌────────────────────────────────────────────┐
│ Card Header                                │
│ 标题                        [操作按钮]     │
│ 描述文字                                   │
│ ────────────────────────────────────────── │
│ Card Content                               │
│                                            │
│                                            │
│ ────────────────────────────────────────── │
│ Card Footer (可选)                         │
└────────────────────────────────────────────┘
```

| 部分 | 样式 |
|------|------|
| 容器 | `rounded-lg border bg-card text-card-foreground shadow-sm` |
| Header | `flex flex-col space-y-1.5 p-6` |
| 标题 | `text-lg font-semibold leading-none tracking-tight` |
| 描述 | `text-sm text-muted-foreground` |
| Content | `p-6 pt-0` |
| Footer | `flex items-center p-6 pt-0` |

---

## 2. 统计卡片 (Dashboard)

```
┌────────────────────────┐
│ 总项目数                │
│                        │
│  12                    │
│  ↑ 2 本周新增          │
└────────────────────────┘
```

| 元素 | 样式 |
|------|------|
| 标题 | `text-sm font-medium text-muted-foreground` |
| 主数字 | `text-3xl font-bold tracking-tight` |
| 趋势 | `text-xs text-muted-foreground`，上升 `text-green-600`，下降 `text-red-600` |
| 内边距 | `p-6` |

---

## 3. 项目卡片 (Client)

```
┌──────────────────────────┐
│ 📁 anserFlow             │
│ AI Agent 协作平台         │
│                          │
│ ████████████░░░░ 68%     │
│                          │
│ 5 进行中 · 12 已完成      │
│ 🤖 3 · 👤 2              │
│                          │
│ 最近更新: 5 分钟前        │
└──────────────────────────┘
```

详见 [client/05-project-list.md](../client/05-project-list.md)

---

## 4. Agent 状态卡片 (Context Panel)

```
┌─────────────────────────┐
│ 🤖 前端Agent            │
│ Frontend · anserAgent   │
│ 当前: Issue #42          │
│ 🟡 执行中               │
└─────────────────────────┘
```

| 属性 | 值 |
|------|------|
| 内边距 | `p-3` |
| 间距 | `mb-2` |
| 圆角 | `rounded-lg border` |
| Hover | `bg-accent/50` |

---

## 5. 可点击 Card

需要点击跳转的卡片添加 `cursor-pointer` + `transition-colors`：

```tsx
<Card
  className="cursor-pointer transition-colors hover:bg-accent/50"
  onClick={() => navigate(`/admin/projects/${id}`)}
>
```
