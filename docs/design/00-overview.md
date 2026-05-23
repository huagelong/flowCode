# AnserFlow 设计文档总览

> 基于 shadcn/ui + Tailwind CSS 的 AI Agent 协作平台 UI 设计规范

## 项目简介

AnserFlow 是一个 AI Agent + 自然人混合协作平台，核心场景：

- 自然人创建项目 → 组建群组（含 CEO/CTO/前端/后端 Agent 角色）→ 发布需求
- Agent 自动讨论生成方案 → 自动拆解 Issue → 自动执行代码
- 自然人审核验收 → PR 合并完成

## 技术栈

| 层级 | 技术 |
|------|------|
| 框架 | Next.js 14 SPA (static export) |
| UI 组件库 | shadcn/ui + Radix UI |
| 样式 | Tailwind CSS 3.4+ |
| 状态管理 | Zustand + TanStack Query |
| 表单 | React Hook Form + Zod |
| 表格 | TanStack Table |
| 图表 | Recharts |
| 国际化 | next-intl |
| 图标 | lucide-react |
| 主题 | next-themes (Light/Dark/System) |
| 动画 | Framer Motion |
| 通知 | Sonner |

## 文档目录

### 设计基础
- [01-design-tokens.md](01-design-tokens.md) — 设计令牌、颜色系统、间距、排版、动效参数

### 管理后台 (Admin Panel)
路由前缀 `/admin/*`，面向组织管理员和超级管理员。

| 文档 | 页面 |
|------|------|
| [admin/00-navigation.md](admin/00-navigation.md) | 导航布局与侧边栏 |
| [admin/01-login.md](admin/01-login.md) | 登录 / 注册 |
| [admin/02-dashboard.md](admin/02-dashboard.md) | 数据仪表盘 |
| [admin/03-organizations.md](admin/03-organizations.md) | 组织管理 |
| [admin/04-agents.md](admin/04-agents.md) | Agent 管理 |
| [admin/05-projects.md](admin/05-projects.md) | 项目管理 |
| [admin/06-issues.md](admin/06-issues.md) | Issue 跟踪 |
| [admin/07-skills.md](admin/07-skills.md) | Skill 管理 |
| [admin/08-groups.md](admin/08-groups.md) | 群组 / 会话管理 |
| [admin/09-notifications.md](admin/09-notifications.md) | 通知中心 |
| [admin/10-settings.md](admin/10-settings.md) | 系统设置 |
| [admin/11-user-profile.md](admin/11-user-profile.md) | 用户资料 |
| [admin/12-runtimes.md](admin/12-runtimes.md) | Runtime 管理 |
| [admin/13-audit-logs.md](admin/13-audit-logs.md) | 审计日志 |

### 客户端 (Client SPA)
路由前缀 `/client/*`，面向普通成员的四栏 IM 布局。

| 文档 | 页面 |
|------|------|
| [client/00-layout.md](client/00-layout.md) | 四栏 IM 布局 |
| [client/01-chat-list.md](client/01-chat-list.md) | 会话列表 |
| [client/02-chat-window.md](client/02-chat-window.md) | 聊天窗口 |
| [client/03-project-context.md](client/03-project-context.md) | 项目上下文面板 |
| [client/04-issue-detail.md](client/04-issue-detail.md) | Issue 详情 / 时间线 |
| [client/05-project-list.md](client/05-project-list.md) | 项目列表 |
| [client/06-settings.md](client/06-settings.md) | 个人设置 |

### 共享组件 (Components)
两端复用的 UI 组件库。

| 文档 | 组件 |
|------|------|
| [components/00-index.md](components/00-index.md) | 组件索引 & 复用矩阵 |
| [components/01-button.md](components/01-button.md) | Button 按钮变体 |
| [components/02-form-controls.md](components/02-form-controls.md) | Input / Select / Textarea / Switch |
| [components/03-data-table.md](components/03-data-table.md) | DataTable 数据表格 |
| [components/04-avatar.md](components/04-avatar.md) | Avatar 头像 |
| [components/05-badge.md](components/05-badge.md) | Badge 标记 & 状态 |
| [components/06-card.md](components/06-card.md) | Card 卡片容器 |
| [components/07-dialog-modal.md](components/07-dialog-modal.md) | Dialog / Modal |
| [components/08-through-13.md](components/08-through-13.md) | Dropdown / Command / MessageBubble / Timeline / CodeBlock / Chart |
| [components/14-through-20.md](components/14-through-20.md) | Notification / EmptyState / Sidebar / FormBuilder / AgentLog / TokenUsage / Pagination |

### 共享页面 (Shared)
- [shared/00-auth-pages.md](shared/00-auth-pages.md) — 登录 / 注册 / OAuth
- [shared/01-error-pages.md](shared/01-error-pages.md) — 404 / 403 / 500 错误页
- [shared/02-invite-accept.md](shared/02-invite-accept.md) — 邀请接受页

## 设计原则

1. **一致性** — 同一组件在 Admin 和 Client 端行为一致，仅布局不同
2. **高效性** — 减少点击层级，高频操作一步直达
3. **实时感** — WebSocket 驱动的状态更新需即时反映在 UI 上
4. **信息密度** — Admin 端偏向信息密度（表格式），Client 端偏向对话流
5. **渐进披露** — 复杂功能按需展示（展开行、侧边面板、Popover）

## 响应式断点

遵循 shadcn/ui / Tailwind 默认断点：

| 断点 | 宽度 | 场景 |
|------|------|------|
| `sm` | ≥640px | 大屏手机横屏 |
| `md` | ≥768px | 平板竖屏 |
| `lg` | ≥1024px | 平板横屏 / 小笔记本 |
| `xl` | ≥1280px | 标准桌面 |
| `2xl` | ≥1536px | 大屏桌面 |

- **Admin 端**：最低 `lg (1024px)`，不主动适配移动端
- **Client 端**：完整响应式，移动端采用折叠面板方案
