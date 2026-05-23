> 来源：`docs/plan/03-client.md` 第 227 行
> 位置：[总目录](../README.md) -> [AnserFlow - Client / Frontend](README.md) -> 客户端（Web SPA）
> 相邻：[上一篇](02-国际化（i18n）.md) · 下一篇：无
> 相关主题：[返回文档入口](README.md) · [国际化（i18n）](02-国际化（i18n）.md)

### 客户端（Web SPA）

客户端统一使用 Next.js 14 SPA（static export），浏览器直接访问，界面以 IM 聊天为核心交互模式。

- **技术栈**：与 admin 一致（Next.js SPA + shadcn/ui + Tailwind CSS + TanStack Query + Zustand）
- **核心路由**：`/client/dashboard`、`/client/projects/:id`、`/client/chat`、`/client/invite/:token`
- **部署方式**：Go embed 嵌入或独立部署，浏览器访问
- **通知方式**：浏览器 Notification API + WebSocket 实时推送

**`/client/chat` IM 四栏布局**：

```
/client/chat                   ← 主聊天页面（四栏布局）
  ① 第一栏：导航栏（窄栏，图标 + 用户头像）
    ├── 用户头像（点击进入个人中心）
    ├── 💬 聊天图标（激活第二栏为聊天视图）
    ├── 📁 项目图标（激活第二栏为项目视图）
    └── ⚙️ 个人中心（设置、通知偏好、退出）

  ② 第二栏：会话列表 / 项目列表
    [聊天视图]
    ├── 搜索/新建双人聊（支持搜索用户和 Agent）
    ├── 双人聊列表项
    │     - 人+人：对方用户头像 + 昵称 + 最后消息 + 未读数
    │     - 人+Agent：Agent 头像 + 名称 + Agent 标识 + 最后消息 + 未读数
    ├── 群聊列表项（群名 + 最后消息 + 未读数）
    │     - 按最后消息时间统一排序（双人聊和群聊混合排列）
    [项目视图]
    ├── 我的项目列表（项目名 + 进度概要）

  ③ 第三栏：聊天窗口
    /client/chat/:group_id     ← 选中会话后展示聊天内容
      - 顶部：会话标题（direct: 对方昵称/Agent名称，从成员信息派生；group: 群名）
      - 中部：消息列表（复用现有 MessageList 组件）
      - 底部：输入框（条件渲染，见下方）

  ④ 第四栏：项目上下文面板
    /client/chat/:group_id/context  ← 当前会话关联的项目信息
      - 项目基本情况（名称、仓库、成员）
      - 关联 Issue 列表（状态 Tab：backlog/todo/in_progress/in_review/done）
      - 参与的 Agent（角色、状态、绑定 Skill）
      - 进度概览（Issue 完成率、PR 状态、最近活动）
```

**Issue 详情页（展开 Issue 行后显示）**：

第四栏的 Issue 列表中，点击某条 Issue 后展开为详情视图，核心是**时间线面板** + **执行控制**：

