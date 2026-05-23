> ???`docs/plan/01-admin.md` ? 5 ?
> ???[???](../../README.md) -> [AnserFlow - Admin Backend](../README.md) -> 完整配置文件 (config.yaml)
> ???[???](../README.md) ? [???](../02-四、构建与部署/README.md)
> ?????[??????](../README.md) ? [四、构建与部署](../02-四、构建与部署/README.md)

## 完整配置文件 (config.yaml)

> AnserFlow 运行时所有配置集中在 `config.yaml`，由 Viper 加载。生产环境敏感字段（数据库密码、API Key 等）可通过环境变量覆盖。
> 
> **配置分级原则**：
> - 🔴 **config.yaml only** — 基础设施，改后需重启服务
> - 🟡 **config.yaml + 后台覆盖** — 有默认值，后台可运行时修改（存 DB，重启后以 DB 为准）
> - 🟢 **纯后台管理** — 不走 config.yaml，存储在 DB 中

```yaml
# config.yaml — AnserFlow 完整配置

# ═══════════════════════════════════════════════════════════════
# 🔴 基础设施（config.yaml only | 存 DB: ❌ | 改后重启: ✅）
# ═══════════════════════════════════════════════════════════════
server:
  port: 8080
  mode: release                  # debug | release | test
  read_timeout: 30s
  write_timeout: 30s

database:
  driver: mysql
  host: 127.0.0.1
  port: 3306
  database: anserflow
  username: root
  password: ${DB_PASSWORD}
  charset: utf8mb4
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600s
  log_level: warn

redis:
  host: 127.0.0.1
  port: 6379
  password: ""
  db: 0
  pool_size: 50
  # ↓ 部署时需在 redis.conf 中配置 AOF 持久化（保障 Asynq 任务队列不丢失）
  # appendonly yes
  # appendfsync everysec

# ═══════════════════════════════════════════════════════════════
# 🟡 服务级（config.yaml 提供默认值，后台 /admin/settings 可覆盖 | 存 DB: ✅ | 改后重启: ❌ 即时生效）
# ═══════════════════════════════════════════════════════════════
jwt:
  secret: ${JWT_SECRET}          # 密钥不入库，仅 config.yaml
  expire_hours: 720              # 30 天，后台可覆盖
  issuer: anserflow

oauth2:
  github:
    client_id: ${GITHUB_CLIENT_ID}
    client_secret: ${GITHUB_CLIENT_SECRET}
    redirect_url: http://localhost:8080/api/auth/github/callback
    scopes: ["user:email"]

cors:
  allow_origins:
    - http://localhost:3000
    - http://localhost:3001

log:
  level: info
  format: json
  output: stdout

smtp:                            # 后台可覆盖
  host: smtp.example.com
  port: 587
  username: noreply@anserflow.io
  password: ${SMTP_PASSWORD}
  from: "AnserFlow <noreply@anserflow.io>"
  ssl: false

invite:                          # 默认值，管理员可在组织设置中覆盖
  link_base_url: http://localhost:8080
  default_expire_hours: 168      # 7 天
  max_uses_default: 0            # 0 = 不限

upgrade:
  channel: stable
  endpoint: https://github.com/anserflow/anserflow/releases/latest/download
  check_interval: 24h

asynq:                           # Worker 默认值，后台可调整全局默认，组织可覆盖
  concurrency: 10
  queues:
    critical: 6
    default: 3
    low: 1
  retry:
    max_retry: 3
    min_backoff: 5s
    max_backoff: 5m
  timeout: 1800s                 # 单任务最长 30 分钟

sandbox:                         # Docker 沙箱默认值
  image: ghcr.io/anserflow/sandbox:latest
  memory: 512                    # MB
  cpu: 2                         # cores
  disk: 1024                     # MB
  timeout: 1800s
  network: restricted
  allowed_domains:
    - github.com
    - api.github.com
    - api.openai.com
  runtime_data_dir: /var/lib/anserflow  # 运行时数据根目录（全局模板 + 项目实例）

# ═══════════════════════════════════════════════════════════════
# 🟢 纯后台管理（config.yaml 仅存默认值，运行时从 DB 读取 | 存 DB: ✅ | 改后重启: ❌ 即时生效）
# ═══════════════════════════════════════════════════════════════
agent:                           # 后台 /admin/settings#agent 配置
  provider: openai
  api_key: ${AGENT_LLM_API_KEY}
  model: gpt-4o
  temperature: 0.7
  max_tokens: 4096
  timeout: 120s
  discuss:
    max_turns: 5
    agent_timeout: 60s
  backlog:
    context_window: 50
    require_project: true
  optimizer:
    model: gpt-4o-mini
    temperature: 0.3
  rate_limit:
    capacity: 100
    refill_rate: 1.67
```

> **环境变量覆盖规则**：Viper 以 `AGENT_LLM_API_KEY` 覆盖 `agent.api_key`，`DB_PASSWORD` 覆盖 `database.password`。所有 `${VAR}` 占位符必须通过环境变量注入。

## ?????

- [配置归属速查](01-配置归属速查.md)
- [配置热更新机制](02-配置热更新机制.md)
- [后台管理页面结构](03-后台管理页面结构.md)
