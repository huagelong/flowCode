# Admin 系统设置

> 路由：`/admin/settings`
> API：`/api/admin/settings/*`
> 权限：仅 `super_admin`

---

## 1. 页面布局

```
┌──────────────────────────────────────────────────────────────────┐
│ 系统设置 (super_admin)                                           │
│                                                                  │
│ ┌──────────────────────────────────────────────────────────────┐ │
│ │ [🟢 Agent] [🟡 认证] [🟡 SMTP] [🟡 邀请] [🟡 沙箱] [🟡 队列] │ │
│ │ [🟡 升级]                                                    │ │
│ └──────────────────────────────────────────────────────────────┘ │
│                                                                  │
│ ┌─── Agent 配置 ────────────────────────────────────────────────┐│
│ │                                                              ││
│ │ 默认 LLM Provider                                            ││
│ │ ┌──────────────┐                                             ││
│ │ │ openai     ▾ │                                             ││
│ │ └──────────────┘                                             ││
│ │                                                              ││
│ │ 默认 Model                                                   ││
│ │ ┌──────────────┐                                             ││
│ │ │ gpt-4o       │                                             ││
│ │ └──────────────┘                                             ││
│ │                                                              ││
│ │ 讨论参数                                                     ││
│ │ 最大上下文消息数                                              ││
│ │ ┌──────────────┐                                             ││
│ │ │ 50           │                                             ││
│ │ └──────────────┘                                             ││
│ │                                                              ││
│ │ Backlog 参数                                                 ││
│ │ 单次最大生成 Issue 数                                        ││
│ │ ┌──────────────┐                                             ││
│ │ │ 5            │                                             ││
│ │ └──────────────┘                                             ││
│ │                                                              ││
│ │ Rate Limit (次/分钟)                                         ││
│ │ ┌──────────────┐                                             ││
│ │ │ 20           │                                             ││
│ │ └──────────────┘                                             ││
│ │                                                              ││
│ │                                              [保存配置]      ││
│ └──────────────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────────────┘
```

---

## 2. 设置分类 Tab

每个 Tab 对应后端 `/api/admin/settings/:section`。

### Tab 标记色

| 颜色 | 含义 |
|------|------|
| 🟢 绿点 | Agent 级别配置，影响 AI 行为 |
| 🟡 黄点 | 服务级别配置，影响系统运行 |

### 2.1 Agent 配置 (`section=agent`)

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| default_provider | Select | openai | 默认 LLM 提供商 |
| default_model | Input | gpt-4o | 默认模型 |
| discuss_max_context | Number | 50 | 讨论 context 最大消息数 |
| backlog_max_issues | Number | 5 | Agent 自动生成方案时单次最大 Issue 数 |
| optimizer_enabled | Switch | true | Prompt 优化器 |
| rate_limit | Number | 20 | 每分钟请求限制 |

### 2.2 认证配置 (`section=auth`)

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| jwt_expire_hours | Number | 72 | JWT 过期时间（小时） |
| github_client_id | Input | — | GitHub OAuth Client ID |
| github_client_secret | Password | — | GitHub OAuth Secret |
| cors_origins | Textarea | * | CORS 允许的源（每行一个） |

### 2.3 SMTP 配置 (`section=smtp`)

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| smtp_host | Input | — | SMTP 服务器地址 |
| smtp_port | Number | 587 | SMTP 端口 |
| smtp_username | Input | — | 用户名 |
| smtp_password | Password | — | 密码 |
| smtp_sender | Input | — | 发件人地址 |
| smtp_use_tls | Switch | true | 使用 TLS |

**测试按钮**: 发送测试邮件到当前用户邮箱，成功/失败 Toast 反馈。

### 2.4 邀请配置 (`section=invite`)

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| default_expiration_hours | Number | 168 | 默认有效期（小时，168=7天） |
| default_max_uses | Number | 0 | 默认最大使用次数（0=无限） |

### 2.5 沙箱配置 (`section=sandbox`)

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| default_cpu_limit | Input | "1.0" | CPU 限制（核心数） |
| default_memory_mb | Number | 1024 | 内存限制（MB） |
| default_disk_mb | Number | 5120 | 磁盘限制（MB） |
| default_timeout_minutes | Number | 30 | 执行超时（分钟） |
| network_whitelist | Textarea | — | 网络白名单（每行一个域名） |
| default_concurrency | Number | 3 | 默认并发数（可被组织级覆盖） |

### 2.6 队列配置 (`section=queue`)

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| asynq_concurrency | Number | 10 | Asynq 并发数 |
| max_retry | Number | 3 | 最大重试次数 |
| retry_delay_seconds | Number | 60 | 重试间隔（秒） |
| timeout_seconds | Number | 1800 | 任务超时（秒） |

### 2.7 升级配置 (`section=upgrade`)

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| channel | Select | stable | 升级通道 (stable/beta/nightly) |
| check_interval_hours | Number | 24 | 检查间隔（小时） |

---

## 3. 保存行为

1. 点击"保存配置"
2. 按钮 loading 态 (Spinner + "保存中...")
3. `PUT /api/admin/settings/:section` + 请求体
4. 成功 → Sonner toast "配置已保存，部分设置立即生效"
5. 失败 → Sonner toast 错误信息，表单保持脏值
6. 后端热更新：写 DB → 更新内存缓存 → Redis Pub/Sub 通知其他实例

---

## 4. 危险操作区域

页面底部：

```
┌─ ⚠️ 危险区域 ────────────────────────────────────────────────┐
│                                                              │
│ 重置所有配置为默认值                                         │
│ 将所有设置恢复为 config.yaml 中的默认值                       │
│                                              [重置配置]      │
│                                                              │
│ 清理沙箱容器                                                 │
│ 停止并删除所有非活跃的沙箱容器                                │
│                                              [清理容器]      │
└──────────────────────────────────────────────────────────────┘
```

- 危险区域有红色边框 `border-destructive/50`
- 操作前需二次确认 AlertDialog
