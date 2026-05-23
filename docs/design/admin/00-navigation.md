# Admin 导航布局

> 路由前缀 `/admin/*`，经典侧边栏 + 内容区布局。

---

## 1. 整体布局结构

```
┌──────────────────────────────────────────────────────────────┐
│ Top Bar (h-14, sticky)                              [User] 🔔 │
├──────────┬───────────────────────────────────────────────────┤
│ Sidebar  │                                                   │
│ (w-64)   │                                                   │
│          │  Main Content Area                                │
│ Collaps- │  (scrollable, p-6)                                │
│ ible to  │                                                   │
│ w-16     │                                                   │
│          │                                                   │
│          │                                                   │
│          │                                                   │
│          │                                                   │
│          │                                                   │
└──────────┴───────────────────────────────────────────────────┘
```

### 布局参数

| 属性 | 值 |
|------|------|
| 侧边栏展开宽度 | `w-64` (256px) |
| 侧边栏折叠宽度 | `w-16` (64px) |
| 顶部栏高度 | `h-14` (56px) |
| 内容区内边距 | `p-6` (24px) |
| 最小内容宽度 | `min-w-[768px]` |
| 内容最大宽度 | 无限制（fluid） |

---

## 2. Top Bar

```
┌───────────────────────────────────────────────────────────────┐
│ [≡] AnserFlow        [组织选择器 ▾]      🔔(3)  [头像▾]      │
└───────────────────────────────────────────────────────────────┘
```

### 组成部分

| 区域 | 组件 | 说明 |
|------|------|------|
| 左侧 | Sidebar Toggle + Logo | `≡` 按钮折叠/展开侧边栏，"AnserFlow" 品牌名 |
| 中左 | Org Selector | Dropdown 选择当前组织，显示组织名 + 角色 badge |
| 右侧 | Notification Bell | 铃铛图标 + 红色未读数 badge，点击进入通知页 |
| 最右 | User Avatar Dropdown | 头像 + 下拉菜单（用户资料、主题切换、退出登录） |

### 组织选择器 Dropdown

```
┌─────────────────────────┐
│ 切换组织                 │
├─────────────────────────┤
│ ✓ My Organization  [owner] │
│   Acme Corp        [admin] │
│   Team Alpha       [member]│
├─────────────────────────┤
│ + 创建新组织              │
└─────────────────────────┘
```

- 当前组织显示 `✓` 标记
- 每项显示组织名 + 用户角色 Badge
- 底部「创建新组织」入口

---

## 3. Sidebar

### 展开态 (Expanded)

```
┌──────────────────┐
│ 📊 Dashboard      │
│ 🏢 Organizations  │
│ 🤖 Agents         │
│ 📁 Projects       │
│ 💬 Groups         │
│ ⚡ Skills          │
│ ──────────────── │
│ ⚙️ Settings       │ (super_admin only)
│ 🔧 Runtimes       │ (super_admin only)
│ ──────────────── │
│ 📋 Audit Logs     │ (super_admin only)
└──────────────────┘
```

### 折叠态 (Collapsed)

```
┌──────┐
│  📊  │
│  🏢  │
│  🤖  │
│  📁  │
│  💬  │
│  ⚡  │
│ ──── │
│  ⚙️  │
│  🔧  │
│ ──── │
│  📋  │
└──────┘
```

### 导航项规范

| 属性 | 展开态 | 折叠态 |
|------|--------|--------|
| 高度 | `h-9` (36px) | `h-10` (40px) |
| 左内边距 | `pl-3` | 居中 |
| 圆角 | `rounded-md` | `rounded-md` |
| 选中态背景 | `bg-accent` | `bg-accent` |
| 选中态文字 | `text-accent-foreground font-medium` | 同左 |
| Hover | `bg-accent/50` | `bg-accent/50` |
| 图标尺寸 | `h-4 w-4` | `h-5 w-5` |
| 文字与图标间距 | `ml-3` | — |
| Tooltip（折叠态） | — | 右侧 Tooltip 显示文字 |

### 路由映射

| 导航项 | 路由 | 图标 (lucide) |
|--------|------|---------------|
| Dashboard | `/admin/dashboard` | `LayoutDashboard` |
| Organizations | `/admin/organizations` | `Building2` |
| Agents | `/admin/agents` | `Bot` |
| Projects | `/admin/projects` | `FolderKanban` |
| Groups | `/admin/groups` | `MessageSquare` |
| Skills | `/admin/skills` | `Zap` |
| Settings | `/admin/settings` | `Settings` |
| Runtimes | `/admin/runtimes` | `Wrench` |
| Audit Logs | `/admin/audit-logs` | `ScrollText` |

### 权限控制

- **所有登录用户可见**: Dashboard, Organizations, Agents, Projects, Groups, Skills
- **仅 super_admin 可见**: Settings, Runtimes, Audit Logs
- 未授权的菜单项不渲染（不是 disable）

---

## 4. 内容区框架

### 页面标题区

```
┌──────────────────────────────────────────────────────┐
│ [← 返回]  页面标题                     [操作按钮]     │
│ 描述文字（可选）                                      │
│ ─────────────────────────────────────────────────────│
│ [Tab 1] [Tab 2] [Tab 3]                 (可选 Tab 栏)│
└──────────────────────────────────────────────────────┘
```

| 元素 | 规范 |
|------|------|
| 标题 | `text-2xl font-semibold tracking-tight` |
| 描述 | `text-sm text-muted-foreground`，标题下方 `mt-1` |
| 返回按钮 | 仅子页面显示，`< ChevronLeft` + "返回" |
| 操作按钮区 | `flex gap-2` 右对齐，主按钮用 `variant="default"` |
| Tab 栏 | 使用 shadcn/ui `Tabs`，与标题间距 `mt-4` |

### 空状态

当页面无数据时显示 EmptyState 组件：
- 图标 (64px) + 标题 + 描述 + 操作按钮
- 例：无 Agent → "还没有 Agent" + "创建你的第一个 AI Agent" + [创建 Agent]

---

## 5. 响应式行为

Admin 端最低支持 `lg (1024px)` 宽度：

| 宽度 | 侧边栏 | 内容区 |
|------|--------|--------|
| < 1024px | 不支持，提示使用桌面浏览器 | — |
| 1024–1279px | 默认折叠 (`w-16`) | 自适应 |
| ≥1280px | 默认展开 (`w-64`) | 自适应 |

用户点击 `≡` 可手动切换，偏好存入 `localStorage`。

---

## 6. 动效

- 侧边栏折叠/展开：`width` 过渡 200ms ease-in-out
- 路由切换：页面内容 `opacity 0→1` + `translateY 8→0` 过渡 150ms
- 导航项 hover：背景色 150ms transition
