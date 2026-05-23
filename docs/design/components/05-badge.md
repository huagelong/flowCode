# Badge 标记 & 状态组件

> 基于 shadcn/ui `Badge` + 项目自定义变体

---

## 1. 标准 Badge

### shadcn/ui 默认变体

| Variant | 样式 | 用途 |
|---------|------|------|
| `default` | `bg-primary text-primary-foreground` | 默认 |
| `secondary` | `bg-secondary text-secondary-foreground` | 次要信息 |
| `destructive` | `bg-destructive text-destructive-foreground` | 错误/危险 |
| `outline` | `border border-current text-foreground` | 边框样式 |

### 通用样式

| 属性 | 值 |
|------|------|
| 内边距 | `px-2.5 py-0.5` |
| 圆角 | `rounded-full` |
| 字号 | `text-xs` |
| 字重 | `font-medium` |
| 行高 | `leading-none` |

---

## 2. 状态 Badge (Issue)

| 状态 | 背景 | 文字 | 图标 |
|------|------|------|------|
| backlog | `bg-slate-100 text-slate-700` | "Backlog" | `CircleDot` |
| todo | `bg-blue-100 text-blue-700` | "Todo" | `Circle` |
| in_progress | `bg-amber-100 text-amber-700` | "进行中" | `Loader2` (spin) |
| paused | `bg-orange-100 text-orange-700` | "已暂停" | `Pause` |
| in_review | `bg-purple-100 text-purple-700` | "审核中" | `Eye` |
| done | `bg-green-100 text-green-700` | "已完成" | `CheckCircle2` |

---

## 3. 优先级 Badge

| 优先级 | 背景 | 文字 | 图标 |
|--------|------|------|------|
| P0 | `bg-red-50 text-red-700` | "P0" | `AlertCircle` |
| P1 | `bg-orange-50 text-orange-700` | "P1" | `ArrowUp` |
| P2 | `bg-amber-50 text-amber-700` | "P2" | `Minus` |
| P3 | `bg-blue-50 text-blue-700` | "P3" | `ArrowDown` |
| P4 | `bg-slate-50 text-slate-700` | "P4" | `ArrowDown` |

---

## 4. 角色 Badge

### 成员角色

| 角色 | 样式 |
|------|------|
| owner | `bg-primary text-primary-foreground` |
| admin | `bg-secondary text-secondary-foreground` |
| member | `bg-muted text-muted-foreground` |

### Agent 角色标签

| 角色 | 样式 |
|------|------|
| CEO | `bg-emerald-100 text-emerald-700` |
| CTO | `bg-violet-100 text-violet-700` |
| Frontend | `bg-sky-100 text-sky-700` |
| Backend | `bg-amber-100 text-amber-700` |
| DevOps | `bg-orange-100 text-orange-700` |
| QA | `bg-pink-100 text-pink-700` |
| Custom | `bg-zinc-100 text-zinc-700` |

---

## 5. 来源 Badge

| 来源 | 样式 | 图标 |
|------|------|------|
| manual | `bg-blue-50 text-blue-700` + `Pencil` | 手动创建 |
| zip | `bg-purple-50 text-purple-700` + `Package` | ZIP 导入 |
| builtin | `bg-zinc-100 text-zinc-600` + `Lock` | 内置 |

---

## 6. 数量 Badge (未读/计数)

### 圆形数字

```
┌──────┐
│  🔔  │ ← 图标或文字
│   (3)│ ← 数量 badge
└──────┘
```

- 位置：绝对定位右上角
- 背景：`bg-destructive text-destructive-foreground`
- 尺寸：`min-w-[18px] h-[18px]`
- 圆角：`rounded-full`
- 字号：`text-[10px] font-medium`
- 数字 > 99 显示 "99+"
- 数字为 0 时隐藏

### 内联数字

```
backlog (3)  todo (5)
```

- `text-xs text-muted-foreground ml-1`
- 无背景
