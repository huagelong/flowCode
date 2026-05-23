# Admin 通知中心

> 路由：`/admin/notifications`
> API：`/api/notifications/*`

---

## 1. 通知列表页

```
┌──────────────────────────────────────────────────────────────────┐
│ 通知中心                             [全部标为已读]              │
│                                                                  │
│ 筛选 [全部 ▾]  [未读]  [Issue]  [Agent]  [邀请]  [提及]         │
│                                                                  │
│ ┌────────────────────────────────────────────────────────────┐  │
│ │ ● 🤖 Agent "前端" 完成了 Issue #42                        │  │
│ │   2026-05-20 14:30                          [查看 →]      │  │
│ │ ────────────────────────────────────────────────────────── │  │
│ │ ● 👤 张三 在讨论组中 @提及了你                             │  │
│ │   2026-05-20 11:20                          [查看 →]      │  │
│ │ ────────────────────────────────────────────────────────── │  │
│ │ ○ 📦 Issue #39 状态变更: 进行中 → 审核中                  │  │
│ │   2026-05-19 18:00                          [查看 →]      │  │
│ │ ────────────────────────────────────────────────────────── │  │
│ │ ○ 🎫 你被邀请加入 "Acme Corp" 组织                        │  │
│ │   2026-05-18 09:00                          [接受 →]      │  │
│ └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│ ← 1 2 3 →                                                       │
└──────────────────────────────────────────────────────────────────┘
```

### 通知项样式

| 属性 | 未读 | 已读 |
|------|------|------|
| 左侧圆点 | `●` `bg-primary` (6px) | `○` `bg-muted` |
| 背景 | `bg-primary/5` | `bg-transparent` |
| 文字 | `text-foreground font-medium` | `text-foreground` |
| 时间 | `text-xs text-muted-foreground` | 同左 |

### 通知类型 & 图标

| type | 图标 (lucide) | 说明 |
|------|---------------|------|
| issue_assigned | `UserPlus` | Issue 分配给你 |
| issue_status_changed | `RefreshCw` | Issue 状态变更 |
| plan_generated | `ClipboardList` | Agent 生成方案，需你确认 |
| agent_completed | `Bot` + `CheckCircle2` | Agent 执行完成 |
| agent_failed | `Bot` + `XCircle` | Agent 执行失败 |
| mention | `AtSign` | @提及 |
| dm_message | `MessageCircle` | 私信消息 |
| invite | `Ticket` | 组织邀请 |

### 筛选器

| 选项 | 筛选逻辑 |
|------|----------|
| 全部 | 无筛选 |
| 未读 | `is_read=false` |
| 待确认 | `type=plan_generated` 且未确认 |
| Issue | `type` 包含 `issue` 或 `plan_generated` |
| Agent | `type` 包含 `agent` |
| 邀请 | `type=invite` |
| 提及 | `type=mention` |

### 操作

- **点击通知**: 标记已读 + 跳转到对应资源页
- **"查看 →"**: 同上
- **"全部标为已读"**: `PUT /api/notifications/read-all` + 刷新列表
- **删除**: 不提供删除，通知为持久记录

---

## 2. 顶部通知铃铛

位于 Top Bar 右侧：

```
     🔔(3)
```

### 样式

| 属性 | 值 |
|------|------|
| 图标 | `Bell` (lucide)，`h-5 w-5` |
| 未读 Badge | 红色圆点 `bg-destructive text-destructive-foreground`，绝对定位右上 |
| Badge 数字 | `text-[10px] font-medium`，>99 显示 "99+" |
| 位置 | 相对于图标 `top-[-4px] right-[-4px]` |

### 点击行为

点击铃铛打开 Popover：

```
┌──────────────────────────────────┐
│ 通知 (4 条未读)       [全部已读] │
│ ──────────────────────────────── │
│ ● 📋 CTO 生成了方案 #42，待确认 │
│   2 分钟前                       │
│ ──────────────────────────────── │
│ ● 🤖 前端Agent 完成 Issue #42   │
│   5 分钟前                       │
│ ──────────────────────────────── │
│ ● 张三 @提及了你                 │
│   1 小时前                       │
│ ──────────────────────────────── │
│ ● Issue #39 状态变更             │
│   昨天 18:00                     │
│ ──────────────────────────────── │
│               [查看全部通知 →]   │
└──────────────────────────────────┘
```

- Popover 宽度 `w-80`
- 最多显示 5 条最新通知
- "查看全部通知" 跳转到 `/admin/notifications`

### 实时更新

- WebSocket 订阅 `user:{id}` 频道
- 收到 `native_notification` 事件：更新 Badge 数 + Popover 列表
- 使用 Sonner 显示桌面通知 toast

---

## 3. WebSocket 推送 & 浏览器通知

### 通知渠道优先级

1. **WebSocket**: 始终推送（在线时）
2. **Browser Notification**: `Notification.requestPermission()` 获取权限后推送
3. **Sonner Toast**: 页面内显示

### Sonner Toast 样式

```tsx
// Agent 完成通知
toast.success("Agent 完成了任务", {
  description: "前端Agent 完成了 Issue #42: 修复登录Bug",
  action: { label: "查看", onClick: () => navigate("/admin/projects/1/issues/42") },
  duration: 5000,
});

// 方案待确认通知
toast.info("方案待确认", {
  description: "CTO 在讨论中生成了方案 #42: 修复登录页面验证逻辑",
  action: { label: "去确认", onClick: () => navigate("/admin/groups/5") },
  duration: 0, // 不自动关闭，直到用户操作
});

// Agent 失败通知
toast.error("Agent 执行失败", {
  description: "后端Agent 执行 Issue #39 失败: 超时",
  action: { label: "查看", onClick: () => navigate("/admin/projects/1/issues/39") },
  duration: 8000,
});

// 邀请通知
toast.info("收到组织邀请", {
  description: "你被邀请加入 Acme Corp",
  duration: 5000,
});
```

---

## 4. 空状态

```
        [🔔 图标 64px]
     没有新通知
 当有新消息时，这里会显示通知
```
