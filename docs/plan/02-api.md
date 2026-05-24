# AnserFlow - API / Backend

---

## 框架补充说明

> 以下为生产级 Gin 项目的标准配套设施，确保系统可维护、可观测、可扩展。
>
> **实现代码**：本文档中的实现级代码示例已外提至 [backend-code-examples.md](../../reference/backend-code-examples.md)，文中通过链接引用。

| 框架/库 | 用途 | 配置要点 |
|---------|------|----------|
| **Viper** (`spf13/viper`) | 配置管理 | `config.yaml` + `AutomaticEnv()` 环境变量覆盖 |
| **Zap** (`uber-go/zap`) | 结构化日志 | JSON 格式输出，GORM 接入 Zap 作为日志后端 |
| **validator** (`go-playground/validator`) | 请求校验 | Gin 原生集成，struct tag 声明规则（`binding:"required,min=1"`） |
| **Casbin** (`casbin/casbin`) | RBAC 权限 | 策略模型存 MySQL，支持组织/项目/Issue 级权限 |
| **Swagger** (`swaggo/swag`) | API 文档 | 代码注解 → OpenAPI 3.0，开发环境 `/swagger/index.html` |
| **CORS** (`gin-contrib/cors`) | 跨域支持 | 白名单 `AllowOrigins`，SPA 跨域访问 API |
| **优雅关闭** | Go `signal` + `http.Server.Shutdown` | SIGINT/SIGTERM 时完成进行中请求，10s 超时 |
| **健康检查** | `/api/health` 端点 | 返回 DB + Redis 连通性，供 K8s/Docker 探活 |

#### OAuth2 — 第三方登录（GitHub）

AnserFlow 支持 GitHub OAuth 登录，降低注册门槛。前端 `login` 页面提供「GitHub 登录」按钮，跳转到 GitHub 授权页；回调后后端完成账号创建/绑定并返回 JWT。

```
用户点击 GitHub 登录
        │
        ▼
GET /api/auth/github/login  → 302 跳转 GitHub
        │
        ▼
用户授权 → GitHub 回调 GET /api/auth/github/callback?code=xxx
        │
        ▼
后端：code → access_token → 获取 GitHub 用户信息
        │
        ▼
┌─────────────────────────────────────────┐
│ github_id 已存在?                        │
│  ├── 是 → 直接生成 JWT，登录成功         │
│  └── 否 → 已登录用户？（绑定）            │
│           ├── 是 → 绑定 github_id 到账户 │
│           └── 否 → 自动注册新用户         │
└─────────────────────────────────────────┘
        │
        ▼
返回 JWT → 前端存储 → 跳转到 /admin/dashboard
```

**非 GitHub 用户的注册兼容**：支持传统邮箱+密码注册（通过 `/api/auth/register` / `/api/auth/login`），与 OAuth 用户共用 `users` 表，`password_hash` 为 NULL 表示仅 OAuth 登录。

