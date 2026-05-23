# Button 按钮变体

> 基于 shadcn/ui `Button`，扩展项目特有变体。

---

## 1. 变体 (Variants)

### 标准 shadcn/ui 变体

| Variant | 样式 | 用途 |
|---------|------|------|
| `default` | `bg-primary text-primary-foreground` | 主要操作 (保存、创建、发送) |
| `destructive` | `bg-destructive text-destructive-foreground` | 危险操作 (删除、停止) |
| `outline` | `border border-input bg-background` | 次要操作 (取消、筛选) |
| `secondary` | `bg-secondary text-secondary-foreground` | 辅助操作 |
| `ghost` | `hover:bg-accent hover:text-accent-foreground` | 工具栏、图标按钮 |
| `link` | `text-primary underline-offset-4 hover:underline` | 文字链接 |

### 项目扩展变体

| Variant | 样式 | 用途 |
|---------|------|------|
| `success` | `bg-green-600 text-white hover:bg-green-700` | 确认操作 (接受邀请) |
| `warning` | `bg-amber-500 text-white hover:bg-amber-600` | 警告操作 (暂停执行) |

---

## 2. 尺寸 (Sizes)

| Size | Padding | 高度 | 字号 | 用途 |
|------|---------|------|------|------|
| `sm` | `h-8 px-3` | 32px | `text-xs` | 表格操作、紧凑区域 |
| `default` | `h-9 px-4 py-2` | 36px | `text-sm` | 标准按钮 |
| `lg` | `h-10 px-6` | 40px | `text-sm` | 页面主操作 |
| `icon` | `h-9 w-9` | 36px | — | 图标按钮 |

---

## 3. 状态

| 状态 | 样式 |
|------|------|
| 默认 | 同 Variant 定义 |
| Hover | `opacity: 0.9` (default/destructive)，`bg-accent` (ghost) |
| Active | `transform: scale(0.98)`，`opacity: 0.8` |
| Focus | `ring-2 ring-ring ring-offset-2` |
| Disabled | `opacity: 0.5`，`pointer-events: none` |
| Loading | 文字替换为 `Loader2` Spinner + "处理中..." |

### Loading 态实现

```tsx
<Button disabled={isLoading}>
  {isLoading ? (
    <>
      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
      处理中...
    </>
  ) : (
    "保存"
  )}
</Button>
```

---

## 4. 图标按钮

### 仅图标

```tsx
<Button variant="ghost" size="icon" aria-label="编辑">
  <Pencil className="h-4 w-4" />
</Button>
```

### 图标 + 文字

```tsx
<Button>
  <Plus className="mr-2 h-4 w-4" />
  创建
</Button>
```

---

## 5. 按钮组合

### Dialog 底部

```
                     [取消]  [确认删除]
```

- 取消：`variant="outline"`
- 确认：`variant="destructive"`

### 表单底部

```
                     [取消]  [保存]
```

- 取消：`variant="outline"`
- 保存：`variant="default"`

### 批量操作栏

```
已选择 3 项        [批量转 Todo]  [批量改优先级]
```

- 批量操作：`variant="outline" size="sm"`

---

## 6. 可访问性

- 所有按钮有明确的 `aria-label`（图标按钮必须有）
- Focus-visible 样式
- Disabled 使用 `aria-disabled="true"`
- Loading 态使用 `aria-busy="true"`
