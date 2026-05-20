# AnserFlow — 开发路线图

## L1 — 基础设施（1-2 周）

| 编号 | 任务 | 验收标准 |
|------|------|----------|
| T01 | Go 项目初始化（Gin + GORM） | `go run main.go server` 启动成功 |
| T02 | Next.js 项目初始化（shadcn/ui） | `npm run dev` 启动成功 |
| T03 | MySQL 表结构创建 + GORM AutoMigrate | 所有表自动创建 |
| T04 | 用户注册/登录（JWT + bcrypt） | Postman 测试通过 |
| T05 | 组织 CRUD + 邀请系统（分享链接 + 邮箱） | 创建组织、生成链接/发送邮件、接受邀请自动入组织、成员管理全流程通 |
| T06 | Redis 集成 + WebSocket 基础 | 单实例 WS 通信正常 |
| T07 | Cobra CLI 框架 + `anserflow init/server` | 命令行可初始化并启动 |
| T08 | Next.js SPA 静态导出 + Go embed 嵌入 | `go build` 产出单文件，浏览器访问正常 |

## L2 — 核心业务（3-4 周）

| 编号 | 任务 | 验收标准 |
|------|------|----------|
| T09 | Agent CRUD + System Prompt 编辑 | 创建/编辑 Agent，保存人设 |
| T10 | Skills CRUD + 手动/ZIP 导入 | 两种方式均可添加 Skill |
| T11 | Agent-Skill 绑定 + 启用控制 | 全局/单Agent 开关生效 |
| T12 | 项目管理 + GitHub 关联（HTTP Token / SSH Key） | 创建项目并绑定 GitHub 仓库，支持两种授权方式 |
| T13 | Issue CRUD + 状态流转 + 优先级 + 子 Issue | Tab 状态视图操作全流程通 |
| T14 | Issue 分配给 Agent/自然人 | 分配并正确记录 |
| T15 | Issue 执行日志记录 | agent_logs 写入 + 前端可查询 |

## L3 — 协作与执行（5-7 周）

| 编号 | 任务 | 验收标准 |
|------|------|----------|
| T16 | 群聊 WebSocket + 消息持久化 | 多人实时聊天正常 |
| T17 | WebSocket Redis Pub/Sub 分布式 | 多实例消息同步 |
| T18 | Eino 集成 + 群聊 Agent 讨论编排 | Agent 自动参与讨论 |
| T19 | 讨论→方案→自动创建 Issue（backlog） | /backlog 指令产出 Issue（状态=backlog），到 backlog Tab 手动确认后转为 todo |
| T20 | Asynq 任务队列集成 | 任务入队/消费正常 |
| T21 | Docker 沙箱执行引擎 | Agent 在容器中执行编码 |
| T22 | GitHub PR 自动提交 | 代码提交 + PR 创建流程通 |
| T22a | Agent 执行日志（agent_logs + 前端筛选） | Agent 讨论/执行过程记录到 agent_logs 表，前端可按时间/类型筛选 |

## L4 — 客户端与交付（8-10 周）

| 编号 | 任务 | 验收标准 |
|------|------|----------|
| T23 | 交叉编译 + CI 构建（win/mac/linux） | 三平台产出单文件 |
| T24 | 客户端 Web SPA（IM 聊天界面） | 浏览器访问，聊天/Issue/项目功能完整 |
| T25 | 通知系统（WebSocket Push + 浏览器通知 + 邮箱） | 状态变更实时通知 |
| T26 | 权限精细化 + 操作审计日志 | RBAC 权限生效 |
| T27 | 管理后台完整 UI | 所有页面可交互 |
| T28 | 性能优化 + 压力测试 | 100 并发 WS 连接稳定 |

## 测试策略

测试贯穿全部四个阶段，不在单独阶段集中编写：

| 层级 | 框架 | 目标 | 触发时机 |
|------|------|------|----------|
| **Go 单元测试** | `testing` + `testify` | Service / Model 层覆盖率 > 70% | 每次 `go test`（CI PR） |
| **Go 集成测试** | `testing` + `testcontainers-go` | MySQL/Redis 真实交互（Handler + DB 层） | PR + main push |
| **前端组件测试** | Vitest + Testing Library | 核心组件（AgentForm / IssueCard / TodoKanban） | PR |
| **E2E 测试** | Playwright | 关键流程：注册→登录→创建Agent→创建Issue→状态流转→邀请 | main push / 发布前 |
| **WebSocket 测试** | `gorilla/websocket` 客户端 + testify | 消息格式 / 心跳 / 重连 / 分布式 Pub/Sub | PR |
| **压力测试** | k6 / vegeta | WS 并发连接 > 100、API QPS > 500 | L4 阶段 |
| **合约测试** | Pact | 前后端 API 契约一致性 | [Phase 2](11-backlog.md) 独立立项 |

> **CI 闭环要求**：当前仓库需补齐 `ci.yml`，至少覆盖 Go `test + lint + build` 与 Next.js `type-check + lint + build`。

## 数据库迁移策略

GORM AutoMigrate 仅处理正向迁移（创建表/添加列）。需要回滚时采用：

| 场景 | 方案 |
|------|------|
| **本地开发** | `anserflow migrate --dry-run` 预览 SQL → 手动执行回滚 DDL |
| **生产发布** | 每次 `anserflow migrate` 前自动生成备份 SQL（`data/migrations/YYYYMMDDHHMMSS_before.sql`） |
| **紧急回滚** | 执行对应时间的备份 SQL 恢复表结构 |
| **当前收口** | 本轮统一采用 `AutoMigrate + 备份 SQL + 种子数据`；`golang-migrate/migrate` 放入 [Phase 2](11-backlog.md) 独立任务 |

```bash
# 迁移前自动备份
anserflow migrate --backup    # → data/migrations/20260514120000_before.sql

# 跳过种子数据（仅 DDL）
anserflow migrate --seed=false
```

### 数据库`restore` 子命令

除 `migrate --backup` 自动备份外，提供独立的 `restore` 子命令用于灾难恢复：

```bash
anserflow restore --file data/migrations/20260514120000_before.sql
  --config  config.yaml 路径
  --yes     跳过二次确认（默认 false）
  --dry-run 预览将执行的 SQL（默认 false）
```

## 种子数据

种子数据 SQL 统一放置在 `internal/seed/` 目录：

```
internal/seed/
├── 001_default_skills.sql     # 系统预置 Skill（anser-coder + 6个角色Skill）
├── 002_casbin_policies.sql    # Casbin RBAC 角色权限策略
├── 003_runtime_skills.sql     # 各运行时默认 Skill 绑定（opencode→anser-coder）
└── 004_example_agent.sql      # 可选：示例 Agent 配置
```

**预置 Skill 清单**：

| Skill 名称 | 用途 | anserAgent 环节 | 核心内容 |
|-----------|------|----------------|----------|
| `anser-coder` | 编码执行规范 | opencode/hermes 沙箱执行 | 代码风格、提交规范、PR 格式 |

> 预置 Skill 仅 `anser-coder`（沙箱执行规范）。调度编排能力由 anserAgent 五层记忆系统提供，不再依赖硬编码 eino-* Skills。


