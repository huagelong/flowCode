# 组件索引 & 复用矩阵

> Admin 和 Client 共享的 UI 组件库，基于 shadcn/ui 扩展。

---

## 组件清单

| # | 组件 | 基于 | Admin 使用 | Client 使用 |
|---|------|------|-----------|-------------|
| 01 | Button | shadcn/ui `Button` | ✅ 全局 | ✅ 全局 |
| 02 | Form Controls | shadcn/ui `Input/Select/Textarea/Switch` | ✅ 全局表单 | ✅ 设置页 |
| 03 | DataTable | shadcn/ui `Table` + TanStack Table | ✅ 列表页 | ❌ |
| 04 | Avatar | shadcn/ui `Avatar` + 自定义 | ✅ 成员/Agent | ✅ 聊天/会话 |
| 05 | Badge | shadcn/ui `Badge` + 自定义变体 | ✅ 状态/角色/来源 | ✅ 状态/角色 |
| 06 | Card | shadcn/ui `Card` | ✅ Dashboard/详情 | ✅ 项目卡片 |
| 07 | Dialog / Modal | shadcn/ui `Dialog` + `AlertDialog` | ✅ 创建/编辑/删除 | ✅ 群组创建 |
| 08 | Dropdown | shadcn/ui `DropdownMenu` | ✅ 操作菜单 | ✅ 更多操作 |
| 09 | Command | shadcn/ui `Command` | ✅ 全局搜索 | ✅ @提及/搜索 |
| 10 | MessageBubble | 自定义 | ❌ (只读) | ✅ 聊天窗口 |
| 11 | Timeline | 自定义 | ✅ Issue 详情 | ✅ Issue 详情 |
| 12 | CodeBlock | 自定义 | ❌ | ✅ Agent 代码输出 |
| 13 | Chart | Recharts 封装 | ✅ Dashboard/Token | ❌ |
| 14 | Notification | Sonner + 自定义 | ✅ 全局 | ✅ 全局 |
| 15 | EmptyState | 自定义 | ✅ 空列表 | ✅ 空聊天 |
| 16 | Sidebar | 自定义 | ✅ 导航栏 | ❌ (用 Nav Bar) |
| 17 | FormBuilder | 自定义 (JSON Schema → 表单) | ✅ Agent Runtime 配置 | ❌ |
| 18 | AgentLog | 自定义 | ✅ 日志列表 | ✅ 时间线 |
| 19 | TokenUsage | 自定义 | ✅ Agent 详情 | ✅ Issue 详情 |
| 20 | Pagination | shadcn/ui `Pagination` | ✅ 列表页 | ✅ 消息历史 |

---

## 复用策略

### 全局共享组件
以下组件在 Admin 和 Client 之间完全共享，无差异：
- Button, Badge, Avatar, Dialog, Dropdown, Command, EmptyState, Notification, Pagination

### 端特定变体
部分组件在两端有不同变体：

| 组件 | Admin 变体 | Client 变体 |
|------|-----------|-------------|
| Card | Dashboard 统计卡片 | 项目列表卡片 |
| Timeline | 完整版 (展开面板) | 紧凑版 (侧边面板) |
| Avatar | 32px (表格) + 40px (详情) | 36px (消息) + 40px (会话列表) + 24px (面板) |
| DataTable | 完整分页表格 | 不使用 |
| MessageBubble | 只读列表 | 可交互 (发送/命令) |

### 不共享组件
| 组件 | 仅 Admin | 仅 Client |
|------|----------|-----------|
| Sidebar | ✅ 左侧导航 | ❌ |
| DataTable | ✅ | ❌ |
| FormBuilder | ✅ | ❌ |
| Chart | ✅ | ❌ |
| CodeBlock | ❌ | ✅ |
| MessageBubble (交互版) | ❌ | ✅ |
