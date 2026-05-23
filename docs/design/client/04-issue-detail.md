# Client Issue 详情 / 时间线

> 位于上下文面板或独立页面
> API：`/api/orgs/:org_id/projects/:project_id/issues/:issue_id/timeline`

---

## 1. 时间线组件 (紧凑版)

### 1.1 视觉规范

```
12:12 📦 状态变更: 进行中 → 审核中
12:10 🤖 📦 commit + push → PR #15
12:08 🤖 ✅ Running tests: 4 passed, 0 failed
12:05 🤖 📄 Generated file: src/login.tsx
12:02 🤖 📝 Starting: 读取 Issue 描述...
12:01 📦 状态变更: todo → 进行中
12:00 📦 状态变更: backlog → todo
```

### 1.2 事件行样式

| 属性 | 值 |
|------|------|
| 时间 | `font-mono text-[10px] text-muted-foreground w-12 shrink-0` |
| 来源图标 | `h-3.5 w-3.5` |
| 内容 | `text-xs` |
| 间距 | `py-1 px-2` |
| 分隔线 | 每行下方 `border-b border-border/50` |

### 1.3 事件类型样式

#### 状态变更 (status_change)

```
12:12 📦 进行中 → 审核中
```

- `text-muted-foreground italic text-xs`
- 状态名带对应颜色
- 图标：`RefreshCw h-3 w-3 text-muted-foreground`

#### Agent 日志 (agent_log)

```
12:08 🤖 ✅ Running tests: 4 passed
```

- `text-xs text-foreground`
- 动作图标 + 颜色见下表

| action | 图标 | 颜色 |
|--------|------|------|
| generate | `FileCode2` | `text-blue-500` |
| test_pass | `CheckCircle2` | `text-green-500` |
| test_fail | `XCircle` | `text-red-500` |
| fix | `Wrench` | `text-amber-500` |
| commit | `GitCommit` | `text-violet-500` |
| create_pr | `GitPullRequest` | `text-purple-500` |
| error | `AlertTriangle` | `text-red-500` |
| paused | `Pause` | `text-orange-500` |
| default | `FileText` | `text-muted-foreground` |

#### 人工指令 (human_prompt)

```
12:15 👤 💬 请确保密码强度校验包含特殊字符
```

- `text-xs text-blue-700 bg-blue-50 rounded px-2 py-1`
- 左侧蓝色竖线 `border-l-2 border-blue-500`
- 图标：`MessageSquare text-blue-500`

#### 系统备注 (system_note)

```
12:00 📦 任务已创建
```

- `text-xs text-muted-foreground italic`
- 图标：`Info text-muted-foreground`

---

## 2. 筛选按钮

```
[全部] [agent_log (8)] [system (3)] [human (1)]
```

- 使用 shadcn/ui `ToggleGroup`
- 每个按钮显示类型名 + 数量
- 选中态：`bg-accent text-accent-foreground`
- 筛选实时生效，无分页（前端过滤）

---

## 3. 实时更新

- WebSocket 订阅 `issue:{issue_id}` 频道
- 收到新事件：
  1. 追加到时间线底部
  2. 自动滚动到底部（如果已在底部）
  3. 如果已向上滚动，显示 "↓ 新事件 (3)" 按钮
- 收到 `status_change`：
  1. 更新顶部状态显示
  2. 更新操作按钮

---

## 4. 加载策略

| 场景 | 行为 |
|------|------|
| 初始加载 | 最近 50 条事件 |
| 向上滚动 | 加载更多（每页 50） |
| 实时事件 | WebSocket 追加 |

---

## 5. 执行控制按钮

根据 Issue 状态动态显示：

| 当前状态 | 按钮 | 图标 | 颜色 |
|----------|------|------|------|
| in_progress | 暂停 | `Pause` | `variant="outline"` |
| in_progress | 停止 | `Square` | `variant="destructive"` |
| paused | 恢复 | `Play` | `variant="default"` |
| paused | 停止 | `Square` | `variant="destructive"` |
| in_review | 查看 PR | `ExternalLink` | `variant="outline"` |
| done | 查看 PR | `ExternalLink` | `variant="outline"` |

按钮大小：`h-8 text-xs`（紧凑版）

---

## 6. Token 用量展示

```
Token 用量
────────────
输入   12,450
输出    3,200
────────────
总计   15,650
费用   $0.15
```

- 数字：`font-mono text-xs`
- 费用：`text-xs font-medium`
- 分隔线：`border-t border-border`