```
┌─ Issue #42: 实现登录表单组件 ────────────────────────────────────┐
│  状态: in_progress   优先级: P1   负责人: Agent-前端              │
│                                                                  │
│  ┌─ 工具栏 ────────────────────────────────────────────────────┐  │
│  │  [编辑]  [⏸ 暂停]  [⏹ 停止]              筛选: [全部 ▾]    │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌─ 时间线 ────────────────────────────────────────────────────┐  │
│  │                                                              │  │
│  │  12:00  system   状态变更: backlog → todo                    │  │
│  │  12:01  system   状态变更: todo → in_progress                │  │
│  │  12:02  agent    📝 开始编码: 正在读取 Issue 描述...          │  │
│  │  12:05  agent    📄 生成文件: src/login.tsx                   │  │
│  │  12:05  agent    📄 生成文件: src/login.test.tsx              │  │
│  │  12:08  agent    ✅ 运行测试: 3 passed, 1 failed              │  │
│  │  12:08  agent    ❌ FAIL: login.test.tsx > 密码验证           │  │
│  │  12:09  agent    🔧 正在修复 login.test.tsx                   │  │
│  │  12:12  agent    ✅ 运行测试: 4 passed, 0 failed              │  │
│  │  12:12  agent    📦 commit + push → PR #15                   │  │
│  │  12:12  system   状态变更: in_progress → in_review           │  │
│  │  ───────────────── 自动滚动到底部 ─────────────────          │  │
│  │                                                              │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌─ 追加提示词 ────────────────────────────────────────────────┐  │
│  │  [                                ] [发送并重新执行]          │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌─ Token 消耗 ────────────────────────────────────────────────┐  │
│  │  输入: 12,450 tokens  输出: 3,200 tokens  合计: 15,650       │  │
│  └────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

**时间线事件类型与样式**：

| event_type | source | 图标 | 样式 | 说明 |
|------------|--------|------|------|------|
| `status_change` | system | 无 | 灰色斜体 | 状态流转记录，显示 `旧状态 → 新状态` + 触发原因 |
| `agent_log` | agent | 按动作分类 | 默认文本 | 沙箱执行日志，action 决定图标（见下表） |
| `human_prompt` | user | ✏️ | 蓝色高亮 | 人工追加提示词，显示用户名 + 内容 |
| `system_note` | system | 📋 | 灰色 | 编辑/备注等管理操作 |

**agent_log 的 action 分类与图标**：

| action | 图标 | 说明 |
|--------|------|------|
| `generate` | 📄 | 生成文件 |
| `test` / `test_pass` | ✅ | 测试通过 |
| `test_fail` / `fix` | ❌ / 🔧 | 测试失败 / 修复中 |
| `commit` | 📦 | 提交代码 |
| `create_pr` | 🔀 | 创建 PR |
| `error` | ⚠️ | 执行异常 |
| `paused` | ⏸ | 执行暂停 |
| 其他 | 📝 | 通用日志 |

**时间线数据流**：

```
沙箱 anserflow stdout
    │
    ▼
Worker streamLogs() 捕获
    ├── 写入 agent_logs 表（结构化：action / status / output）
    ├── 写入 issue_timeline 表（展示用：source / event_type / content）
    └── WebSocket 推送 {type: "agent_log", text, ts}
            │
            ▼
前端 IssueTimeline 组件
    ├── 历史加载: GET /api/issues/:id/timeline?page=1&size=50
    ├── 实时追加: WebSocket issue:{id} 频道订阅
    └── 筛选过滤: 前端内存过滤（数据量可控，无需服务端筛选）
```

**前端组件结构**：

> 📎 完整代码示例: [reference/frontend-code-examples.md](../../reference/frontend-code-examples.md) §客户端 — IssueDetail 组件 / IssueTimeline 组件

**工具栏按钮状态控制**：

| Issue 状态 | 编辑 | 暂停 | 恢复 | 停止 | 追加提示词 |
|-----------|------|------|------|------|----------|
| `backlog` | ✅ | - | - | - | - |
| `todo` | ✅ | - | - | - | ✅ |
| `in_progress` | - | ✅ | - | ✅ | ✅ |
| `paused` | - | - | ✅ | ✅ | ✅ |
| `in_review` | - | - | - | - | ✅（PR 被拒绝后退回 todo 时复用沙箱） |
| `done` | - | - | - | - | - |

**API 接口**：

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/issues/:id` | GET | Issue 详情（标题/状态/优先级/负责人/Token 统计） |
| `/api/issues/:id/timeline` | GET | 时间线列表（分页，?page=1&size=50） |
| `/api/issues/:id/pause` | POST | 暂停执行 |
| `/api/issues/:id/resume` | POST | 恢复执行 |
| `/api/issues/:id/stop` | POST | 停止执行 |
| `/api/issues/:id/prompt` | POST | 追加人工提示词 |

**direct 类型下的 UI 条件渲染**：

| 场景 | 隐藏 | 显示 |
|------|------|------|
| 人+人（direct, 无 Agent） | @Agent 选择器、/backlog 按钮、Agent 成员头像、成员管理面板 | 纯文本输入框、/new 按钮 |
| 人+Agent（direct, 有 Agent） | @Agent 选择器（只有 1 个 Agent，无需 @）、成员管理面板 | /backlog 按钮、/new 按钮、Agent 头像标识、Agent 回复消息 |
| 群聊（group） | — | 全部功能 |
