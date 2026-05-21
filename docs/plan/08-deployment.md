# AnserFlow - 部署指南

## 架构总览

```
┌─────────────────────────────────────────────────────┐
│  服务器                                              │
│  ┌────────────┐   ┌──────────┐   ┌───────────────┐ │
│  │ anserflow   │   │  MySQL   │   │    Redis      │ │
│  │ (单二进制)  ├──►│  8.0     │   │  (AOF 持久化) │ │
│  │ :8080       │   │  :3306   │   │  :6379        │ │
│  └──────┬──────┘   └──────────┘   └───────────────┘ │
│         │                                            │
│         │ Docker API                                 │
│         ▼                                            │
│  ┌──────────────────────────────────────────────┐   │
│  │  Docker Engine                                │   │
│  └──────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

## 前置依赖

| 组件 | 版本要求 | 用途 |
|------|---------|------|
| **MySQL** | 8.0+ | 主数据库（组织/项目/Issue/Agent 等） |
| **Redis** | 7.0+ | Asynq 任务队列 + WebSocket Hub + 缓存 |
| **Docker** | 24.0+ | 沙箱容器管理 |
| **Git** | 2.x+ | 仓库操作 |

## 部署模式

### 模式 A：单机一体化（推荐起步）

HTTP 服务 + Asynq Worker 运行在同一进程，适合中小规模（<50 并发 Issue）：

```bash
# 1. 准备基础设施（MySQL + Redis）
# Redis 必须开启 AOF 持久化（保障 Asynq 任务队列不丢失）
# redis.conf 中配置：
#   appendonly yes
#   appendfsync everysec

# 2. 构建沙箱镜像
docker build -t anserflow/sandbox:latest -f docker/sandbox/Dockerfile .

# 3. 初始化
./anserflow init \
  --db "root:password@tcp(127.0.0.1:3306)/anserflow" \
  --redis "127.0.0.1:6379" \
  --data /var/lib/anserflow
# 自动生成 config.yaml + 数据库表结构 + 全局运行时模板目录

# 4. 编辑配置（重点项）
vim config.yaml
#   - database.password（或用环境变量 DB_PASSWORD）
#   - sandbox.runtime_data_dir → /var/lib/anserflow
#   - sandbox.allowed_domains（网络白名单）
#   - smtp（邮件通知，可选）

# 5. 启动（HTTP + Worker 一体）
./anserflow server --config ./config.yaml
# 等价于同时执行 gateway + worker，访问 http://your-server:8080/admin  → 后台管理
#                                      http://your-server:8080/client → 客户端 IM
```

### 模式 B：分布式部署

适合大规模（50+ 并发 Issue），Worker 可水平扩展：

```bash
# 节点 1: API 网关（可多实例 + 负载均衡）
anserflow gateway --config config.yaml

# 节点 2~N: Worker（仅处理沙箱任务）
anserflow worker --config config.yaml --mode sandbox

# 数据库/Redis: 独立部署或托管服务
```

## 部署检查清单

| 步骤 | 命令/操作 | 验证 |
|------|----------|------|
| MySQL 就绪 | 创建 `anserflow` 库 | `mysql -e "SHOW DATABASES"` |
| Redis AOF | `appendonly yes` | `redis-cli INFO Persistence` |
| Docker 运行 | Docker daemon 启动 | `docker info` |
| 沙箱镜像 | `docker build ...` | `docker images | grep anserflow` |
| 运行时数据目录 | `mkdir -p /var/lib/anserflow` | 目录存在 |
| 全局模板初始化 | 首次 `anserflow init` | `/var/lib/anserflow/runtimes/anseragent/` 存在 |
| 健康检查 | `curl /api/health` | 返回 `{"status":"ok"}` |

## 服务器关键目录结构

```
/var/lib/anserflow/
├── runtimes/                     # 全局运行时模板
│   └── anseragent/               # anserAgent 默认配置 + Skills
│       ├── config.yaml
│       └── skills/
├── projects/
│   └── {project_id}/
│       ├── runtime/              # 项目级运行时实例（从模板复制）
│       └── workspace/            # Named Volume 数据
```

## 注意事项

1. **Redis 必须开 AOF** — 否则服务器重启后 Asynq 任务队列丢失，`in_progress` Issue 无法恢复
2. **Docker 容器 AutoRemove=false** — 配合 `RecoverRunningIssues` 恢复机制
3. **端口 8080** — 唯一对外端口（API + WebSocket + 两个 SPA 全在同一个端口）
4. **单二进制部署** — `admin/dist` 和 `client/dist` 通过 Go embed 打入，服务器不需要 Node.js 运行时
5. **沙箱镜像仅含 anserflow 二进制** — `anserflow agent run`（Go 二进制）+ git + bash，`ENTRYPOINT ["/entrypoint.sh"]` 由 Go Docker SDK 调用


