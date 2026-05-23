# Form Controls 表单控件

> 基于 shadcn/ui 表单组件 + React Hook Form + Zod

---

## 1. Input 文本输入

### 标准 Input

```
标签 *
┌──────────────────────────────────────┐
│ 占位符文字                            │
└──────────────────────────────────────┘
辅助说明文字
```

| 状态 | 样式 |
|------|------|
| 默认 | `border border-input bg-background rounded-md h-9 px-3 text-sm` |
| Focus | `ring-2 ring-ring ring-offset-1 border-ring` |
| Error | `border-destructive focus-visible:ring-destructive` |
| Disabled | `opacity-50 cursor-not-allowed` |

### 字段布局

```
┌─── 标签区 ────────────────────────────────────┐
│ 标签名 *              (可选) 字段说明          │
└───────────────────────────────────────────────┘
┌─── 输入框 ────────────────────────────────────┐
│                                               │
└───────────────────────────────────────────────┘
┌─── 辅助/错误 ─────────────────────────────────┐
│ 辅助说明 或 错误信息                           │
└───────────────────────────────────────────────┘
```

- 标签：`text-sm font-medium`
- 必填标记：`text-destructive`
- 辅助文字：`text-xs text-muted-foreground`
- 错误信息：`text-xs text-destructive`

### 特殊 Input 变体

| 变体 | 用途 | 扩展样式 |
|------|------|----------|
| Password | 密码输入 | 右侧 `Eye`/`EyeOff` 切换按钮 |
| Search | 搜索框 | 左侧 `Search` 图标，`pl-9` |
| Number | 数字输入 | 右侧 +/- 按钮 |
| Combobox | 可搜索下拉 | Input + Popover + Command List |

---

## 2. Select 下拉选择

```
标签 *
┌──────────────────────────┐
│ 当前选中值             ▾ │
└──────────────────────────┘
```

### Dropdown 面板

```
┌──────────────────────────┐
│ 🔍 搜索...               │
│ ──────────────────────── │
│ ○ 选项一                 │
│ ● 选项二 (当前选中)      │
│ ○ 选项三                 │
│ ○ 选项四                 │
└──────────────────────────┘
```

- 选中项：`bg-accent`
- Hover：`bg-accent/50`
- 分组支持：分组标题 `px-2 py-1.5 text-xs font-semibold text-muted-foreground`
- 最大高度：`max-h-64`，超出滚动

---

## 3. Textarea 多行文本

```
标签
┌──────────────────────────────────────┐
│                                      │
│                                      │
│                                      │
└──────────────────────────────────────┘
字数: 150/10000
```

- 默认 3 行高度，`min-h-[80px]`
- 可配置 `minRows` / `maxRows` 自动扩展
- 底部字数统计 `text-xs text-muted-foreground`

### Auto-resize Textarea (聊天输入)

- 最小 1 行，最大 6 行
- `Enter` 发送，`Shift+Enter` 换行
- 自动高度调整，`overflow-y-auto` 当超过 6 行

---

## 4. Switch 开关

```
标签描述                    [● 开启]
```

| 状态 | 样式 |
|------|------|
| 开启 | `bg-primary`，圆形滑块右移 |
| 关闭 | `bg-input`，圆形滑块左移 |
| Disabled | `opacity-50 cursor-not-allowed` |

- 宽度 `w-10`，高度 `h-5`
- 滑块：`h-4 w-4 rounded-full`，带 transition 200ms

---

## 5. Checkbox 复选框

```
☑ 标签文字
```

| 状态 | 样式 |
|------|------|
| 未选 | `border border-input` |
| 选中 | `bg-primary border-primary` + `Check` 图标 |
| 半选 | `bg-primary` + `Minus` 图标 |
| Disabled | `opacity-50` |

---

## 6. RadioGroup 单选组

```
○ 选项一
● 选项二 (选中)
○ 选项三
```

### 卡片式 RadioGroup (主题选择)

```
┌────────┐ ┌────────┐ ┌────────┐
│  ☀️   │ │  🌙   │ │  💻   │
│ 浅色   │ │ 深色   │ │ 系统   │
│ (选中) │ │        │ │        │
└────────┘ └────────┘ └────────┘
```

- 使用 `RadioGroup` + 自定义样式
- 选中态：`border-primary ring-2 ring-primary/20`
- 卡片内居中排列图标和文字

---

## 7. 表单验证

### Zod Schema 集成

```typescript
const schema = z.object({
  name: z.string().min(2, "名称至少2个字符").max(64, "名称最多64个字符"),
  email: z.string().email("请输入有效的邮箱地址"),
});

type FormValues = z.infer<typeof schema>;
```

### 错误展示

- 错误信息显示在对应字段下方
- `text-xs text-destructive mt-1`
- 提交时一次性验证所有字段
- 实时验证：`mode: "onBlur"` 或 `mode: "onChange"`

### 全局表单规则

| 规则 | 实现 |
|------|------|
| 必填标记 | 标签右侧红色 `*` |
| 可选标记 | 标签右侧灰色 "(可选)" |
| 提交防重 | `isSubmitting` 禁用按钮 |
| 脏值检测 | `form.formState.isDirty` 关闭时提示 |
