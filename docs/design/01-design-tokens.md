# 设计令牌 & 技术规范

> 所有页面和组件的设计基准，基于 shadcn/ui 默认主题 + Tailwind CSS。

---

## 1. 颜色系统

采用 shadcn/ui 的 CSS 变量体系，支持 Light / Dark 双主题。

### 1.1 基础调色板 (HSL)

```css
:root {
  /* ---- 背景 ---- */
  --background: 0 0% 100%;         /* #ffffff */
  --foreground: 240 10% 3.9%;      /* #0a0a0c */

  /* ---- 卡片 ---- */
  --card: 0 0% 100%;
  --card-foreground: 240 10% 3.9%;

  /* ---- 弹出层 ---- */
  --popover: 0 0% 100%;
  --popover-foreground: 240 10% 3.9%;

  /* ---- 主色 (Zinc 黑灰系) ---- */
  --primary: 240 5.9% 10%;         /* #18181b */
  --primary-foreground: 0 0% 98%;  /* #fafafa */

  /* ---- 次要 ---- */
  --secondary: 240 4.8% 95.9%;     /* #f4f4f5 */
  --secondary-foreground: 240 5.9% 10%;

  /* ---- 静默 ---- */
  --muted: 240 4.8% 95.9%;
  --muted-foreground: 240 3.8% 46.1%; /* #71717a */

  /* ---- 强调 ---- */
  --accent: 240 4.8% 95.9%;
  --accent-foreground: 240 5.9% 10%;

  /* ---- 破坏性 ---- */
  --destructive: 0 84.2% 60.2%;    /* #ef4444 */
  --destructive-foreground: 0 0% 98%;

  /* ---- 边框 ---- */
  --border: 240 5.9% 90%;          /* #e4e4e7 */
  --input: 240 5.9% 90%;
  --ring: 240 5.9% 10%;

  /* ---- 圆角 ---- */
  --radius: 0.5rem;
}

.dark {
  --background: 240 10% 3.9%;
  --foreground: 0 0% 98%;
  --card: 240 10% 3.9%;
  --card-foreground: 0 0% 98%;
  --popover: 240 10% 3.9%;
  --popover-foreground: 0 0% 98%;
  --primary: 0 0% 98%;
  --primary-foreground: 240 5.9% 10%;
  --secondary: 240 3.7% 15.9%;
  --secondary-foreground: 0 0% 98%;
  --muted: 240 3.7% 15.9%;
  --muted-foreground: 240 5% 64.9%;
  --accent: 240 3.7% 15.9%;
  --accent-foreground: 0 0% 98%;
  --destructive: 0 62.8% 30.6%;
  --destructive-foreground: 0 0% 98%;
  --border: 240 3.7% 15.9%;
  --input: 240 3.7% 15.9%;
  --ring: 240 4.9% 83.9%;
}
```

### 1.2 语义色扩展 (项目自定义)

```css
:root {
  /* ---- 状态色 ---- */
  --success: 142 76% 36%;          /* #16a34a green-600 */
  --success-foreground: 0 0% 100%;
  --warning: 38 92% 50%;           /* #f59e0b amber-500 */
  --warning-foreground: 0 0% 100%;
  --info: 221 83% 53%;             /* #3b82f6 blue-500 */
  --info-foreground: 0 0% 100%;
}

.dark {
  --success: 142 71% 45%;
  --success-foreground: 0 0% 100%;
  --warning: 38 92% 50%;
  --warning-foreground: 0 0% 100%;
  --info: 221 83% 53%;
  --info-foreground: 0 0% 100%;
}
```

### 1.3 Issue 状态色

Issue 状态语义统一如下：`backlog` 表示原始需求；`todo` 表示从某个 backlog 需求分析出的任务列表，任务通过 `parent_id` 归属到同一个 backlog Issue 下；`in_progress` 表示正在运行的任务；`in_review` 表示需要人工审核，此时创建/查看 PR；`done` 表示 PR 已完成并合并。除 `in_review` 外，其余状态不设计人工确认关卡。

