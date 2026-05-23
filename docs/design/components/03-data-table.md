# DataTable 数据表格

> 基于 shadcn/ui `Table` + TanStack Table
> 仅 Admin 端使用

---

## 1. 表格结构

```
┌──────────────────────────────────────────────────────────────────┐
│ ☐ │ 列 1       │ 列 2       │ 列 3       │ 列 4       │ 操作  │
│───┼────────────┼────────────┼────────────┼────────────┼───────│
│ ☐ │ 数据 1     │ 数据 2     │ 数据 3     │ 数据 4     │  ⋯   │
│ ☐ │ 数据 1     │ 数据 2     │ 数据 3     │ 数据 4     │  ⋯   │
│ ☐ │ 数据 1     │ 数据 2     │ 数据 3     │ 数据 4     │  ⋯   │
└──────────────────────────────────────────────────────────────────┘
│ 显示 1-20 / 共 50 条                ← 1 2 3 →                   │
└──────────────────────────────────────────────────────────────────┘
```

---

## 2. 表头 (Header)

| 属性 | 值 |
|------|------|
| 背景 | `bg-muted/50` |
| 文字 | `text-xs font-medium text-muted-foreground uppercase` |
| 内边距 | `h-10 px-4` |
| 排序图标 | `ArrowUpDown` (可排序列)，`ChevronUp/Down` (当前排序列) |
| Hover | 可排序列 `cursor-pointer hover:bg-muted` |

---

## 3. 表行 (Row)

| 状态 | 样式 |
|------|------|
| 默认 | `border-b border-border` |
| Hover | `bg-muted/50` |
| 选中 | `bg-primary/5` |
| 可点击 | `cursor-pointer` |
| 内边距 | `h-12 px-4` |

---

## 4. 单元格样式

| 内容类型 | 样式 |
|----------|------|
| 文本 | `text-sm` |
| ID | `font-mono text-sm text-muted-foreground` |
| 时间 | `text-sm text-muted-foreground` |
| Avatar + 名称 | `flex items-center gap-2`，Avatar 32px |
| Badge | 使用 Badge 组件 |
| 操作 | DropdownMenu `⋯` 按钮 |
| Checkbox | 居中对齐 |

---

## 5. 空表格

```
┌──────────────────────────────────────────────────────────┐
│                                                          │
│                  [📋 图标 48px]                          │
│                                                          │
│               暂无数据                                   │
│          点击 [创建] 按钮添加第一条数据                    │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

---

## 6. 排序

- 点击表头切换排序方向：无 → 升序 → 降序 → 无
- 排序图标动画：`transition-transform duration-200`
- URL 同步：排序参数写入 URL query string

---

## 7. 分页

```
显示 1-20 / 共 50 条          ← 1 2 3 →
```

| 元素 | 样式 |
|------|------|
| 页码信息 | `text-sm text-muted-foreground` |
| 页码按钮 | `h-8 w-8 text-sm`，当前页 `bg-primary text-primary-foreground` |
| 前后箭头 | `variant="outline" size="icon"` |
| 每页条数 | Select: 20 / 50 / 100 |

---

## 8. 批量选择

- 表头 Checkbox：全选/全不选/半选
- 行 Checkbox：单项选择
- 选中后底部浮现批量操作栏（见 Issue 列表）