**实现代码**: [backend-code-examples.md §OAuth2](../../reference/backend-code-examples.md#oauth2--github-第三方登录)

#### 后端：go-i18n 错误码映射实现

后端 API 返回统一的国际化错误码，前端根据当前 locale 映射为对应语言文案。前端 i18n（next-intl）方案详见 [03-client.md §国际化](03-client.md)。

**错误码枚举**（完整清单）：

| 错误码 | HTTP 状态码 | 说明 |
|--------|-----------|------|
| `ERR_ISSUE_NOT_FOUND` | 404 | Issue 不存在 |
| `ERR_PROJECT_NOT_FOUND` | 404 | 项目不存在 |
| `ERR_AGENT_NOT_FOUND` | 404 | Agent 不存在 |
| `ERR_ORG_NOT_FOUND` | 404 | 组织不存在 |
| `ERR_SKILL_NOT_FOUND` | 404 | Skill 不存在 |
| `ERR_USER_NOT_FOUND` | 404 | 用户不存在 |
| `ERR_VALIDATION_FAILED` | 400 | 请求参数校验失败 |
| `ERR_UNAUTHORIZED` | 401 | 未认证（JWT 过期/无效） |
| `ERR_PERMISSION_DENIED` | 403 | 权限不足 |
| `ERR_ORG_LIMIT_EXCEEDED` | 403 | 组织并发 Agent 数已达上限 |
| `ERR_SANDBOX_TIMEOUT` | 500 | Docker 沙箱执行超时 |
| `ERR_SANDBOX_CREATE_FAILED` | 500 | Docker 沙箱创建失败 |
| `ERR_GIT_CLONE_FAILED` | 500 | Git 仓库克隆失败 |
| `ERR_GIT_PUSH_FAILED` | 500 | Git 推送失败 |
| `ERR_LLM_API_ERROR` | 502 | LLM API 调用失败 |
| `ERR_INVITE_EXPIRED` | 400 | 邀请链接已过期 |
| `ERR_INVITE_MAX_USES` | 400 | 邀请链接已达使用上限 |
| `ERR_RUNTIME_NOT_FOUND` | 404 | 运行时不存在 |
| `ERR_INTERNAL` | 500 | 服务器内部错误 |

**Gin 错误响应中间件**：

**实现代码**: [backend-code-examples.md §i18n 中间件](../../reference/backend-code-examples.md#gin-错误响应中间件)

**翻译文件示例**：

**翻译文件**: [backend-code-examples.md §翻译文件](../../reference/backend-code-examples.md#翻译文件示例)

> **前端消费**：前端 `apiFetch` 封装层拦截 APIError，根据 `code` 从 `messages/{locale}.json` 的 `Errors` 段读取对应文案作为 fallback。后端错误码优先显示，前端翻译仅为降级兜底。

---

## 一、分布式架构设计

### 1.1 WebSocket 分布式方案

#### 1.1.1 架构拓扑

```
                    ┌─────────────┐
                    │  Redis       │
                    │  Pub/Sub     │
                    └──┬───┬───┬──┘
                       │   │   │
          ┌────────────┘   │   └────────────┐
          ▼                ▼                ▼
    ┌──────────┐    ┌──────────┐    ┌──────────┐
    │ Gin #1   │    │ Gin #2   │    │ Gin #3   │
    │ WS连接池 │    │ WS连接池 │    │ WS连接池 │
    │ Hub      │    │ Hub      │    │ Hub      │
    └──────────┘    └──────────┘    └──────────┘
```

**原理**：每个 Gin 实例维护自己的 WebSocket Hub（连接池）。消息发送时，先推送给本地连接的客户端，再通过 Redis Pub/Sub 广播到其他实例，各实例转发给自己持有的客户端。

**Go 库**：`github.com/gorilla/websocket` + `github.com/redis/go-redis/v9`

#### 1.1.2 连接建立与频道订阅

WebSocket 连接建立时仅需认证，**不绑定任何频道**。连接建立后，客户端通过 `subscribe` 消息动态订阅感兴趣的资源频道：

**连接地址**：`/ws?token=xxx`（仅认证，不含 group_id）

**频道类型**：

| 频道格式 | 说明 | 推送内容 |
|---------|------|----------|
| `group:{id}` | 群聊/双人聊 | 聊天消息、指令响应、系统通知 |
| `issue:{id}` | Issue 时间线 | agent_log、status_change、执行控制 |
| `project:{id}` | 项目通知 | Issue 状态变更摘要、调度通知 |
| `user:{id}` | 用户私信 | 个人通知（浏览器通知、@提及等） |

**订阅协议**：

```json
// 客户端订阅
{ "type": "subscribe", "channels": ["group:42", "issue:7", "project:3", "user:1"] }

// 客户端取消订阅
{ "type": "unsubscribe", "channels": ["issue:7"] }

// 服务端确认
{ "type": "subscribe_ack", "channels": ["group:42", "issue:7", "project:3", "user:1"] }
```

> **设计说明**：连接建立后默认自动订阅 `user:{自身ID}` 频道，其余频道需显式订阅。前端路由切换时（如从群聊页进入 Issue 详情页）发送 `subscribe` / `unsubscribe` 动态调整。

**Hub 内部路由**：

```go
// internal/ws/hub.go
type Hub struct {
    // 连接 → 已订阅频道集合
    connChannels map[*Conn]map[string]bool
    // 频道 → 已订阅连接集合
    channelConns map[string]map[*Conn]bool
    // Redis Pub/Sub 订阅
    redisSub *redis.PubSub
}

// Subscribe 添加频道订阅
func (h *Hub) Subscribe(c *Conn, channels []string) {
    for _, ch := range channels {
        h.connChannels[c][ch] = true
        h.channelConns[ch][c] = true
    }
    // 如果是本实例首次订阅该频道，加入 Redis Pub/Sub
    h.ensureRedisSubscription(channels)
}

// SendToChannel 向指定频道推送消息（核心方法）
func (h *Hub) SendToChannel(channel string, msg interface{}) {
    // 1. 本地推送
    for conn := range h.channelConns[channel] {
        conn.WriteJSON(msg)
    }
    // 2. Redis Pub/Sub 广播到其他实例
    h.redis.Publish(context.Background(), "ws:"+channel, msg)
}

// 便捷方法：基于 SendToChannel 派生
func (h *Hub) SendToGroup(groupID uint, msg interface{}) {
    h.SendToChannel(fmt.Sprintf("group:%d", groupID), msg)
}
func (h *Hub) SendToIssue(issueID uint, msg interface{}) {
    h.SendToChannel(fmt.Sprintf("issue:%d", issueID), msg)
}
func (h *Hub) SendToProject(projectID uint, msg interface{}) {
    h.SendToChannel(fmt.Sprintf("project:%d", projectID), msg)
}
func (h *Hub) SendToUser(userID uint, msg interface{}) {
    h.SendToChannel(fmt.Sprintf("user:%d", userID), msg)
}
```

> **权限校验**：`Subscribe` 时需校验用户是否有权订阅该频道（如 `group:42` 需要用户是该群成员，`issue:7` 需要用户是关联项目的成员）。校验失败返回 `subscribe_nack`。

#### 1.1.3 消息协议

所有 WebSocket 通信统一采用 JSON 信封格式：

```json
{
  "type": "string",       // 消息类型（见下表）
  "seq": 12345,           // 消息序号（服务端分配，全局递增）
  "ts": 1715678900,       // 服务端时间戳（Unix 秒）
  "channel": "group:42",  // 所属频道（group/issue/project/user）
  "sender": {
    "type": "user|agent",
    "id": 1,
    "name": "张三",
    "avatar_url": "..."
  },
  "content": {},           // 消息体（字段见下表）
  "error": null            // 错误信息（仅 type:error）
}
```

**消息类型全景**：

| type | 方向 | 说明 | content 关键字段 | 持久化 |
|------|------|------|-----------------|--------|
| `message` | C→S / S→C | 普通聊天消息 | `text: string` | ✅ messages 表 |
| `annotation` | S→C | Agent 分析/技术评审（非对话消息） | `text: string`, `role: "analysis"\|"review"\|"estimate"` | ✅ messages 表 |
| `backlog` | S→C | 需求产出（`/backlog` 指令触发，Issue 状态=backlog） | `issue: {title, status: "backlog", priority, assignee, description}` | ✅ messages 表 |
| `todo` | S→C | 任务列表产出（从 backlog 分析出的子 Issue，状态=todo） | `issues: [{title, status: "todo", parent_id, priority, assignee, description}]` | ✅ messages 表 |
| `review_decision` | C→S | 审核中人工决策 | `issue_id: string`, `decision: "approved"\|"changes_requested"`, `comment: string` | ✅ issue_timeline 表 |
| `system` | S→C | 系统通知（Issue 创建/状态变更） | `text: string`, `resource_type: "issue"\|"agent"`, `resource_id` | ✅ messages 表 |
| `agent_log` | S→C | 沙箱执行日志（推送到 issue 频道） | `text: string`, `ts: int64` | ✅ agent_logs + issue_timeline 表 |
| `status_change` | S→C | 执行控制状态变更（暂停/恢复/停止） | `status: string`, `hint: string` | ✅ issue_timeline 表 |
| `agent_assign` | S→C | Agent @其他 Agent 布置任务 | `target_agent_id: int`, `task: string`, `context: string` | ❌ 不持久化（走 message） |
| `new_session` | S→C | 自然人 `/new` 开启新会话 | `session_id: string` | ✅ messages 表 |
| `subscribe` | C→S | 订阅频道 | `channels: string[]` | ❌ |
| `subscribe_ack` | S→C | 订阅确认 | `channels: string[]` | ❌ |
| `unsubscribe` | C→S | 取消订阅 | `channels: string[]` | ❌ |
| `ping` | C→S | 心跳请求（客户端每 30s 发送） | 无 | ❌ |
| `pong` | S→C | 心跳响应 | 无 | ❌ |
| `typing` | C→S / S→C | 正在输入状态 | `is_typing: bool` | ❌ |
| `error` | S→C | 错误响应 | `error.code: string`, `error.message: string` | ❌ |
| `command` | S→C | 指令消息回显（用户输入的 /xxx 指令原文） | `text: string`, `command: string` | ✅ messages 表 |
| `native_notification` | S→C | 浏览器原生通知触发 | `title: string`, `body: string`, `channel: string` | ❌ |

**`/todo` 任务列表生成指令**：群内自然人发送 `/todo` 后，anserAgent 从当前 backlog 需求与群聊讨论上下文中分析任务列表，创建一组 `status=todo` 的子 Issue，并通过 `parent_id` 归属到同一个 backlog Issue 下。该步骤不需要人工确认；生成后的任务可被调度器按优先级进入 `in_progress`。

**@Agent 任务布置**：群内 Agent 在讨论或执行过程中，可根据其他 Agent 的角色定义（`agents.system_prompt` + 绑定的 Skills），通过 `@AgentName` 语法向指定 Agent 布置子任务。被 @ 的 Agent 接收消息后，由 anserAgent 根据其角色人设和记忆系统自动生成响应或执行操作。典型场景：CTO Agent 在讨论中说 `@前端Agent 你负责登录页 UI 实现`，系统解析 `@前端Agent` 匹配群内 Agent 成员，将任务描述注入该 Agent 的上下文。

**`/new` 新会话指令**：群内自然人发送 `/new` 后，系统在当前群组内创建一个新的会话上下文（`session_id`）。新会话之前的消息不再作为 Agent 讨论的上下文窗口内容，Agent 仅感知 `/new` 之后的消息历史。这使自然人可以在同一群组内切换讨论主题，避免上下文混淆和 Token 浪费。`/new` 不清除历史消息（历史仍可滚动查看），仅重置 Agent 上下文窗口的起点。

#### 1.1.4 心跳与重连

- 客户端每 30s 发送 `ping`，服务端回复 `pong`
- 90s 内未收到任何消息视为断连，服务端主动关闭连接并清理所有频道订阅
- 客户端重连采用指数退避：`1s → 2s → 4s → 8s → 16s → 32s (max)`
- 重连后客户端需重新发送 `subscribe` 恢复频道订阅，并携带 `last_seq` 请求遗漏消息

#### 1.1.5 消息持久化规则

| 类别 | 持久化目标 | 说明 |
|------|-----------|------|
| `message` / `system` / `annotation` / `backlog` / `todo` / `new_session` / `command` | `messages` 表 | 聊天记录、系统通知、方案产出、指令回显 |
| `agent_log` | `agent_logs` + `issue_timeline` 表 | 沙箱执行日志（不写入 messages 表） |
| `status_change` | `issue_timeline` 表 | 执行控制事件（暂停/恢复/停止） |
| `typing` / `ping` / `pong` / `subscribe` / `unsubscribe` / `native_notification` | 不持久化 | 瞬时状态/控制信令 |

> **原则**：需要历史回看的消息写 `messages` 表；仅 Issue 维度的执行记录写 `agent_logs` / `issue_timeline` 表；瞬时状态和信令不持久化。

**Hub 消息路由 — Agent 编排判断**：

Hub 的 `OnMessage` 入口统一根据 `HasAgentMember()` 决定是否触发 anserAgent 编排，`commandHandler` 独立调用以确保 `/new` 全模式可用（`/backlog` 和 `/todo` 仅含 Agent 时可用）：

**Hub 消息路由实现**: [backend-code-examples.md §WebSocket](../../reference/backend-code-examples.md#websocket-分布式架构)

> **设计说明**：
> - `HasAgentMember()` 查询 `group_members` 表中是否存在 `member_type = 'agent'` 的记录，结果可在 Hub 连接生命周期内缓存，无需每次消息都查库。
> - **指令消息以 `command` 类型广播后立即返回**，避免同时触发 anserAgent 编排导致 Agent 重复响应。CommandHandler 内部根据 `HasAgentMember()` 决定 `/backlog` 和 `/todo` 是否可用，但 `/new` 在所有模式下均可用（会话上下文隔离对所有场景都有意义）。
> - `seq` 由服务端统一分配（Redis INCR `ws:seq:global`），客户端不参与序号生成，避免多客户端 seq 冲突。

**CommandHandler 实现**: [backend-code-examples.md §WebSocket](../../reference/backend-code-examples.md#websocket-分布式架构)

**Agent 编排判断规则**：

`HasAgentMember()` 是决定是否触发 anserAgent 编排的唯一条件，与 `group.type` 无关：

| 约束 | 说明 |
|------|------|
| `HasAgentMember()` 决定 anserAgent 编排和 /backlog、/todo | 群聊和双人聊共享同一套逻辑，不按 type 分支 |
| `/new` 全模式可用 | CommandHandler 独立于 HasAgentMember()，在 Hub 层直接调用 |
| `/todo` 仅含 Agent 时可用 | 需要 anserAgent 基于 backlog 分析任务列表，生成 `status=todo` 的子 Issue |
| `/backlog` 仅含 Agent 时可用 | 需要 anserAgent 编排产出需求 Issue，状态为 backlog |
| 指令消息不触发 anserAgent 编排 | Hub 检测到指令后 `return`，避免 CommandHandler 内部编排和 anserAgent 双重触发 |
| 无 Agent 时不触发 `MentionResolver` | 群聊无 Agent 时同样跳过，双人聊天然无 @场景 |
| direct 成员不可变更 | Handler 层对 direct 类型返回 400（group 类型不受此限制） |

**`review_decision` 审核闭环**：

用户只在 `in_review` 状态对 PR 进行人工审核。Hub 收到 `review_decision` 消息后触发状态流转：

**review_decision 处理实现**: [backend-code-examples.md §WebSocket](../../reference/backend-code-examples.md#websocket-分布式架构)

> **前端交互**：Agent 执行通过后进入 `in_review`，群聊中展示 PR 卡片 + [审核通过] [退回修改] 按钮。审核通过等待 GitHub merged webhook 后进入 `done`；退回修改则回到 `todo` 并保留任务上下文。

**Redis 消息缓存**（断线重连恢复）：

断线重连时客户端通过 `seq` 号请求遗漏消息。为减少 MySQL 查询压力，在 Redis 中维护每个频道的最近消息滑动窗口：

**Redis 消息缓存实现**: [backend-code-examples.md §WebSocket](../../reference/backend-code-examples.md#websocket-分布式架构)

| 参数 | 值 | 理由 |
|------|-----|------|
| 缓存条目数 | 最近 500 条/频道 | 覆盖正常离线窗口 5-10min |
| TTL | 24 小时（每次写入续期） | 活跃频道保持缓存，冷频道自动过期释放内存 |
| 数据结构 | Redis ZSET（seq → JSON） | 按 seq 范围查询 O(logN+ M)，比 List 更适合续传场景 |
| 内存估算 | 500 条 × 200 频道 × 2KB ≈ 200MB | 生产环境足够 |
| seq 生成 | Redis INCR `ws:seq:global` | 服务端统一分配，全局递增，避免多客户端冲突 |

**重连流程**：

```
客户端断线重连
    → 建立 WS 连接（/ws?token=xxx）
    → 发送 subscribe 恢复频道订阅
    → 发送 {type: "resync", channels: [{channel: "group:42", last_seq: 1230}]}
    → 服务端从 Redis ZSET 拉取 seq > 1230 的消息（最多 500 条）
    → Redis 命中 → 直接返回（~2ms）
    → Redis 未命中（冷频道或超 500 条）→ 回退 MySQL 查询
```

### 1.2 任务队列方案

选用 **Asynq**（https://github.com/hibiken/asynq），基于 Redis 的分布式任务队列：

```
Issue 状态变为 in_progress (assignee = agent)
        │
        ▼
┌──────────────────┐
│ Asynq Client      │  →  enqueue("agent:execute", payload)
│ (Gin HTTP 层)    │     Priority: P0 > P1 > P2...
└──────────────────┘     Timeout: 30min
        │                MaxRetry: 3
        ▼                Payload: {issue_id, agent_id, human_prompts[]}
┌──────────────────┐
│ Redis Queue       │
│ ├── critical (P0) │
│ ├── default  (P1) │
│ └── low      (P2+)│
└──────────────────┘
        │
        ▼
┌──────────────────┐
│ Asynq Worker      │  →  HandleFunc("agent:execute", handler)
│ (独立进程/协程)   │     1. 创建 Docker 沙箱
└──────────────────┘     2. 注入 anserAgent 配置 + Agent 人设
        │                3. 注入补充指令（来自 issue_timeline）
        ▼                4. anserAgent run 执行编码
┌──────────────────┐     5. anserAgent 检查结果
│ Docker Sandbox    │     6. 通过 → commit + push + PR → in_review
└──────────────────┘     7. 失败 → 写入时间线 → 回到 todo 等待重试
```

Asynq 核心特性：

| 特性 | Agent 执行场景 |
|------|---------------|
| 任务优先级 | P0 Issue 插队执行 |
| 重试机制 | 执行失败自动重试 3 次 |
| 超时控制 | 单任务最长 30 分钟 |
| 死信队列 | 3 次重试仍失败 → 保持 todo 并标记为升级处理 |
| 定时任务 | 延迟执行（Agent 启动冷却期） |
| Web UI | Asynqmon 可视化管理面板 |

**Issue 调度器**（todo → in_progress 自动调度）：

系统的调度器作为一个轻量的 Gin 后台协程运行，与 Asynq Worker 解耦：

**调度器实现**: [backend-code-examples.md §任务队列](../../reference/backend-code-examples.md#任务队列asynq)

> 每个组织默认最多 3 个 Agent 同时执行（可通过 org settings 调整），超过上限的 Issue 保持 todo 等待。

并发统计直接查询 `issues` 表，无需额外 Redis 计数器：

**并发统计实现**: [backend-code-examples.md §任务队列](../../reference/backend-code-examples.md#任务队列asynq)

#### 调度器对 paused 状态的处理

调度器扫描可调度 Issue 时**明确跳过 `paused` 状态**，避免重复入队：

**调度器查询实现**: [backend-code-examples.md §任务队列](../../reference/backend-code-examples.md#任务队列asynq)

> 调度器只扫 `todo`，不会扫到 `paused`。`paused` 状态的 Issue 由 Worker 心跳循环自行管理，与调度器完全解耦。

#### Asynq 任务状态与 Issue 状态双向同步

| 事件 | Issue 状态 | Asynq 任务状态 | 说明 |
|------|-----------|--------------|------|
| 调度器入队 | `todo → in_progress` | Enqueued | 写入 Redis queue |
| Worker 消费 | `in_progress` | Active（已 dequeue） | 不在 Redis 队列中 |
| 人工暂停 | `in_progress → paused` | Active（Worker 心跳等待） | Docker 容器冻结 |
| 人工恢复 | `paused → in_progress` | Active（Worker 心跳跳出） | Docker 容器解冻 |
| 停止执行 | `in_progress/paused → todo` | 终止（goroutine 退出） | worktree 保留，anserAgent 进程终止 |
| anserAgent 成功 | `in_progress → in_review` | Completed | 任务返回 nil |
| anserAgent 失败 | `in_progress → todo` | Failed → 重新入队（retry_count+1） | 保留沙箱 |
| 重试耗尽 | `todo`（标记升级处理） | Archived（死信队列） | 保留 Issue 与 parent backlog 关系 |

#### 调度器无限重试防护

为防止配置错误的 Issue 无限循环，在 `issues` 表增加 `retry_count` 字段：

```sql
-- issues 表新增字段
ALTER TABLE issues ADD COLUMN retry_count INT DEFAULT 0;
-- 重试次数仅在重新生成任务列表或手动重新执行时重置为 0
```

**调度器过滤**（已在 `FindSchedulableIssues` 中加入 `WHERE retry_count < 3`）。

**执行失败时递增**：

**重试处理实现**: [backend-code-examples.md §任务队列](../../reference/backend-code-examples.md#任务队列asynq)

**重试重置**（仅当重新生成任务列表或手动重新执行时重置）：

**重试重置实现**: [backend-code-examples.md §任务队列](../../reference/backend-code-examples.md#任务队列asynq)

> **防护效果**：同一 Issue 最多经历 3 次自动重试循环（`todo → in_progress → 失败 → todo`），第 4 次保持 `todo` 并标记升级处理，不回退到 `backlog`，避免把任务错误地变回需求。

### 1.3 整体分布式拓扑

```
                    ┌──────────────┐
                    │  MySQL       │
                    └──────┬───────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
  ┌─────┴─────┐      ┌─────┴─────┐      ┌─────┴─────┐
  │ Gin #1    │      │ Gin #2    │      │ Gin #3    │
  │ :8080     │      │ :8081     │      │ :8082     │
  │ WS Hub    │      │ WS Hub    │      │ WS Hub    │
  └─────┬─────┘      └─────┬─────┘      └─────┬─────┘
        │                  │                  │
        └──────────────────┼──────────────────┘
                           │
                    ┌──────┴───────┐
                    │  Redis       │
                    │  ├─ Pub/Sub  │ (WS 跨实例广播)
                    │  ├─ Queue    │ (Asynq 任务队列)
                    │  └─ Cache    │ (会话/热点数据)
                    └──────┬───────┘
                           │
                    ┌──────┴───────┐
                    │  Worker #1   │ (Asynq Worker)
                    │  Worker #2   │
                    │  Worker #N   │
                    │  Docker SDK  │
                    └──────────────┘
```

---

---

## 二、核心数据模型

### 2.0 角色与权限管理（RBAC）

AnserFlow 采用双层 RBAC 模型：**系统级角色** + **组织级角色**，由 Casbin 统一管理策略，MySQL 存储策略表，运行时动态加载。

#### 角色体系

```
┌─────────────────────────────────────────────────┐
│  系统级 (System-Level)                           │
│  ┌───────────────────────────────────────────┐  │
│  │  super_admin  平台超级管理员                │  │
│  │  • 管理所有组织/用户/Agent                  │  │
│  │  • 系统配置（邮件/LLM/存储）                │  │
│  │  • 查看审计日志                            │  │
│  └───────────────────────────────────────────┘  │
├─────────────────────────────────────────────────┤
│  组织级 (Organization-Level)                     │
│  ┌──────────────┐ ┌──────────────┐             │
│  │ owner        │ │ admin        │  member     │
│  │ • 完全控制    │ │ • 管理资源   │  • 只读协作 │
│  │ • 删除组织    │ │ • 邀请成员   │  • 查看     │
│  │ • 转移所有权  │ │ • 管理项目   │  • 评论     │
│  │ • 所有CRUD   │ │ • 管理Agent  │  • 创建Issue│
│  └──────────────┘ └──────────────┘             │
└─────────────────────────────────────────────────┘
```

| 层级 | 角色 | 权限范围 | 典型用户 |
|------|------|---------|----------|
| 系统 | `super_admin` | 全平台 | 平台运营者 |
| 组织 | `owner` | 单个组织完全控制 | 组织创建者 |
| 组织 | `admin` | 单个组织管理 | 团队负责人 |
| 组织 | `member` | 单个组织协作 | 普通成员 |

#### 权限矩阵

Casbin 使用 `(sub, obj, act)` 模型：`主体 + 资源 + 操作`。

**Casbin 配置**: [backend-code-examples.md §RBAC](../../reference/backend-code-examples.md#rbac-权限管理)

| 资源 (obj) | 操作 (act) | owner | admin | member |
|-----------|-----------|-------|-------|--------|
| `org` | `read` / `update` / `delete` | ✅ | ❌ | ❌ |
| `org` | `read` | ✅ | ✅ | ✅ |
| `member` | `invite` / `remove` / `update_role` | ✅ | ✅ | ❌ |
| `member` | `list` | ✅ | ✅ | ✅ |
| `project` | `create` / `update` / `delete` | ✅ | ✅ | ❌ |
| `project` | `read` / `list` | ✅ | ✅ | ✅ |
| `issue` | `create` / `update` / `delete` | ✅ | ✅ | ✅(仅自己) |
| `issue` | `read` / `list` | ✅ | ✅ | ✅ |
| `issue` | `assign` / `change_status` | ✅ | ✅ | ❌ |
| `agent` | `create` / `update` / `delete` | ✅ | ✅ | ❌ |
| `agent` | `read` / `list` | ✅ | ✅ | ✅ |
| `skill` | `create` / `update` / `delete` | ✅ | ✅ | ❌ |
| `skill` | `read` / `list` | ✅ | ✅ | ✅ |
| `group` | `create` / `manage` | ✅ | ✅ | ❌ |
| `group` | `read` / `send_message` | ✅ | ✅ | ✅ |
| `direct` | `create` | ✅ | ✅ | ✅ |
| `direct` | `read` / `send_message` | ✅ | ✅ | ✅ |
| `webhook` | `manage` | ✅ | ✅ | ❌ |
| `settings` | `manage` | ✅ | ✅ | ❌ |

#### 数据库设计

Casbin 策略存储在 `casbin_rules` 表中，完整 DDL + 预置策略见 [ddl.sql](../ddl.sql#认证与权限)。

#### Go 中间件集成

Casbin 作为 Gin 中间件，在每个 API 请求前校验权限：

```go
// internal/middleware/rbac.go
package middleware

import (
    "net/http"
    "strconv"

    "github.com/casbin/casbin/v2"
    gormadapter "github.com/casbin/gorm-adapter/v3"
    "github.com/gin-gonic/gin"
)

var enforcer *casbin.Enforcer

func InitRBAC(dsn string) error {
    adapter, err := gormadapter.NewAdapter("mysql", dsn)
    if err != nil {
        return err
    }
    enforcer, err = casbin.NewEnforcer("config/rbac_model.conf", adapter)
    if err != nil {
        return err
    }
    return enforcer.LoadPolicy()
}

// RequirePermission 中间件：检查当前用户是否有资源操作权限
// 用法: r.POST("/api/orgs/:org_id/projects", RequirePermission("project", "create"))
func RequirePermission(obj, act string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetInt64("user_id")        // JWT 中间件注入
        orgID := c.Param("org_id")             // 路径参数

        sub := buildSubject(userID, orgID)      // "user:123@org:1"

        ok, err := enforcer.Enforce(sub, obj, act)
        if err != nil || !ok {
            c.JSON(http.StatusForbidden, gin.H{
                "code":    "ERR_PERMISSION_DENIED",
                "message": "权限不足",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}

// 构建主体标识
func buildSubject(userID int64, orgID string) string {
    return "user:" + strconv.FormatInt(userID, 10) + "@org:" + orgID
}

// GetUserRole 获取用户在组织中的角色
func GetUserRole(userID int64, orgID uint) string {
    roles, _ := enforcer.GetRolesForUser(
        buildSubject(userID, strconv.Itoa(int(orgID))),
    )
    for _, role := range roles {
        switch {
        case role == "org_role:owner":
            return "owner"
        case role == "org_role:admin":
            return "admin"
        case role == "org_role:member":
            return "member"
        }
    }
    return ""
}
```

#### 路由权限配置

```go
// internal/handler/router.go
func SetupRoutes(r *gin.Engine) {
    api := r.Group("/api")
    api.Use(middleware.JWTAuth())           // ① 先鉴权（JWT）

    // ── 组织管理（仅 owner） ──
    org := api.Group("/orgs/:org_id")
    {
        org.PUT("", middleware.RequirePermission("org", "update"))
        org.DELETE("", middleware.RequirePermission("org", "delete"))

        // ── 成员管理（owner / admin） ──
        org.POST("/members/invite", middleware.RequirePermission("member", "invite"))
        org.DELETE("/members/:user_id", middleware.RequirePermission("member", "remove"))

        // ── 项目管理 ──
        org.POST("/projects", middleware.RequirePermission("project", "create"))
        project := org.Group("/projects/:project_id")
        {
            project.PUT("", middleware.RequirePermission("project", "update"))
            project.DELETE("", middleware.RequirePermission("project", "delete"))

            // Issue (member 可创建/编辑自己的)
            project.POST("/issues", middleware.RequirePermission("issue", "create"))
            project.PUT("/issues/:issue_id", middleware.RequirePermission("issue", "update"))
        }

        // ── Agent 管理（owner / admin） ──
        org.POST("/agents", middleware.RequirePermission("agent", "create"))
        org.PUT("/agents/:agent_id", middleware.RequirePermission("agent", "update"))
        org.DELETE("/agents/:agent_id", middleware.RequirePermission("agent", "delete"))

        // ── 组织设置 ──
        org.PUT("/settings", middleware.RequirePermission("settings", "manage"))
    }
}
```

#### 前端权限控制

```tsx
// packages/shared-ui/src/lib/use-permission.ts
import { useQuery } from '@tanstack/react-query'

interface UserRole {
  orgRole: 'owner' | 'admin' | 'member' | ''
  isSuperAdmin: boolean
}

// 获取当前用户在指定组织中的角色
function useOrgRole(orgId: string): UserRole {
  return useQuery({
    queryKey: ['user-role', orgId],
    queryFn: () => fetch(`/api/orgs/${orgId}/my-role`).then(r => r.json()),
    staleTime: 5 * 60 * 1000,
  }).data ?? { orgRole: '', isSuperAdmin: false }
}

// 权限检查 Hook
export function useCan(orgId: string, action: string): boolean {
  const { orgRole, isSuperAdmin } = useOrgRole(orgId)

  if (isSuperAdmin) return true

  const permissions: Record<string, string[]> = {
    'project:create': ['owner', 'admin'],
    'project:delete': ['owner', 'admin'],
    'member:invite': ['owner', 'admin'],
    'org:delete':    ['owner'],
    'settings:manage':['owner', 'admin'],
  }

  return permissions[action]?.includes(orgRole) ?? false
}
```

```tsx
// 条件渲染按钮
import { useCan } from '@/lib/use-permission'

function ProjectHeader({ orgId }: { orgId: string }) {
  const canCreate = useCan(orgId, 'project:create')

  return (
    <div>
      {canCreate && (
        <Button onClick={openCreateDialog}>创建项目</Button>
      )}
    </div>
  )
}
```

#### 权限变更流程

```
┌─────────┐    ┌──────────┐    ┌──────────────┐
│ 操作者   │    │  API      │    │  Casbin       │
│ (owner)  │    │  Service  │    │  (MySQL)      │
└────┬────┘    └────┬─────┘    └──────┬───────┘
     │              │                 │
     │ PUT /members/3/role          │
     │ body: {"role": "admin"}      │
     │─────────────>│                │
     │              │                │
     │              │ ① 校验操作者    │
     │              │   是 org owner  │
     │              │                │
     │              │ ② 更新 members  │
     │              │   表 role 字段  │
     │              │                │
     │              │ ③ 修改 Casbin   │
     │              │   g 策略:       │
     │              │   user:3 →      │
     │              │   org_role:admin│
     │              │───────────────>│
     │              │                │
     │              │ ④ LoadPolicy() │
     │              │   即时生效      │
     │<─────────────│                │
     │  200 OK      │                │
```

#### 数据一致性保障

`members.role` 字段作为冗余缓存，必须与 Casbin g 策略保持同步：

```go
// internal/service/member_service.go
func (s *MemberService) UpdateRole(ctx context.Context, orgID, userID uint, newRole string) error {
    // ① 更新 members.role 冗余字段
    if err := s.memberRepo.UpdateRole(ctx, orgID, userID, newRole); err != nil {
        return err
    }

    // ② 更新 Casbin g 策略（删除旧角色 + 添加新角色）
    subject := fmt.Sprintf("user:%d@org:%d", userID, orgID)
    // 先删除该用户在组织中的所有旧角色
    s.enforcer.RemoveFilteredGroupingPolicy(0, subject)
    // 添加新角色
    s.enforcer.AddGroupingPolicy(subject, "org_role:"+newRole)

    // ③ Casbin 重新加载策略（当前实例即时生效）
    s.enforcer.LoadPolicy()

    // ④ Redis Pub/Sub 通知其他实例重新加载
    s.redis.Publish(ctx, "casbin:policy_changed", fmt.Sprintf("%d:%d:%s", orgID, userID, newRole))

    return nil
}
```

**定期一致性校验**（每 30 分钟，防止异常导致的 drift）：

```go
// internal/service/consistency_checker.go
func (c *ConsistencyChecker) Run(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Minute)
    for range ticker.C {
        c.checkMembersRoles(ctx)
    }
}

func (c *ConsistencyChecker) checkMembersRoles(ctx context.Context) {
    // 全量比对 members.role vs Casbin g 策略
    members := c.memberRepo.FindAll(ctx)
    for _, m := range members {
        casbinRoles := c.enforcer.GetRolesForUser(
            fmt.Sprintf("user:%d@org:%d", m.UserID, m.OrgID),
        )
        expectedRole := ""
        for _, r := range casbinRoles {
            switch {
            case strings.Contains(r, "owner"):
                expectedRole = "owner"
            case strings.Contains(r, "admin"):
                expectedRole = "admin"
            case strings.Contains(r, "member"):
                expectedRole = "member"
            }
        }
        // 不一致时以 Casbin 为准修复 members 表
        if expectedRole != "" && m.Role != expectedRole {
            c.memberRepo.UpdateRole(ctx, m.OrgID, m.UserID, expectedRole)
            c.logger.Warn("role drift fixed",
                zap.Uint("org_id", m.OrgID),
                zap.Uint("user_id", m.UserID),
                zap.String("was", m.Role),
                zap.String("now", expectedRole),
            )
        }
    }
}
```

> **冲突仲裁**：以 Casbin 策略为准。`members.role` 只是冗余缓存，仅用于前端快速展示角色（避免每次都查 Casbin），不参与权限判断。

### 2.1 ER 关系图

```mermaid
erDiagram
    User ||--o{ Member : "属于"
    User ||--o{ IssueAssignee : "被分配"
    User ||--o{ Invitation : "生成"
    User ||--o{ InvitationUsage : "接受邀请"
    User ||--o{ Notification : "接收"
    User ||--o{ AuditLog : "操作"
    User ||--o{ GroupMember : "参与"
    User ||--o{ Todo : "被分配/创建"
    User ||--|| UserSettings : "偏好设置"
    Organization ||--o{ Member : "拥有"
    Organization ||--o{ Project : "拥有"
    Organization ||--o{ Agent : "拥有"
    Organization ||--o{ Invitation : "关联"
    Organization ||--o{ AuditLog : "审计范围"
    Organization ||--o{ Notification : "通知范围"
    Project ||--o{ Issue : "包含"
    Project ||--o{ Group : "关联（仅 group 类型）"
    Project ||--o{ Todo : "包含"
    Organization ||--o{ Group : "关联（direct 类型不绑定项目）"
    Issue ||--o{ Issue : "父子关系"
    Issue ||--o{ IssueAssignee : "分配"
    Issue ||--o{ IssueTimeline : "时间线"
    Issue ||--o{ AgentLog : "产生日志"
    Group ||--o{ GroupMember : "包含"
    Group ||--o{ Message : "包含"
    Group ||--o{ AgentLog : "产生日志"
    Group ||--o{ GroupReadState : "已读状态"
    Agent ||--o{ GroupMember : "参与"
    Agent ||--o{ AgentSkill : "绑定"
    Agent ||--o{ IssueAssignee : "被分配"
    Agent ||--o{ AgentLog : "执行日志"
    Agent ||--o{ Todo : "被分配"
    Skill ||--o{ AgentSkill : "属于"
    Runtime ||--o{ RuntimeSkill : "默认Skills"
    Skill ||--o{ RuntimeSkill : "属于"
    Invitation ||--o{ InvitationUsage : "记录接受"
    Todo ||--o{ Issue : "关联同步"
```

### 2.2 关键表映射

> 完整建表语句 + 种子数据见 **[ddl.sql](../ddl.sql)**。以下为表结构速查。

| 表名 | 说明 | 关键字段 |
|------|------|---------|
| `users` | 用户（自然人） | id, username, email, github_id, locale, is_super_admin |
| `organizations` | 组织 | id, name, owner_id → users |
| `members` | 组织成员（与 Casbin 双重管理） | org_id, user_id, role(owner/admin/member) |
| `runtimes` | Agent 运行时注册表 | name, docker_image, execute_template, config_schema(JSON) |
| `agents` | Agent 定义 | org_id, name, role_label, system_prompt, runtime_id, runtime_config(JSON) |
| `agent_logs` | Agent 执行日志 | agent_id, issue_id, group_id, type, action, token_usage(JSON) |
| `skills` | Skills 定义 | org_id, name, definition(Markdown), source_type, file_tree(JSON) |
| `agent_skills` | Agent ↔ Skill 绑定 | agent_id, skill_id, enabled |
| `runtime_skills` | Runtime ↔ Skill 默认绑定 | runtime_id, skill_id, enabled |
| `projects` | 项目 | org_id, git_platform, git_repo_url, git_auth_credential, sandbox_container_id |
| `issues` | Issue 需求/任务统一表 | project_id, parent_id, status, priority, source_group_id, pr_url |
| `issue_assignee` | Issue 分配（1 Issue = 1 assignee） | issue_id(UNIQUE), user_id, agent_id |
| `issue_timeline` | Issue 状态时间线 | issue_id, actor_type, event_type, old_status, new_status, comment |
| `groups` | 会话组（group/direct） | org_id, type, project_id, name |
| `group_members` | 群成员 | group_id, user_id, agent_id |
| `messages` | 群消息 | group_id, session_id, sender_type, content, mention_agent_id |
| `invitations` | 组织邀请 | org_id, token(UNIQUE), invite_type, email, max_uses |
| `invitation_usages` | 邀请接受记录 | invitation_id, user_id, assigned_role |
| `audit_logs` | 审计日志 | org_id, user_id, action, resource_type, resource_id, detail(JSON) |
| `notifications` | 通知 | user_id, type, title, is_read, push_channel |
| `user_settings` | 用户偏好 | user_id(UNIQUE), theme, notify_* flags |
| `group_read_state` | 会话已读状态 | group_id + user_id(UNIQUE), last_read_message_id |
| `casbin_rules` | Casbin 策略表 | ptype(p/g), v0(sub), v1(obj), v2(act) |

**表总数**: 22 张（含 Casbin 策略表）

> **Issue 与 Todo 的关系**：`backlog`、`todo`、`in_progress`、`in_review`、`done` 都是 `issues.status` 的不同状态，不再引入独立 `todos` 表。`backlog` 表示需求；`todo` 表示从 backlog 分析出的任务 Issue；同一个 backlog 的任务通过 `issues.parent_id` 放在该 backlog Issue 下。只有 `in_review` 需要人工审核，此时关联/创建 PR；`done` 表示 PR 已完成。

### 2.3 邀请机制说明

**两种邀请方式**：

```
┌─────────────────────────────────────────────┐
│  分享链接邀请                                │
│  ┌───────────────────────────────────────┐  │
│  │ 管理员生成链接                         │  │
│  │ → POST /api/invitations               │  │
│  │   { type: "link", role: "member" }    │  │
│  │                                       │  │
│  │ → 返回: https://xxx/invite/abc123     │  │
│  │                                       │  │
│  │ 目标用户访问链接 → 注册/登录 →        │  │
│  │ 自动加入组织                           │  │
│  └───────────────────────────────────────┘  │
├─────────────────────────────────────────────┤
│  邮箱邀请                                    │
│  ┌───────────────────────────────────────┐  │
│  │ 管理员输入邮箱                         │  │
│  │ → POST /api/invitations               │  │
│  │   { type: "email", email: "u@x.com" } │  │
│  │                                       │  │
│  │ → 系统发送邮件（含邀请链接）            │  │
│  │ → 目标用户点击链接 → 注册/登录 →       │  │
│  │ 自动加入组织                           │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

**安全控制**：

| 机制 | 说明 |
|------|------|
| Token 唯一 | 64 位随机字符串，不可猜测 |
| 过期时间 | 默认 7 天，管理员可设置 |
| 使用次数限制 | `max_uses` 控制，0=不限 |
| 角色预分配 | 受邀进入组织时自动分配角色 |
| 邮箱验证 | 邮箱邀请时验证邮箱归属 |

### 2.4 邮件服务

邮件发送采用 `gopkg.in/gomail.v2`，通过 SMTP 发送邀请邮件和系统通知。

```go
import "gopkg.in/gomail.v2"

func SendInviteEmail(to string, inviteLink string) error {
    m := gomail.NewMessage()
    m.SetHeader("From", "noreply@anserflow.io")
    m.SetHeader("To", to)
    m.SetHeader("Subject", "您被邀请加入 AnserFlow 组织")
    m.SetBody("text/html", fmt.Sprintf(`
        <p>点击以下链接接受邀请：</p>
        <a href="%s">%s</a>
        <p>链接 7 天内有效</p>
    `, inviteLink, inviteLink))

    d := gomail.NewDialer("smtp.example.com", 587, "username", "password")
    return d.DialAndSend(m)
}
```

**SMTP 配置**（存储在 config.yaml 中）：

| 配置项 | 说明 | 示例 |
|--------|------|------|
| `smtp.host` | SMTP 服务器地址 | `smtp.gmail.com` |
| `smtp.port` | 端口 | `587` |
| `smtp.username` | 发件账号 | `noreply@anserflow.io` |
| `smtp.password` | 授权码/密码 | — |
| `smtp.from` | 发件人显示名 | `"AnserFlow <noreply@anserflow.io>"` |
| `smtp.ssl` | 是否 SSL | `false` (STARTTLS) |

**邮件触发场景**：

| 场景 | 邮件内容 |
|------|---------|
| 邮箱邀请 | 含邀请链接，引导注册/登录后自动入组织 |
| Issue 状态变更 | 当 Issue 从 InReview→Done 或被退回时通知相关人 |
| Agent 执行完成 | PR 已提交 / 执行失败 通知 |
| 密码重置 | 密码重置链接 |

**双语邮件模板**：邮件服务根据用户语言偏好发送对应语言版本：

```go
// internal/email/sender.go
func (s *Sender) SendInviteEmail(
    to string,
    inviterName string,
    orgName string,
    inviteLink string,
    locale string, // "zh-CN" | "en-US"
) error {
    m := gomail.NewMessage()
    m.SetHeader("From", s.from)
    m.SetHeader("To", to)

    switch locale {
    case "zh-CN":
        m.SetHeader("Subject", fmt.Sprintf("%s 邀请你加入 AnserFlow 组织「%s」", inviterName, orgName))
        m.SetBody("text/html", fmt.Sprintf(`
            <h2>你被邀请加入组织</h2>
            <p><strong>%s</strong> 邀请你加入 <strong>%s</strong> 组织。</p>
            <p><a href="%s" style="display:inline-block;padding:12px 24px;background:#4F46E5;color:white;border-radius:6px;text-decoration:none;">接受邀请</a></p>
            <p style="color:#6B7280;">链接 7 天内有效</p>
        `, inviterName, orgName, inviteLink))
    default: // en-US
        m.SetHeader("Subject", fmt.Sprintf("%s invited you to join %s on AnserFlow", inviterName, orgName))
        m.SetBody("text/html", fmt.Sprintf(`
            <h2>You've been invited</h2>
            <p><strong>%s</strong> has invited you to join <strong>%s</strong>.</p>
            <p><a href="%s" style="display:inline-block;padding:12px 24px;background:#4F46E5;color:white;border-radius:6px;text-decoration:none;">Accept Invitation</a></p>
            <p style="color:#6B7280;">Link expires in 7 days</p>
        `, inviterName, orgName, inviteLink))
    }

    d := gomail.NewDialer(s.host, s.port, s.username, s.password)
    return d.DialAndSend(m)
}

func (s *Sender) SendAgentNotification(
    to string,
    agentName string,
    issueTitle string,
    success bool,
    locale string,
) error {
    // 类似双语模板切换逻辑
    // ...
}
```

> **locale 来源**：`users.locale` 字段（注册时根据浏览器语言设置，可在个人设置中修改）。未登录用户（邮箱邀请）默认按邀请人 locale 发送。

### 2.5 API 路由总览

所有 API 挂载在 `/api` 下，需要认证的端点由 JWT 中间件保护（标注 `🔒`），敏感操作额外受 Casbin RBAC 约束（标注 `🔐`）。

```
/api
├── /health                                GET  → 健康检查
│
├── /auth                                  认证模块
│   ├── /register                          POST → 邮箱注册
│   ├── /login                             POST → 邮箱登录 → JWT
│   ├── /github/login                      GET  → GitHub OAuth 入口
│   ├── /github/callback                   GET  → GitHub OAuth 回调
│   └── /me                                GET  → 🔒 当前用户信息
│
├── /orgs                                  组织模块（🔒）
│   ├── /                                  GET  → 我加入的组织列表
│   ├── /                                  POST → 创建组织
│   ├── /:org_id                           GET  → 组织详情
│   ├── /:org_id                           PUT  → 🔐 更新组织
│   ├── /:org_id                           DELETE → 🔐 删除组织
│   ├── /:org_id/my-role                   GET  → 当前用户角色
│   ├── /:org_id/dashboard                 GET  → 仪表盘聚合数据（Issue分布/Agent活跃度/项目概览）
│   ├── /:org_id/members                   GET  → 成员列表
│   ├── /:org_id/members/invite            POST → 🔐 邀请成员
│   ├── /:org_id/members/:user_id          DELETE → 🔐 移除成员
│   ├── /:org_id/members/:user_id/role     PUT  → 🔐 修改角色
│   │
│   ├── /:org_id/invitations              邀请模块（🔒）
│   │   ├── /                              POST → 🔐 创建邀请（link/email）
│   │   └── /:token/accept                 POST → 接受邀请
│   │
│   ├── /:org_id/settings                  GET/PUT → 🔐 组织设置
│   │
│   ├── /:org_id/projects                 项目管理（🔒）
│   │   ├── /                              GET  → 项目列表
│   │   ├── /                              POST → 🔐 创建项目
│   │   ├── /:project_id                   GET  → 项目详情
│   │   ├── /:project_id                   PUT  → 🔐 更新项目
│   │   ├── /:project_id                   DELETE → 🔐 删除项目
│   │   │
│   │   ├── /:project_id/issues           Issue 管理
│   │   │   ├── /                          GET  → Issue 列表
│   │   │   ├── /                          POST → 🔒 创建 Issue
│   │   │   ├── /:issue_id                 GET  → Issue 详情
│   │   │   ├── /:issue_id                 PUT  → 🔒 更新 Issue
│   │   │   ├── /:issue_id                 DELETE → 🔐 删除 Issue
│   │   │   ├── /:issue_id/status          PUT  → 🔒 状态流转
│   │   │   ├── /:issue_id/assign          PUT  → 🔐 分配/更改负责人
│   │   │   ├── /:issue_id/children        GET  → 子 Issue 列表
│   │   │   ├── /batch-analyze               POST → 🔒 批量分析 backlog，生成 todo 子 Issue
│   │   │   ├── /:issue_id/timeline         GET  → 状态时间线（含补充指令历史）
│   │   │   ├── /:issue_id/prompt           POST → 🔒 追加补充指令（重新触发执行）
│   │   │   ├── /:issue_id/pause            POST → 🔒 暂停执行（冻结沙箱）
│   │   │   ├── /:issue_id/resume           POST → 🔒 恢复执行（解冻沙箱）
│   │   │   ├── /:issue_id/stop             POST → 🔒 停止执行（终止沙箱 → todo）
│   │   │   └── /:issue_id/review           POST → 🔒 审核中决策（通过/退回）
│   │   │
│   │   └── /:project_id/groups           项目群组
│   │       ├── /                          GET  → 群组列表
│   │       └── /                          POST → 🔐 创建群组
│   │
│   ├── /:org_id/direct-messages          双人聊（🔒）
│   │   └── /                              POST → 🔒 发起双人聊（幂等，人+人或人+Agent）
│   │
│   ├── /:org_id/conversations            会话列表（🔒）
│   │   └── /                              GET  → 混合群聊+双人聊列表（按最后消息时间倒序）
│   │
│   ├── /:org_id/agents                   Agent 管理（🔒）
│   │   ├── /                              GET  → Agent 列表
│   │   ├── /                              POST → 🔐 创建 Agent
│   │   ├── /:agent_id                     GET  → Agent 详情
│   │   ├── /:agent_id                     PUT  → 🔐 更新 Agent
│   │   ├── /:agent_id                     DELETE → 🔐 删除 Agent
│   │   ├── /:agent_id/enable              PUT  → 🔐 启用/禁用
│   │   ├── /:agent_id/skills              GET  → 已绑定 Skills
│   │   ├── /:agent_id/skills              PUT  → 🔐 更新 Skills 绑定
│   │   └── /:agent_id/logs                GET  → 执行日志（来源: agent_logs 表）
│   │   └── /:agent_id/token-usage         GET  → Token 用量详情（按日期聚合，含成本估算）
│   │
│   ├── /:org_id/token-summary             GET  → 组织级 Token 用量汇总（按 Agent + 按日期）
│   │
│   └── /:org_id/skills                   Skills 管理（🔒）
│       ├── /                              GET  → Skills 列表
│       ├── /                              POST → 🔐 创建 Skill
│       ├── /:skill_id                     GET  → Skill 详情
│       ├── /:skill_id                     PUT  → 🔐 更新 Skill
│       ├── /:skill_id                     DELETE → 🔐 删除 Skill
│       ├── /:skill_id/enable              PUT  → 🔐 启用/禁用
│       └── /import/zip                    POST → 🔐 ZIP 导入
│
├── /groups/:group_id                      群组/会话模块（🔒）
│   ├── /                                  GET  → 群组/会话信息（兼容 group 和 direct 类型）
│   ├── /members                           GET  → 成员列表
│   ├── /members                           POST → 🔐 添加成员（direct 类型返回 400）
│   ├── /members/:id                       DELETE → 🔐 移除成员（direct 类型返回 400）
│   ├── /messages                          GET  → 历史消息（分页）
│   ├── /messages                          POST → 发送消息
│   └── /read-state                        PUT  → 🔒 标记已读（更新 last_read_message_id）
│
├── /user                                  当前用户模块（🔒）
│   └── /settings                          GET/PUT → 个人偏好设置（user_settings 表）
│
├── /admin                                 系统管理（🔒 🔐 super_admin only）
│   ├── /settings                           GET/PUT → 全局系统配置（按 section 读写）
│   ├── /settings/{section}                 GET/PUT → section = agent/auth/smtp/sandbox/queue/upgrade
│   ├── /runtimes                           GET/POST → 运行时列表 / 注册新运行时
│   │   └── /:runtime_id                    GET/PUT/DELETE → 运行时详情/更新/删除（内置不可删）
│   │       └── /skills                     GET/PUT → 该运行时的默认 Skills 绑定
│
├── /notifications                         通知模块（🔒）
│   ├── /                                  GET  → 通知列表（分页）
│   ├── /unread-count                      GET  → 未读数
│   ├── /:id/read                          PUT  → 标记已读
│   └── /read-all                          PUT  → 全部已读
│
├── /audit-logs                            审计日志（🔒 🔐）
│   └── /                                  GET  → 按 org + 时间筛选
│
├── /webhook                                 Webhook 接收
│   └── /github                             POST → GitHub Webhook（PR merge → Issue→done）
│
└── /ws                                    WebSocket（认证，不绑定频道）
    └── ?token=xxx                         → 认证后通过 subscribe 消息动态订阅 group/issue/project/user 频道

/invite/:token [公开]                      GET  → 查看邀请详情 → 注册/登录后接受
```

> **图例**：无标注 = 公开端点 | `🔒` = 需 JWT 认证 | `🔐` = 需 JWT + Casbin RBAC | `[远期]` = [Phase 2](11-backlog.md) 实施

### 2.5.1 双人聊 API 详细说明

**POST /api/orgs/:org_id/direct-messages** — 发起双人聊（幂等）

请求体（二选一）：

```json
// 人+人
{ "target_user_id": 42 }

// 人+Agent
{ "target_agent_id": 5 }
```

响应体（已存在返回 200，新创建返回 201）：

```json
{
  "id": 99,
  "type": "direct",
  "org_id": 1,
  "name": null,
  "created_by": 7,
  "created_at": "2026-05-15T10:00:00Z",
  "members": [
    {"user_id": 7, "name": "张三", "avatar_url": "..."},
    {"user_id": 42, "name": "李四", "avatar_url": "..."}
  ]
}
```

> **设计说明**：direct 类型 `name` 为 `null`（双人聊不需要群名）。前端显示名由 API 自动计算后通过 `display_name` 字段返回，或者由前端从 `members` 中取对方成员信息派生。

校验规则：`target_user_id` 不能是自己；目标必须属于同一组织；调用 `GetOrCreateDirect` / `GetOrCreateDirectWithAgent` 幂等返回。

去重查询（应用层）：

```sql
-- 人+人去重
SELECT g.id FROM groups g
JOIN group_members gm1 ON g.id = gm1.group_id AND gm1.user_id = :userID AND gm1.agent_id IS NULL
JOIN group_members gm2 ON g.id = gm2.group_id AND gm2.user_id = :targetUserID AND gm2.agent_id IS NULL
WHERE g.type = 'direct' AND g.org_id = :orgID LIMIT 1;

-- 人+Agent 去重
SELECT g.id FROM groups g
JOIN group_members gm1 ON g.id = gm1.group_id AND gm1.user_id = :userID AND gm1.agent_id IS NULL
JOIN group_members gm2 ON g.id = gm2.group_id AND gm2.user_id IS NULL AND gm2.agent_id = :agentID
WHERE g.type = 'direct' AND g.org_id = :orgID LIMIT 1;
```

**GET /api/orgs/:org_id/conversations** — 会话列表

混合返回群聊和双人聊，按最后消息时间倒序排列。响应体：

```json
[
  {
    "id": 99, "type": "direct", "name": "李四", "avatar_url": "...",
    "last_message": { "content": "好的", "sender_name": "李四", "created_at": "..." },
    "unread_count": 2
  },
  {
    "id": 5, "type": "group", "name": "登录功能讨论组", "project_id": 3,
    "last_message": { "content": "前端预计 4h", "sender_name": "前端Agent", "created_at": "..." },
    "unread_count": 0
  }
]
```

direct 类型的 `name` 为 `null`（双人聊不需要群名）。显示名通过 `display_name` 字段返回，值为对方昵称（人+人）或 Agent 名称（人+Agent）。`unread_count` 通过 `group_read_state.last_read_message_id` 与 `messages` 表计算。

```json
[
  {
    "id": 99, "type": "direct", "name": null, "display_name": "李四", "avatar_url": "...",
    "last_message": { "content": "好的", "sender_name": "李四", "created_at": "..." },
    "unread_count": 2
  },
  {
    "id": 5, "type": "group", "name": "登录功能讨论组", "display_name": "登录功能讨论组", "project_id": 3,
    "last_message": { "content": "前端预计 4h", "sender_name": "前端Agent", "created_at": "..." },
    "unread_count": 0
  }
]
```

**PUT /api/groups/:group_id/read-state** — 标记已读

请求体：`{ "last_read_message_id": 12345 }`

更新 `group_read_state` 表，用于计算未读消息数。

### 2.5.2 Dashboard 聚合 API

`GET /api/orgs/:org_id/dashboard` 返回仪表盘所需的聚合数据：

```go
// internal/handler/dashboard_handler.go
type DashboardResponse struct {
    IssueDistribution IssueDistribution     `json:"issue_distribution"`
    AgentActivity     AgentActivity         `json:"agent_activity"`
    ProjectOverview   ProjectOverview       `json:"project_overview"`
    RecentActivity    []RecentActivityItem  `json:"recent_activity"` // 最近 7 天事件
}

// Issue 分布（按状态计数）
type IssueDistribution struct {
    Backlog    int64 `json:"backlog"`
    Todo       int64 `json:"todo"`
    InProgress int64 `json:"in_progress"`
    InReview   int64 `json:"in_review"`
    Done       int64 `json:"done"`
    Total      int64 `json:"total"`
    DoneRate   float64 `json:"done_rate"`  // done / total * 100
}

// Agent 活跃度（近 7 天）
type AgentActivity struct {
    ActiveAgents   int64                 `json:"active_agents"`   // 近 7 天有执行的 Agent 数
    TotalAgents    int64                 `json:"total_agents"`
    ExecutionsByDay []ExecutionCount     `json:"executions_by_day"`
}

type ExecutionCount struct {
    Date  string `json:"date"`  // "2026-05-14"
    Count int64  `json:"count"`
}

// 项目概览
type ProjectOverview struct {
    TotalProjects   int64 `json:"total_projects"`
    ActiveProjects  int64 `json:"active_projects"`  // 近 7 天有 Issue 活动
    TotalIssues     int64 `json:"total_issues"`
}

type RecentActivityItem struct {
    Type     string `json:"type"`      // issue_created / agent_completed / member_joined
    Message  string `json:"message"`
    Timestamp string `json:"timestamp"`
}
```

**GORM 查询实现**：

```go
func (r *DashboardRepo) GetIssueDistribution(ctx context.Context, orgID uint) (*IssueDistribution, error) {
    var dist IssueDistribution
    r.db.WithContext(ctx).
        Table("issues").
        Joins("JOIN projects ON projects.id = issues.project_id").
        Where("projects.org_id = ?", orgID).
        Select(`
            SUM(CASE WHEN issues.status='backlog' THEN 1 ELSE 0 END) as backlog,
            SUM(CASE WHEN issues.status='todo' THEN 1 ELSE 0 END) as todo,
            SUM(CASE WHEN issues.status='in_progress' THEN 1 ELSE 0 END) as in_progress,
            SUM(CASE WHEN issues.status='in_review' THEN 1 ELSE 0 END) as in_review,
            SUM(CASE WHEN issues.status='done' THEN 1 ELSE 0 END) as done,
            COUNT(*) as total
        `).Scan(&dist)

    if dist.Total > 0 {
        dist.DoneRate = float64(dist.Done) / float64(dist.Total) * 100
    }
    return &dist, nil
}

func (r *DashboardRepo) GetAgentActivity(ctx context.Context, orgID uint) (*AgentActivity, error) {
    since := time.Now().AddDate(0, 0, -7)
    var act AgentActivity
    r.db.WithContext(ctx).
        Table("agent_logs").
        Joins("JOIN agents ON agents.id = agent_logs.agent_id").
        Where("agents.org_id = ? AND agent_logs.created_at >= ?", orgID, since).
        Select("COUNT(DISTINCT agent_logs.agent_id) as active_agents").
        Scan(&act)

    r.db.WithContext(ctx).
        Raw(`SELECT DATE(created_at) as date, COUNT(*) as count
             FROM agent_logs
             WHERE org_id = ? AND created_at >= ?
             GROUP BY DATE(created_at)
             ORDER BY date`, orgID, since).
        Scan(&act.ExecutionsByDay)

    r.db.Model(&Agent{}).Where("org_id = ?", orgID).Count(&act.TotalAgents)
    return &act, nil
}
```

### 2.6 通知生成与分发

服务端内部通过 `NotificationService` 自动生成通知，非由客户端 API 创建：

```go
// internal/service/notification_service.go
// 所有通知统一通过 ChannelManager 分发，不直接调用 ws/smtp
type NotificationService struct {
    repo   *repository.NotificationRepo
    chMgr  *notification.ChannelManager  // 统一通知渠道管理器
}

// 典型触发场景：Issue 状态变更时自动生成通知
func (s *NotificationService) NotifyIssueStatusChanged(
    ctx context.Context,
    issue *model.Issue,
    assigneeID uint,
) {
    // 1. 写入 notifications 表
    notif := &model.Notification{
        OrgID:        issue.Project.OrgID,
        UserID:       assigneeID,
        Type:         "issue_status_changed",
        Title:        fmt.Sprintf("Issue #%d 状态变更为 %s", issue.ID, issue.Status),
        ResourceType: "issue",
        ResourceID:   issue.ID,
    }
    s.repo.Create(ctx, notif)

    // 2. 统一分发：ChannelManager 根据用户偏好自动路由到 WS/Email/浏览器通知
    s.chMgr.Notify(ctx, &notification.NotifyPayload{
        UserID:  assigneeID,
        Title:   notif.Title,
        Body:    notif.Title,
        Event:   notification.EventIssueStatusChange,
        Data:    map[string]interface{}{"issue_id": issue.ID, "status": issue.Status},
        GroupID: issue.SourceGroupID, // 群聊系统消息用
    })
}
```

> **设计原则**：`NotificationService` 不再直接调用 `s.ws.SendToUser()` 或 `s.smtp.SendXxx()`，所有推送统一走 `ChannelManager.Notify()`。ChannelManager 内部根据用户偏好自动决定通过哪些渠道推送（WebSocket / Email / 浏览器通知 / 群聊系统消息）。新增渠道只需注册 Channel，无需修改业务代码。

触发场景覆盖：Issue 分配/状态变更、Agent 执行完成/失败、群聊 @提及、组织邀请。

#### NotifyAgentComplete — Agent 执行完成通知

```go
// internal/service/notification_service.go
func (s *NotificationService) NotifyAgentComplete(
    ctx context.Context,
    issue *model.Issue,
    agent *model.Agent,
    success bool,
    detail string,
) {
    // 确定通知的负责人（Issue assignee 或项目 owner）
    assigneeID := s.getIssueAssigneeID(ctx, issue.ID)

    title := fmt.Sprintf("Agent '%s' 执行完成", agent.Name)
    if !success {
        title = fmt.Sprintf("Agent '%s' 执行失败", agent.Name)
    }

    notif := &model.Notification{
        OrgID:        issue.Project.OrgID,
        UserID:       assigneeID,
        Type:         "agent_completed",
        Title:        title,
        Body:         detail,
        ResourceType: "issue",
        ResourceID:   issue.ID,
    }
    s.repo.Create(ctx, notif)

    // 统一分发
    s.chMgr.Notify(ctx, &notification.NotifyPayload{
        UserID:  assigneeID,
        Title:   title,
        Body:    detail,
        Event:   notification.EventAgentCompleted,
        Data:    map[string]interface{}{"issue_id": issue.ID, "agent_id": agent.ID, "success": success},
        GroupID: issue.SourceGroupID,
    })
}
```

#### NotifyMention — 群聊 @提及通知

```go
func (s *NotificationService) NotifyMention(
    ctx context.Context,
    orgID uint,
    mentionedUserID uint,
    byUserName string,
    message string,
    groupID uint,
) {
    notif := &model.Notification{
        OrgID:        orgID,
        UserID:       mentionedUserID,
        Type:         "mention",
        Title:        fmt.Sprintf("%s 在群聊中提到了你", byUserName),
        Body:         message,
        ResourceType: "group",
        ResourceID:   groupID,
    }
    s.repo.Create(ctx, notif)
    // 统一分发
    s.chMgr.Notify(ctx, &notification.NotifyPayload{
        UserID:  mentionedUserID,
        Title:   fmt.Sprintf("%s 提到了你", byUserName),
        Body:    message,
        Event:   notification.EventMention,
        Data:    map[string]interface{}{"group_id": groupID, "by_user": byUserName},
        GroupID: groupID,
    })
}
```

#### NotifyOfflineMembers — 无 Agent 会话离线成员通知

当会话**无 Agent 成员**时（人+人 双人聊、无 Agent 群聊），对**未连接**该会话 WebSocket 的成员推送通知。有 Agent 的会话由 Eino 自动响应，不需要此通知。

```go
func (s *NotificationService) NotifyOfflineMembers(
    ctx context.Context,
    orgID uint,
    senderID uint,
    messageText string,
    groupID uint,
) {
    group := s.groupRepo.FindByID(groupID)
    // 获取会话中所有人类成员（排除 sender）
    members := s.groupRepo.FindHumanMembers(groupID)
    for _, member := range members {
        if member.UserID == senderID {
            continue
        }
        // 已在线且正在该会话页面 → 不推送
        if s.chMgr.IsUserConnectedToGroup(member.UserID, groupID) {
            continue
        }

        senderName := s.userRepo.FindByID(senderID).Name
        title := fmt.Sprintf("%s 发来一条消息", senderName)
        if group.Type == "group" {
            title = fmt.Sprintf("[%s] %s", group.Name, senderName)
        }

        notif := &model.Notification{
            OrgID:        orgID,
            UserID:       member.UserID,
            Type:         "dm_message",
            Title:        title,
            Body:         truncate(messageText, 100),
            ResourceType: "group",
            ResourceID:   groupID,
        }
        s.repo.Create(ctx, notif)
        // 统一分发
        s.chMgr.Notify(ctx, &notification.NotifyPayload{
            UserID:  member.UserID,
            Title:   title,
            Body:    truncate(messageText, 100),
            Event:   notification.EventNewDM,
            Data:    map[string]interface{}{"group_id": groupID, "sender_id": senderID},
        })
    }
}
```

#### NotifyInvite — 组织邀请通知

```go
func (s *NotificationService) NotifyInvite(
    ctx context.Context,
    invitation *model.Invitation,
    inviterName string,
) {
    if invitation.InviteType == "link" {
        return // 分享链接无需推送通知，用户访问链接时自行接受
    }

    // 邮箱邀请：发送邮件
    org, _ := s.orgRepo.FindByID(ctx, invitation.OrgID)
    inviteLink := fmt.Sprintf("%s/invite/%s", s.inviteBaseURL, invitation.Token)

    // 统一分发（ChannelManager 自动处理邮件 + WS 推送）
    // 同时写入 notifications 表（如果被邀请者已是平台用户）
    if existingUser := s.userRepo.FindByEmail(ctx, invitation.Email); existingUser != nil {
        notif := &model.Notification{
            OrgID:        invitation.OrgID,
            UserID:       existingUser.ID,
            Type:         "invite",
            Title:        fmt.Sprintf("%s 邀请你加入组织 %s", inviterName, org.Name),
            ResourceType: "organization",
            ResourceID:   invitation.OrgID,
        }
        s.repo.Create(ctx, notif)
        // 邮件由 ChannelManager 内部的 EmailChannel 统一处理
        s.chMgr.Notify(ctx, &notification.NotifyPayload{
            UserID: existingUser.ID,
            Title:  notif.Title,
            Body:   inviteLink,
            Event:  notification.Event("invite"),
            Data:   map[string]interface{}{"invite_link": inviteLink, "org_name": org.Name},
        })
    } else {
        // 非平台用户：仅发邮件（ChannelManager 的 EmailChannel 直接发送）
        s.chMgr.NotifyEmailOnly(ctx, invitation.Email, &notification.NotifyPayload{
            Title:  fmt.Sprintf("%s 邀请你加入组织 %s", inviterName, org.Name),
            Body:   inviteLink,
            Event:  notification.Event("invite"),
        })
    }
}
```

#### 完整通知触发场景矩阵

| 触发场景 | type | WS 推送 | 浏览器通知 | 群聊系统消息 | 邮件通知 | 触发位置 |
|---------|------|---------|------------|------------|---------|---------|
| Issue 分配 | `issue_assigned` | ✅ | ✅ | ❌ | ✅ (受 notify_issue_assigned 控制) | `IssueService.Assign` |
| Issue 状态变更 | `issue_status_changed` | ✅ | ❌ | ✅ | ✅ (受 notify_issue_assigned 控制) | `IssueService.TransitionStatus` |
| Agent 执行完成 | `agent_completed` | ✅ | ✅ | ✅ | ✅ (受 notify_agent_completed 控制) | `Worker.handleCompletion` |
| Agent 执行失败 | `agent_completed` | ✅ | ✅ | ✅ | ✅ (受 notify_agent_completed 控制) | `Worker.handleCompletion` |
| 群聊 @提及 | `mention` | ✅ | ✅ | ❌ | ✅ (受 notify_mention 控制) | `MessageService.Send` (解析 @) |
| 无 Agent 会话新消息（人+人 双人聊 / 无 Agent 群聊） | `dm_message` | ✅ | ✅ | ❌ | ✅ (受 notify_dm 控制) | `Hub.OnMessage` `notifyOfflineMembers` 分支（仅未连接该会话 WS 的成员推送）|
| 组织邀请 (邮箱) | `invite` | ✅ | ❌ | ❌ | ✅ (直接发送，不受偏好控制) | `InviteService.Create` |
| PR 已合并 (Issue done) | `issue_status_changed` | ✅ | ❌ | ✅ | ✅ | `WebhookHandler.HandleGitHub` |
| PR 被拒绝 | `issue_status_changed` | ✅ | ❌ | ✅ | ✅ | `WebhookHandler.HandleGitHub` |
| 重试次数耗尽 | `system` | ✅ | ❌ | ✅ | ✅ | `IssueScheduler.Run` |

> **推送说明**：客户端通过 WS 消息更新 UI（如通知 bell 红点），浏览器 Notification API 用于 OS 级弹窗提示。用户可通过 `user_settings` 表控制邮件和通知偏好。
>
> **群聊系统消息**：Issue 的 `source_group_id` 记录了来源群聊。每次 Issue 状态变更时，`TransitionStatus` 向该群聊发送 `type: system` 消息（写入 `messages` 表），群聊成员在聊天窗口实时可见状态变更通知。实现：
>
> ```go
> // TransitionStatus 中，每次状态变更后
> if issue.SourceGroupID != 0 {
>     s.messageService.SendSystemMessage(ctx, issue.SourceGroupID, fmt.Sprintf(
>         "Issue #%d %s → %s", issue.ID, oldStatus, newStatus))
> }
> ```

---

---

## 三、核心业务流程

### 3.1 需求讨论 → backlog 需求 → todo 任务列表

```mermaid
sequenceDiagram
    participant H as 自然人
    participant G as 群聊(WebSocket)
    participant ORC as 群聊Orchestrator(Eino)
    participant CEO as CEO Agent
    participant CTO as CTO Agent
    participant DEV as 前端 Agent
    participant SYS as Issue 服务

    H->>G: 发布需求："做一个用户登录页"
    ORC->>CEO: 调度 CEO Agent 分析需求
    CEO->>G: [分析] 这是产品需求，拆解为：<br/>1.登录UI 2.认证API 3.密码加密
    ORC->>CTO: 调度 CTO Agent 技术评审
    CTO->>G: [技术评审] 建议用 JWT + bcrypt
    ORC->>DEV: 调度 DEV Agent 评估工时
    DEV->>G: [评估] 前端预计 4h

    H->>G: /backlog 生成需求
    ORC->>ORC: 收集讨论上下文 → 结构化需求
    ORC->>SYS: 调用 Issue 服务创建（状态=backlog）

    SYS->>SYS: 创建 Issue #1: 用户登录页需求 (backlog, P1)
    SYS->>G: 系统消息: 已生成需求 Issue #1（backlog）
    H->>G: /todo 分析任务列表
    ORC->>ORC: 基于 Issue #1 + 当前讨论分析任务
    ORC->>SYS: 创建 parent_id=#1 的 todo 子 Issue
    SYS->>SYS: 创建 Issue #2/#3/#4 (todo, parent_id=#1)
    SYS->>G: 系统消息: Issue #1 下已生成 3 个任务
```

**Tab 视图操作**（backlog 是需求，todo 是同一需求下的任务列表）：

```
┌─ 项目 Issues ──────────────────────────────────────────────────────────────────┐
│  [ 需求 3 ]   任务 5    in_progress 2    in_review 1    done 12               │
├────────────────────────────────────────────────────────────────────────────────┤
│  ← 当前选中 backlog Tab，下面展示需求 Issue 与其任务列表摘要                         │
│                                                                                │
│  Issue #1  用户登录页需求                              P1  张三      12:00     │
│  │ 描述: 支持账号密码登录、错误提示、登录态保持...       [分析任务] [编辑]       │
│  │ ── 时间线 ─────────────────────────────────────────────────────────────── │
│  │ 12:00  system   /backlog 创建此 Issue                                      │
│  │ 12:05  system   /todo 生成 3 个子任务                                       │
│  │ ── 任务列表 ───────────────────────────────────────────────────────────── │
│  │ #2 登录表单 UI（todo） · #3 JWT 登录 API（todo） · #4 登录态保持（todo）        │
│  ├────────────────────────────────────────────────────────────────────────────┤
│  Issue #5  权限管理需求                                P2  李四      11:30     │
│  │ 描述: 管理员邀请成员并分配角色...                    [分析任务] [编辑]       │
│  └────────────────────────────────────────────────────────────────────────────┘
└────────────────────────────────────────────────────────────────────────────────┘
```

点击 `todo 5` 标签后，列表切换为：

```
│  [ 需求 3 ]  [ 任务 5 ]  in_progress 2    in_review 1    done 12             │
├────────────────────────────────────────────────────────────────────────────────┤
│  ← 当前选中 todo Tab，下面展示所有待执行任务 Issue，可按 parent backlog 分组          │
│                                                                                │
│  需求 #1 用户登录页需求                                                           │
│  │ Issue #2  登录表单 UI                           P1  前端Agent  11:50       │
│  │ 状态: 排队中（第 2 位），预估等待 3 分钟                                       │
│  ├────────────────────────────────────────────────────────────────────────────┤
│  │ Issue #3  JWT 登录 API                          P0  后端Agent  11:40       │
│  │ 状态: 排队中（第 1 位）                                                       │
│  └────────────────────────────────────────────────────────────────────────────┘
```

> 每个 Tab 的数字表示该状态的 Issue 数量，实时更新。当前选中的 Tab 高亮显示。

**Tab 设计要点**：

| 特性 | 说明 |
|------|------|
| 顶部 Tab 栏 | 按状态分页，Tab 上显示各状态的 Issue 数量 |
| 当前 Tab | 默认进入 `backlog` Tab，展开第一个 Issue 的时间线 |
| Issue 行 | 可展开/折叠，展开后显示描述摘要 + 最近 3 条时间线 |
| 任务列表 | `todo` 是从 backlog 分析出的子 Issue，通过 `parent_id` 归属到同一个 backlog |
| 批量操作 | 勾选多行 → 批量分析任务 / 批量编辑优先级 |
| 状态流转 | todo Tab 中 Issue 会自动消失（调度器转为 in_progress 后移到对应 Tab） |

**Issue 编辑能力**（backlog / todo 阶段均可）：

| 操作 | 说明 |
|------|------|
| 编辑标题/描述 | 直接修改，写入 `issue_timeline`（event_type=system_note） |
| Eino 优化描述 | 调用 Eino 阅读描述 → 识别模糊点 → 补充技术细节、验收条件、边界情况 → 输出优化版 |
| 调整优先级 | P0-P4 下拉切换 |
| 修改负责人 | 重新分配 Agent 或自然人 |
| 删除 Issue | 仅 backlog 状态可删除 |
| 批量分析任务 | Shift/Ctrl 多选 backlog Issue（来自多次 /backlog 调用）→ 批量生成各自的 todo 子 Issue |
| @关联 Issue | 描述中通过 `@Issue #N` 引用其他 Issue，anserAgent 自动读取被引用 Issue 的内容注入到 anserAgent 执行提示词 |

### 3.1.1 @Agent 任务布置（Agent 间协作）

群内 Agent 可根据其他 Agent 的角色定义（`system_prompt` + 绑定 Skills），通过 `@AgentName` 语法向指定 Agent 布置任务。这是 Agent 间自主协作的核心机制：

```mermaid
sequenceDiagram
    participant ORC as 群聊Orchestrator(Eino)
    participant CTO as CTO Agent
    participant FE as 前端 Agent
    participant BE as 后端 Agent
    participant SYS as Issue 服务

    Note over ORC: 自然人 /new 开启新会话后
    ORC->>CTO: 调度 CTO 讨论技术方案
    CTO->>ORC: "方案确认：前端负责UI，后端负责API"
    CTO->>ORC: "@前端Agent 请实现登录页表单组件，使用React Hook Form"
    ORC->>ORC: 解析 @前端Agent → 匹配群成员
    ORC->>FE: 注入任务上下文 + 角色定义 + Skills
    FE->>ORC: "收到，我会使用 React Hook Form + Zod 实现表单"
    CTO->>ORC: "@后端Agent 请实现 /api/auth/login 接口，JWT 签发"
    ORC->>BE: 注入任务上下文 + 角色定义 + Skills
    BE->>ORC: "收到，JWT + bcrypt 方案，预计需要实现 3 个接口"
```

**调度规则**：

| 规则 | 说明 |
|------|------|
| 角色感知 | `MentionResolver` 查询被 @ Agent 的角色定义，注入 System Prompt |
| Skill 继承 | 被 @ Agent 自动加载其绑定的 Skills（含 Runtime 默认 + Agent 独立绑定） |
| 上下文隔离 | 仅注入当前 `session_id` 内的消息历史，避免跨会话干扰 |
| 权限校验 | Casbin 校验发起 Agent 是否有权 @ 目标 Agent（同群成员即可） |
| 防循环 | 同一轮讨论中同一 Agent 最多被 @ 3 次，防止无限调度 |

### 3.1.2 `/new` 会话上下文切换

自然人发送 `/new` 开启新会话上下文。群聊消息历史仍在，但 Agent 的上下文窗口重置：

```mermaid
sequenceDiagram
    participant H as 自然人
    participant G as 群聊
    participant ORC as Orchestrator
    participant AG as Agent们

    H->>G: 讨论需求A："做登录页"
    ORC->>AG: 调度讨论（上下文窗口: session-1）
    AG->>G: 各 Agent 回复
    H->>G: /new 下一阶段：部署上线
    G->>G: 生成新 session_id，广播 new_session
    Note over ORC: 上下文窗口切换到 session-2
    H->>G: "准备部署流程"
    ORC->>AG: 调度讨论（上下文窗口: session-2，不含 session-1 消息）
```

**设计要点**：

| 要点 | 说明 |
|------|------|
| 历史可见 | `/new` 不删除历史消息，客户端仍可向上滚动查看 |
| 上下文隔离 | Agent 的 LLM 调用仅注入当前 session 的消息，避免 Token 浪费和主题混淆 |
| 会话标题 | `/new` 后跟的文本（如 `/new 部署上线`）作为新会话标题，显示在消息列表的分割线处 |
| Agent 感知 | Agent 收到新会话的首条消息时，其 System Prompt 追加 "这是一个新讨论主题的开端" |
| `/backlog` 作用域 | `/backlog` 指令仅收集当前 session 内的讨论上下文 |

### 3.2 Agent 自动执行（todo→in_progress→in_review→done）

```mermaid
sequenceDiagram
    participant H as 自然人(Issue Tab)
    participant ISS as Issue 服务
    participant Q as Asynq Queue
    participant W as Worker
    participant D as Docker 沙箱
    participant GH as GitHub
    participant WS as WebSocket
    participant G as 群聊

    Note over ISS,Q: ① todo 任务自动调度
    ISS->>ISS: 读取 parent backlog 下的 todo 子 Issue
    ISS->>Q: 加入排队（按优先级+创建时间）
    ISS->>WS: 推送 "Issue #2 已排队等待执行"
    ISS->>G: 系统消息: "Issue #2 已进入任务队列"

    Note over ISS,Q: ② 自动调度 todo → in_progress
    Q->>Q: 调度器检查队列 + 资源可用
    ISS->>ISS: Issue 状态 → in_progress
    ISS->>ISS: 记录状态变更 + 注入补充指令
    ISS->>ISS: 检查 assignee 类型 = agent
    ISS->>Q: enqueue("agent:execute", {issue_id, agent_id, human_prompts})
    Q-->>ISS: task_id
    ISS->>WS: 推送 "Agent 开始执行 Issue #2"
    ISS->>G: 系统消息: "Issue #2 开始执行，Agent 已启动"

    Note over W,D: ③ Worker 执行
    W->>Q: dequeue → "agent:execute"
    W->>W: 加载 Agent 配置 + Skills + 补充指令
    W->>W: 解析描述中的 @Issue #N → 读取关联 Issue 内容
    W->>D: 创建沙箱容器
    W->>D: git clone → 注入配置 + 关联Issue上下文 → anserAgent run
    W->>D: 监控日志输出 → 实时推送 Issue 时间线
    W->>WS: 推送 "编码中..." / "运行测试..."

    Note over H,D: ③.5 执行控制（暂停/恢复/停止）
    H->>ISS: 点击 [暂停]
    ISS->>W: ContainerPause
    W->>D: 冻结容器（内存保留）
    ISS->>ISS: Issue → paused + 时间线
    ISS->>WS: 推送 "执行已暂停"
    ISS->>G: 系统消息: "Issue #2 执行已暂停"
    H->>ISS: 点击 [恢复]
    ISS->>W: ContainerUnpause
    W->>D: 解冻容器
    ISS->>ISS: Issue → in_progress + 时间线
    alt 点击 [停止]
        H->>ISS: 点击 [停止]
        ISS->>W: ContainerStop + Remove
        W->>D: 终止 + 销毁容器
        ISS->>ISS: Issue → todo + 时间线
        ISS->>WS: 推送 "执行已停止"
        ISS->>G: 系统消息: "Issue #2 执行已停止"
    end

    Note over W,GH: ④ anserAgent 自检查 → PR
    W->>D: anserAgent 返回执行结果
    W->>W: 检查结果（代码完整性/lint/test 通过）
    alt anserAgent 检查通过
        W->>D: git add → commit → push
        W->>GH: 创建 Pull Request
        W->>ISS: 更新 Issue status → in_review + 写入 pr_url
        W->>WS: 推送通知 "PR 已提交，等待审核"
        W->>G: 系统消息: "Issue #2 PR 已提交，等待审核"
    else anserAgent 检查失败
        W->>ISS: 写入失败原因到时间线
        W->>ISS: Issue 状态 → todo（保留沙箱）
        W->>WS: 推送 "执行失败: {原因}，已退回任务列表"
        W->>G: 系统消息: "Issue #2 执行失败: {原因}"
        Note over H,D: 可在任务详情追加执行指令 → 重新调度
        ISS->>ISS: 收到补充指令 → Issue → in_progress
        ISS->>Q: enqueue("agent:execute", {issue_id, human_prompts_optimized})
        W->>Q: dequeue → 检测已有沙箱 → 复用旧沙箱
        W->>D: 注入新的优化提示词 → anserAgent run（基于上次工作区继续）
    end

    Note over H,GH: ⑤ in_review 人工审核 → merge → done
    H->>GH: Review PR → Approve → Merge
    GH-->>ISS: Webhook: PR merged
    ISS->>ISS: Issue 状态 → done
    ISS->>ISS: 记录状态变更
    ISS->>WS: 推送 "Issue #2 已完成"
    ISS->>G: 系统消息: "Issue #2 已完成 ✅"
    W->>D: 容器自动销毁
```

**状态流转规则**：

| 状态 | 触发方式 | 下一状态 | 触发条件 |
|------|---------|---------|---------|
| `backlog` | `/backlog` 指令自动创建 | `todo` 子 Issue | `/todo` 分析需求后生成任务列表，子任务通过 `parent_id` 归属到 backlog |
| `todo` | `/todo` 任务列表生成后 | `in_progress` | 系统自动调度（按优先级+排队顺序） |
| `in_progress` | 系统自动 | `in_review` | anserAgent 检查通过 + PR 创建成功；此状态进入人工审核 |
| `in_progress` | 系统自动 | `paused` | 人工点击 [暂停] |
| `in_progress` | 系统自动 | `todo` | 点击 [停止]（终止 anserAgent 进程，worktree 保留） |
| `paused` | 人工暂停 | `in_progress` | 人工点击 [恢复]（SIGCONT 恢复 anserAgent 进程） |
| `paused` | 人工暂停 | `todo` | 点击 [停止]（终止 anserAgent 进程，worktree 保留） |
| `in_progress` | 系统自动 / 人工恢复 | `todo` | anserAgent 检查失败 → 保留 worktree → 等待重新调度 |
| `in_review` | anserAgent 检查通过后 | `done` | GitHub Webhook 通知 PR 已 merge → git worktree remove |
| `in_review` | 人工审核 | `todo` | PR 被拒绝或要求修改 → 保留 worktree → 复用上下文重新执行 |

**补充指令介入机制**：

Issue 的时间线面板允许自然人在 `todo` / `in_progress` / `paused` 阶段追加补充指令，干预 anserAgent 的下一步执行；`in_review` 阶段只处理 PR 审核动作。

```
┌─────────────────────────────────────────────────────────┐
│  Issue 时间线                                            │
│  ┌───────────────────────────────────────────────────┐  │
│  │ 12:00  system   从需求 #1 分析出任务 #2              │  │
│  │ 12:01  system   状态变更: todo → in_progress       │  │
│  │ 12:02  agent    开始编码: 正在读取 Issue 描述...    │  │
│  │ 12:05  agent    生成文件: src/login.tsx             │  │
│  │ 12:08  agent    运行测试: 2 passed, 1 failed       │  │
│  │ ─────────────────────────────────────────────      │  │
│  │ 12:09  张三     提示词: "login.tsx 的密码框需要     │  │
│  │                 autocomplete='new-password'"        │  │
│  │ 12:09  system   收到补充指令，重新执行 anserflow      │  │
│  │ 12:12  agent    修复完成: lint + test 全部通过      │  │
│  └───────────────────────────────────────────────────┘  │
│                                                         │
│  ┌─ 追加提示词 ──────────────────────────────────────┐  │
│  │ [                    ]  [发送并重新执行]            │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

**实现**：

- `POST /api/orgs/:org_id/projects/:project_id/issues/:issue_id/prompt` → 写入 `issue_timeline`（event_type=human_prompt）
- Issue 服务收到补充指令后调用 `prompt_optimizer.Enhance()`（Eino 优化改写）
- 优化后的指令写入时间线，触发 Issue 重新进入 in_progress
- Worker 检测到已有 worktree（`project_id → 容器 → /workspace/issue-{id}`），在 worktree 内继续执行
- anserflow 基于上次工作区状态 + 优化指令继续编码
- Issue done 后清理 worktree（`git worktree remove`），容器常驻不销毁
- anserflow 重新执行时保留之前的 `issue_timeline` 日志记录

**GitHub Webhook 处理器**：

```go
// internal/handler/webhook.go
func (h *WebhookHandler) HandleGitHub(c *gin.Context) {
    event := c.GetHeader("X-GitHub-Event")
    if event != "pull_request" { return }

    var payload github.PullRequestEvent
    c.ShouldBindJSON(&payload)

    if payload.Action == "closed" && payload.PullRequest.Merged {
        // ① PR 已合并 → Issue done
        issue := h.issueRepo.FindByPRURL(payload.PullRequest.HTMLURL)
        h.issueRepo.UpdateStatus(issue.ID, "done")
        h.timelineRepo.Append(issue.ID, "system", "status_change",
            "in_review", "done", "PR merged by @"+payload.Sender.Login)
        // 销毁沙箱容器 + 清空 container_id
        if issue.SandboxContainerID != "" {
            h.sandbox.Destroy(ctx, issue.SandboxContainerID)
            h.issueRepo.ClearContainerID(issue.ID)
        }
        h.ws.SendToProject(issue.ProjectID, "Issue #%d 已完成", issue.ID)
        h.notification.NotifyIssueStatusChanged(ctx, issue, issue.CreatedBy)
        return
    }

    if payload.Action == "closed" && !payload.PullRequest.Merged {
        // ② PR 被拒绝（closed but not merged）→ Issue 回到 todo
        issue := h.issueRepo.FindByPRURL(payload.PullRequest.HTMLURL)
        rejectReason := fmt.Sprintf("PR #%d rejected by @%s",
            payload.PullRequest.Number, payload.Sender.Login)

        h.issueRepo.UpdateStatus(issue.ID, "todo")
        h.timelineRepo.Append(issue.ID, "system", "status_change",
            "in_review", "todo", rejectReason)

        // 保留沙箱容器（不销毁），后续补充指令可复用上次工作区
        h.ws.SendToProject(issue.ProjectID,
            "Issue #%d PR 被拒绝，已退回 todo。可在时间线追加补充指令后重新执行。", issue.ID)
        h.notification.NotifyIssueStatusChanged(ctx, issue, issue.CreatedBy)
        return
    }
}
```

> **PR 被拒绝的处理策略**：
> - 沙箱容器**保留不销毁**，后续追加补充指令 → 重新入队时复用旧沙箱，基于上次工作区继续修改
> - 如果管理员判断 PR 方向完全错误需要重来，可通过 Issue 详情页 **[停止]** 按钮手动销毁沙箱后重新编辑
> - 状态流转：`in_review → todo`（而非 backlog），表示仍需处理但保留已有上下文

> GitHub Webhook 需在项目关联的 GitHub 仓库中配置 Payload URL: `https://<host>/api/webhook/github`，Events: `Pull requests`。

---