| 状态 | 颜色 | Tailwind Class | CSS 变量 |
|------|------|----------------|----------|
| `backlog`（需求） | Slate 灰 | `bg-slate-100 text-slate-700` | `--status-backlog` |
| `todo`（任务列表） | Blue 蓝 | `bg-blue-100 text-blue-700` | `--status-todo` |
| `in_progress` | Amber 琥珀 | `bg-amber-100 text-amber-700` | `--status-in-progress` |
| `paused` | Orange 橙 | `bg-orange-100 text-orange-700` | `--status-paused` |
| `in_review`（审核中/PR） | Purple 紫 | `bg-purple-100 text-purple-700` | `--status-in-review` |
| `done`（已完成） | Green 绿 | `bg-green-100 text-green-700` | `--status-done` |

### 1.4 优先级颜色

| 优先级 | 颜色 | Tailwind Class |
|--------|------|----------------|
| P0 | Red | `text-red-600 bg-red-50` |
| P1 | Orange | `text-orange-600 bg-orange-50` |
| P2 | Amber | `text-amber-600 bg-amber-50` |
| P3 | Blue | `text-blue-600 bg-blue-50` |
| P4 | Slate | `text-slate-600 bg-slate-50` |

### 1.5 Agent 角色标签色

| 角色 | 颜色 | Tailwind Class |
|------|------|----------------|
| CEO | Emerald | `bg-emerald-100 text-emerald-700` |
| CTO | Violet | `bg-violet-100 text-violet-700` |
| Frontend | Sky | `bg-sky-100 text-sky-700` |
| Backend | Amber | `bg-amber-100 text-amber-700` |
| DevOps | Orange | `bg-orange-100 text-orange-700` |
| QA | Pink | `bg-pink-100 text-pink-700` |
| Custom | Zinc | `bg-zinc-100 text-zinc-700` |

---

## 2. 排版 (Typography)

### 2.1 字体栈

```css
/* 默认无衬线 */
--font-sans: "Inter", ui-sans-serif, system-ui, -apple-system, sans-serif;

/* 等宽（代码） */
--font-mono: "JetBrains Mono", ui-monospace, SFMono-Regular, monospace;
```

### 2.2 字号体系

| 级别 | Tailwind | 尺寸 | 行高 | 用途 |
|------|----------|------|------|------|
| h1 | `text-3xl` | 30px | 36px (1.2) | 页面标题（极少使用） |
| h2 | `text-2xl` | 24px | 32px (1.33) | 模块标题 |
| h3 | `text-xl` | 20px | 28px (1.4) | 区块标题 |
| h4 | `text-lg` | 18px | 28px (1.55) | 卡片标题 |
| body | `text-sm` | 14px | 20px (1.43) | 正文 / 表格 / 表单 |
| caption | `text-xs` | 12px | 16px (1.33) | 辅助文字 / 时间戳 |
| micro | `text-[10px]` | 10px | 14px (1.4) | Badge 内文字 |

字重：`font-normal (400)` 正文，`font-medium (500)` 标题/标签，`font-semibold (600)` 强调。

---

## 3. 间距系统

基于 Tailwind 默认 4px 基数：

| Token | 值 | Tailwind | 典型用途 |
|-------|------|----------|----------|
| 0.5 | 2px | `gap-0.5` | 图标与文字间距 |
| 1 | 4px | `gap-1 p-1` | 紧凑内边距 |
| 1.5 | 6px | `gap-1.5` | Badge 内边距 |
| 2 | 8px | `gap-2 p-2` | 表单项间距 |
| 3 | 12px | `gap-3 p-3` | 卡片内边距 |
| 4 | 16px | `gap-4 p-4` | 区块间距 |
| 5 | 20px | `gap-5` | 模块间距 |
| 6 | 24px | `gap-6 p-6` | 页面内边距 |
| 8 | 32px | `gap-8` | 大区块分隔 |
| 10 | 40px | `gap-10` | 页面级分隔 |
| 12 | 48px | `gap-12` | 最大间距 |

---

## 4. 圆角

| Token | 值 | Tailwind | 用途 |
|-------|------|----------|------|
| `--radius` | 6px | `rounded-md` | 默认：按钮、输入框、卡片 |
| — | 8px | `rounded-lg` | Dialog、Dropdown |
| — | 12px | `rounded-xl` | Modal 大面板 |
| — | 9999px | `rounded-full` | Avatar、Badge 药丸形 |

---

## 5. 阴影

| 级别 | Tailwind | 场景 |
|------|----------|------|
| 无 | `shadow-none` | 内嵌面板 |
| sm | `shadow-sm` | 卡片、表格行 |
| md | `shadow-md` | Dropdown、Popover |
| lg | `shadow-lg` | Dialog、Modal |

