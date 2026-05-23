> 来源：`docs/plan/04-sandbox.md` 第 1241 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱 / Agent 基础设施](../README.md) -> [一、AI Agent 框架：Eino + 自研封装](README.md) -> Token 用量与成本追踪
> 相邻：[上一篇](15-@Agent-任务布置.md) · [下一篇](../02-二、Docker-沙箱方案/README.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### Token 用量与成本追踪

系统 Token 消耗来自两个阶段，需要分别追踪后汇总：

| 阶段 | 谁调 LLM | 追踪方式 | 占比（典型） |

|------|---------|---------|------------|

| **anserAgent 调度** | Go 后端进程 | `TokenTracker.Record(agentID, usage, "agent")` | ~10-20% |

| **anserAgent 执行** | Docker 沙箱内 | 解析 stdout JSON + session 文件 | ~80-90% |

#### TokenTracker — 统一记录（区分来源）

```go

// internal/agent/token_tracker.go

type TokenTracker struct { redis *redis.Client }

// Record(agentID, usage, source) → 按 Agent+日期聚合 Redis Hash（key=tokens:agent:{id}:date:{date}, TTL 30天）

//   source="agent"  → agent_prompt_tokens / agent_completion_tokens

//   source="anseragent" → anseragent_prompt_tokens / anseragent_completion_tokens

// GetDailyUsage(agentID, date) → prompt, completion 汇总

```

#### anserAgent 调度阶段 — 回调记录（已有）

```go

// anserAgent 阶段：Eino WithCallbacks → OnEnd 回调中 tokenTracker.Record(agentID, usage, "agent")

```

#### anserAgent 执行阶段 — 双通道采集

**通道 ① 实时：`anserflow agent run --format json` stdout 解析**

```go

// internal/worker/executor.go — anserflow agent run --format json → stdout JSON Lines 解析

//   解析 "token_usage" 字段 → tokenTracker.Record(agentID, ..., "anseragent")

//   解析 "content" 字段 → timelineRepo.Append(issueID, ...)

//   非 JSON 行 → parseStdoutLine 兼容处理

```

**通道 ② 事后汇总：anserAgent session 文件解析**

anserAgent 在 `/home/sandbox/.anseragent/sessions/` 下保存 JSONL 格式的会话文件，每条消息包含 `token_usage` 字段。Worker 在执行完成后从容器中提取：

```go

// internal/worker/session_parser.go — 事后汇总：读取 anserAgent session JSONL → 累加 token_usage → RecordFinal

//   去重：取 max(实时, 事后) 作为最终值，弥补实时 JSON 解析遗漏

func (w *Worker) collectSessionTokens(ctx, containerID, agentID, issueID) error

```

> **双通道去重策略**：实时通道在 anserAgent 执行过程中持续累加 token，事后通道在执行完成后读取 session 文件获得最终精确值。取 `max(实时, 事后)` 作为最终值，覆盖 Redis 中的 anserAgent 部分（通过 `RecordFinal` 实现）。这样即使实时通道部分 JSON 解析失败，事后通道也能兜底。

#### Token 总量公式

```

Agent 总 Token = anserAgent 调度 Token + anserAgent 执行 Token

              = (讨论 + /backlog + 提示词优化) + (编码 + 测试 + 修复 + PR)

```

| 来源 | 包含的 LLM 调用 | 触发位置 |

|------|---------------|---------|

| `agent` | 群聊 Agent 讨论、`/backlog` 方案拆解、`PromptOptimizer.Enhance()` | `anserAgent.Invoke` / `CommandHandler.HandleBacklog` |

| `anseragent` | `anserflow agent run` 全过程（读取代码、生成代码、运行测试、修复错误、commit） | `Worker.executeWithTokenTracking` |

> **LLM API Key 安全模型**：API Key 在 `agents.runtime_config.llm.api_key_encrypted` 中以 AES-256 加密存储；Agent 执行时 Worker 解密后通过环境变量注入 Docker 沙箱，不写入容器文件系统。

#### Token 用量 API 暴露

提供按 Agent/按组织维度的 Token 用量查询接口，用于 Dashboard 成本展示：

```go

// internal/handler/token_handler.go

// GET /api/orgs/:org_id/agents/:agent_id/token-usage?from=&to=  → TokenUsageResponse（含来源明细 + 成本估算）

// GET /api/orgs/:org_id/token-summary?period=7d                → OrgTokenSummary（按 Agent + 按日期聚合）

```

**成本估算函数**：

```go

// internal/agent/cost.go — 按 provider 单价估算（per 1M tokens）：

//   gpt-4o: $2.50/$10.00 | gpt-4o-mini: $0.15/$0.60 | claude-3.5: $3.00/$15.00 | deepseek-v3: $0.14/$0.28

func estimateCost(providerKey string, promptTokens, completionTokens int64) float64

```

---

---
