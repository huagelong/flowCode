# Avatar 头像组件

> 基于 shadcn/ui `Avatar` + 自定义扩展

---

## 1. 变体

### 用户 Avatar

| 尺寸 | 用途 | 圆形直径 |
|------|------|----------|
| `xs` | 群组成员列表 | 24px |
| `sm` | 消息气泡、表格 | 32px |
| `md` | 会话列表、导航栏 | 40px |
| `lg` | 详情页头部 | 56px |
| `xl` | 个人资料页 | 64px |

### Agent Avatar

与用户 Avatar 相同尺寸，但右下角叠加 Bot 角标：

```
┌────────┐
│        │
│   🤖   │
│     ◉──┤ ← Bot 角标
└────────┘
```

- Bot 角标：圆形，16px（xs 12px）
- 背景色：`bg-primary`
- 图标：`Bot` (lucide)，白色，`h-2.5 w-2.5`（xs `h-2 w-2`）
- 位置：绝对定位 `bottom-[-2px] right-[-2px]`
- 边框：`ring-2 ring-background`

### Fallback (无头像)

| 类型 | 样式 |
|------|------|
| 用户 | 用户名首字母大写，`bg-muted text-muted-foreground` |
| Agent | `Bot` 图标，`bg-primary/10 text-primary` |

---

## 2. 群组 Avatar

当显示群组头像时，使用成员头像 2×2 网格：

```
┌────────┐
│ 👤 │ 🤖 │
│────┼────│
│ 🤖 │ 👤 │
└────────┘
```

- 最大显示 4 个成员头像
- 每个头像为群组 Avatar 的 45%
- 间距 `gap-0.5`
- 超出 4 人：第 4 格显示 "+N"

---

## 3. 在线状态指示器

Avatar 右下角显示在线状态圆点：

| 状态 | 颜色 | 尺寸 |
|------|------|------|
| 在线 | `bg-green-500` | 10px (md Avatar) |
| 执行中 | `bg-amber-500` | 10px |
| 离线 | `bg-zinc-400` | 10px |

- 绝对定位 `bottom-0 right-0`
- 边框 `ring-2 ring-background`

---

## 4. Avatar 组 (Stack)

用于显示"成员"列表时多个头像叠加：

```
  ┌──┐┌──┐┌──┐
  │👤││🤖││🤖│ +2
  └──┘└──┘└──┘
```

- 重叠 `-ml-2`（第一个无偏移）
- 最多显示 4 个，超出显示 `+N`
- 每个带 `ring-2 ring-background` 分隔
- 尺寸：24px 或 32px