---

## 6. 动效参数

### 6.1 过渡时长

| 名称 | 时长 | Tailwind | 用途 |
|------|------|----------|------|
| instant | 100ms | `duration-100` | 按钮 hover/active 状态 |
| fast | 150ms | `duration-150` | 背景色、边框变化 |
| normal | 200ms | `duration-200` | 展开/折叠 |
| slow | 300ms | `duration-300` | Dialog 进出 |
| slower | 500ms | `duration-500` | 页面切换 |

### 6.2 缓动函数

```css
--ease-default: cubic-bezier(0.4, 0, 0.2, 1);   /* Tailwind 默认 */
--ease-in: cubic-bezier(0.4, 0, 1, 1);
--ease-out: cubic-bezier(0, 0, 0.2, 1);
--ease-in-out: cubic-bezier(0.4, 0, 0.2, 1);
--ease-spring: cubic-bezier(0.34, 1.56, 0.64, 1);  /* 弹性效果 */
```

### 6.3 Framer Motion 配置

```tsx
// 页面切换
const pageTransition = {
  initial: { opacity: 0, y: 8 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -8 },
  transition: { duration: 0.2, ease: [0.4, 0, 0.2, 1] },
};

// 列表项入场
const listItemVariant = {
  hidden: { opacity: 0, x: -8 },
  visible: (i: number) => ({
    opacity: 1, x: 0,
    transition: { delay: i * 0.03, duration: 0.15 },
  }),
};

// Dialog 弹出
const dialogVariant = {
  initial: { opacity: 0, scale: 0.95 },
  animate: { opacity: 1, scale: 1 },
  exit: { opacity: 0, scale: 0.95 },
  transition: { duration: 0.15 },
};

// Sidebar 展开/折叠
const sidebarVariant = {
  expanded: { width: 256 },
  collapsed: { width: 64 },
  transition: { duration: 0.2, ease: [0.4, 0, 0.2, 1] },
};
```

---

## 7. 图标规范

使用 `lucide-react` 图标库。

| 尺寸 | Tailwind | 用途 |
|------|----------|------|
| 14px | `h-3.5 w-3.5` | 表格内行内图标 |
| 16px | `h-4 w-4` | 按钮、菜单项、表单标签 |
| 18px | `h-[18px] w-[18px]` | 导航图标 |
| 20px | `h-5 w-5` | 空状态、大按钮 |
| 24px | `h-6 w-6` | 页面标题旁 |

线宽统一 `strokeWidth={1.5}` 或 `strokeWidth={2}`（默认 2）。

---

## 8. z-index 层级

| 层级 | 值 | 用途 |
|------|------|------|
| base | 0 | 正常内容 |
| dropdown | 50 | Dropdown、Popover |
| sticky | 100 | 粘性头部 |
| overlay | 200 | 背景遮罩 |
| modal | 300 | Dialog、Modal |
| popover | 400 | Tooltip、Command Palette |
| toast | 500 | Sonner 通知 |
| debug | 9999 | 开发调试浮层 |

---

## 9. 聚焦 & 无障碍 (A11y)

### 9.1 焦点环

```css
/* shadcn/ui 默认 focus-visible 环 */
.focus-visible\:ring-2 {
  outline: none;
  box-shadow: 0 0 0 2px hsl(var(--background)),
              0 0 0 4px hsl(var(--ring));
}
```

### 9.2 A11y 规范

- 所有交互元素必须可键盘 Tab 导航
- Dialog 打开时焦点捕获 (focus trap)
- `aria-label` 用于图标按钮
- `role="status"` 用于实时更新区域 (聊天消息流、Timeline)
- `aria-live="polite"` 用于通知数量变化
- 颜色对比度 ≥ 4.5:1 (WCAG AA)
- 不依赖颜色唯一传达信息，必须配合文字或图标

---

## 10. 暗色模式策略

- 使用 `next-themes` 的 `class` 策略（在 `<html>` 上添加 `.dark`）
- 所有颜色通过 CSS 变量引用，不硬编码颜色值
- 图片/Logo 提供 light/dark 两个版本
- 暗色模式下降低背景饱和度，保持可读性
- 代码块暗色模式统一使用深色背景 (One Dark Pro 主题)
