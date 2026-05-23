> ???`docs/plan/11-backlog.md` ? 169 ?
> ???[???](../../README.md) -> [AnserFlow — 远期 Backlog（Phase 2+）](../README.md) -> [工程基础设施（Phase 2）](README.md) -> RTK 命令输出压缩
> ???[???](03-Crowdin-Lokalise-翻译管理.md) ? [???](../04-可参考项目.md)
> ?????[??????](README.md) ? [??????](../README.md)

### RTK 命令输出压缩

> 参考 [yolobox](https://github.com/finbarr/yolobox) 的 RTK（command-output compression proxy）设计思路。

#### 问题

anserAgent 执行过程中通过 stdout 输出大量命令结果（测试日志、构建输出、lint 报告、文件搜索等），当前方案全量写入 `agent_logs` 并通过 WebSocket 推送给前端。体积大、信息密度低，带来三重开销：

| 开销类型 | 影响 | 典型场景 |
|---------|------|---------|
| **存储开销** | MySQL `agent_logs` 表快速增长 | 一次完整 Issue 执行可能产生数千条日志 |
| **传输开销** | WebSocket 推送大量冗余数据 | 前端渲染 500 行测试输出，实际关键行只有 5 行 |
| **Token 开销** | 命令输出回传 LLM 消耗 Token | `npm test` 输出 2000 行，LLM 只需看 "3 failed" 摘要 |

#### 方案

在 Worker 的 stdout 处理流水线中插入 **压缩适配器（OutputCompressor）**，对命令输出做智能摘要后存储/推送/注入：

```
anserAgent stdout
    │
    ▼
┌─────────────────────┐
│ OutputCompressor    │  ← 新增：智能压缩层
│ ├── 错误行识别       │     保留 stderr / FAIL / ERROR 行完整内容
│ ├── 关键行保留       │     保留首行、末行、总结行（如 "Tests: 3 failed, 7 passed"）
│ ├── 中间行抽样       │     每 N 行保留 1 行（可配置）
│ └── 摘要生成         │     生成 "[压缩] 原始 2000 行 → 摘要 30 行" 标记
└──────┬──────────────┘
       │
       ├──→ agent_logs（压缩版）
       ├──→ WebSocket（压缩版）
       └──→ LLM 上下文（极度压缩版）
```

**压缩策略**（分三种场景）：

| 场景 | 策略 | 保留内容 | Token 节省（估算） |
|------|------|---------|-------------------|
| **存储** (`agent_logs`) | 智能摘要 | 错误行 100% + 关键行 + 每 N 行抽样 | ~60-80% |
| **传输** (WebSocket) | 折叠截断 | 错误行 100% + 前端可展开折叠 | ~70-90% |
| **上下文** (LLM) | 极度压缩 | 仅错误行 + 总结行 + 统计摘要 | ~90-95% |

**接口设计**：

```go
// internal/worker/compressor.go

type CompressionLevel int

const (
    CompressStorage   CompressionLevel = iota  // 存储级：智能摘要
    CompressTransport                          // 传输级：折叠截断
    CompressContext                            // 上下文级：极度压缩
)

type OutputCompressor interface {
    // Compress 压缩单行 stdout 输出
    // 返回 nil 表示该行被完全丢弃
    // 返回 CompressedLine 表示压缩结果
    Compress(line string, level CompressionLevel) *CompressedLine
    
    // Summary 生成压缩统计摘要
    Summary() CompressionSummary
}

type CompressedLine struct {
    Text      string `json:"text"`                // 压缩后文本
    Original  string `json:"original,omitempty"`  // 原始文本（仅关键行保留）
    Truncated bool   `json:"truncated"`           // 是否被截断
    KeepFull  bool   `json:"keep_full"`           // 是否保留完整原始内容
}

type CompressionSummary struct {
    OriginalLines int     `json:"original_lines"`
    KeptLines     int     `json:"kept_lines"`
    Ratio         float64 `json:"ratio"`
}
```

**配置文件**：

```yaml
# config.yaml
output_compression:
  enabled: true                # 是否启用（Phase 2 默认开启）
  storage:
    sample_rate: 5             # 每 N 行保留 1 行
    max_lines: 200             # 单次执行最多保留行数
  transport:
    fold_threshold: 500        # 超过此行数触发前端折叠
    error_only: false          # true=仅推送错误行
  context:
    max_lines: 20              # LLM 上下文最大行数
    error_only: true           # 上下文仅注入错误行
    include_summary: true      # 注入统计摘要
```

#### 实施路径

1. **Phase 2-1**：实现 `OutputCompressor` 核心逻辑 + MySQL 存储级压缩
2. **Phase 2-2**：WebSocket 推送接入传输级压缩，前端适配展开/折叠
3. **Phase 2-3**：LLM 上下文注入接入压缩，与 `OutputParser` 协同工作

---
