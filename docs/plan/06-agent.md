# AnserFlow - anserAgent 智能体系统

> 设计参考：Eino ADK（Agent Development Kit）— 统一 Agent 接口、Workflow 编排、中断恢复、Human-in-the-Loop

---

## 一、系统定位

anserAgent 是 AnserFlow 的通用智能体内核，基于 Eino ADK（Go 原生 AI Agent 框架）构建，实现统一的 `Agent` 接口，替代原 eino-* Skills 硬编码编排。承担两大职责：

| 职责 | 场景 | 说明 |
|------|------|------|
| **调度编排** | 群聊讨论、/backlog 方案拆解、/todo 快速建任务 | Agent 参与 IM 群聊协作 |
| **执行编排** | 沙箱编码任务（Issue 执行） | Agent 在 Docker 沙箱内完成编码 |

设计理念对齐 Eino ADK 的三大原则：
- **少写胶水**：统一 `Agent` 接口 + `AsyncIterator` 事件流，编排器无需关心子 Agent 类型
- **快速编排**：内置 Sequential / Parallel / Loop Workflow Agent，复杂流程通过组合而非硬编码
- **更可控**：CheckPointStore 中断恢复 + Human-in-the-Loop 审批，执行过程可暂停、可审计

```
┌─────────────────────────────────────────────────────┐
│  anserAgent                                          │
│                                                      │
│  ┌──────────────────────────────────────────────┐   │
│  │  Eino ADK（底层引擎）                          │   │
│  │  ├── ChatModelAgent  ReAct 推理循环          │   │
│  │  ├── Workflow Agents  Sequential/Parallel/Loop│   │
│  │  ├── Runner         事件流 + CheckPoint 管理  │   │
│  │  └── Callbacks      回调/日志/Token 追踪      │   │
│  └──────────────────────────────────────────────┘   │
│                                                      │
│  ┌────────────┐  ┌────────────┐  ┌───────────────┐ │
│  │ 五层记忆    │  │ Skills 引擎 │  │ 深度分析引擎  │ │
│  │ (L0~L4)   │  │ (自改进)    │  │ (任务/方案)   │ │
│  └────────────┘  └────────────┘  └───────────────┘ │
└─────────────────────────────────────────────────────┘
```

---

## 二、设计基础 — Agent 统一接口

参照 Eino ADK 的 `Agent` 接口设计，anserFlow 中所有 Agent 类型（ChatModel Agent / Workflow Agent / Custom Agent）都实现同一个接口：

```go
// core/interface.go

type Agent interface {
    // Name 返回 Agent 名称，用于日志和 Agent 间发现
    Name(ctx context.Context) string

    // Description 返回 Agent 描述，用于 Supervisor 模式下的任务路由
    Description(ctx context.Context) string

    // Run 执行 Agent，返回异步事件迭代器
    // 每个事件可以是推理输出（Output）、工具调用（Action）或中断（Interrupted）
    Run(ctx context.Context, input *AgentInput) *AsyncIterator[*AgentEvent]
}
```

**核心优势**：
- 编排器（Runner / Orchestrator）统一调度任意 Agent，无需类型断言
- Workflow Agent 的子 Agent 可以是任意类型，支持嵌套组合
- 中断/恢复机制在接口层统一注入，所有 Agent 自动获得可中断能力

---

## 三、目录结构

```
internal/agent/
├── core/                       # anserAgent 内核
│   ├── interface.go            # [新增] Agent 统一接口
│   ├── agent.go                # ChatModelAgent 实现（ReAct 主循环）
│   ├── config.go               # Agent 配置（模型/记忆/Skills 绑定）
│   ├── context.go              # 上下文构建（System Prompt + 五层记忆注入）
│   ├── runner.go               # [新增] Runner 执行容器
│   └── checkpoint.go           # [新增] CheckPointStore 接口
├── workflow/                   # [新增] Workflow Agent — 确定性编排
│   ├── sequential.go           # SequentialAgent（顺序执行）
│   ├── parallel.go             # ParallelAgent（并发执行）
│   └── loop.go                 # LoopAgent（循环执行）
├── hitl/                       # [新增] Human-in-the-Loop
│   ├── approval.go             # 审批中断（如 commit 前确认）
│   ├── review.go               # 审查编辑（Agent 输出人工审阅）
│   └── feedback.go             # 反馈循环（迭代优化）
├── memory/                     # 五层记忆系统
│   ├── manager.go              # MemoryManager（读/写/检索/归档）
│   ├── meta_rules.go           # L0 元规则加载
│   ├── insight_index.go        # L1 记忆索引（路由表）
│   ├── global_facts.go         # L2 全局事实（稳定知识）
│   ├── skill_sop.go            # L3 任务 Skills / SOPs（可复用流程）
│   ├── session_archive.go      # L4 会话归档（长程召回）
│   ├── retriever.go            # 记忆检索（关键词 + 向量相似度）
│   └── wiki_store.go           # Wiki 格式存储（Markdown 文件读写）
├── skill_engine/               # Skills 自改进引擎
│   ├── manager.go              # SkillManager（加载/注册/调度）
│   ├── generator.go            # Skill 自动生成器（结晶化）
│   ├── improver.go             # Skill 自改进（执行后评估 → 优化）
│   ├── validator.go            # Skill 校验器（过时/报错检测）
│   └── rules.go                # Skill 生成规则定义（R01~R05）
├── analysis/                   # 深度分析引擎
│   ├── task_analyzer.go        # 任务拆解 + 依赖分析
│   ├── code_reviewer.go        # 代码审查建议
│   └── plan_evaluator.go       # 方案对比评估
└── orchestrator/               # 编排器（调度 + 执行）
    ├── group_orchestrator.go   # 群聊编排（基于 Runner）
    └── sandbox_orchestrator.go # 沙箱执行编排（基于 Runner + 中断）
```

---

## 四、Workflow Agents — 确定性编排

参照 Eino ADK 的 WorkflowAgents 设计，提供三种预设执行流程的 Agent 类型，与 LLM 动态决策的 ChatModelAgent 形成互补。

### 4.1 SequentialAgent — 顺序执行

```go
// workflow/sequential.go

type SequentialAgentConfig struct {
    Name        string
    Description string
    SubAgents   []Agent   // 按顺序执行的子 Agent 列表
}

func NewSequentialAgent(ctx context.Context, config *SequentialAgentConfig) (Agent, error)
```

**执行规则**：
- 线性执行：严格按照 `SubAgents` 数组顺序依次执行
- History 传递：每个子 Agent 的执行结果追加到 History，后续 Agent 可访问
- 提前退出：任一子 Agent 产生 `Interrupted` 或错误时终止整个流程

**anserFlow 场景**：CI/CD 风格编码流水线
```
需求分析 Agent → 代码生成 Agent → 测试运行 Agent → 报告生成 Agent
```

### 4.2 ParallelAgent — 并发执行

```go
// workflow/parallel.go

type ParallelAgentConfig struct {
    Name        string
    Description string
    SubAgents   []Agent   // 并发执行的子 Agent 列表
}

func NewParallelAgent(ctx context.Context, config *ParallelAgentConfig) (Agent, error)
```

**执行规则**：
- 所有子 Agent 并发启动
- 等待全部完成后汇总结果
- 任一子 Agent 失败将导致 ParallelAgent 整体返回错误

**anserFlow 场景**：多渠道信息搜集（前端文档 + 后端代码 + API 定义并行分析）

### 4.3 LoopAgent — 循环执行

```go
// workflow/loop.go

type LoopAgentConfig struct {
    Name          string
    Description   string
    SubAgents     []Agent   // 每次循环执行的子 Agent 序列
    MaxIterations int       // 最大迭代次数，0 表示无限
}

func NewLoopAgent(ctx context.Context, config *LoopAgentConfig) (Agent, error)
```

**执行规则**：
- 重复执行 SubAgents 序列直到满足退出条件
- 每次迭代的 History 累积，后续迭代可访问全部历史
- 退出条件：最大迭代次数到达 / 子 Agent 产生 ExitAction

**anserFlow 场景**：代码 Review → 修改 → 再 Review 的迭代优化闭环

### 4.4 组合示例：沙箱编码流水线

```go
// 构建编码流水线：分析 → 编码 → 测试循环
// 注：NewAnserAgent 为 anserFlow 的 ChatModelAgent 工厂函数
analyzeAgent := NewAnserAgent(ctx, analyzeConfig)
codeAgent := NewAnserAgent(ctx, codeConfig)
testAgent := NewAnserAgent(ctx, testConfig)

// 内层循环：编码 + 测试（最多 3 次迭代）
codeTestLoop := NewLoopAgent(ctx, &LoopAgentConfig{
    Name:          "CodeTestLoop",
    SubAgents:     []Agent{codeAgent, testAgent},
    MaxIterations: 3,
})

// 外层顺序：分析 → 编码测试循环
pipeline := NewSequentialAgent(ctx, &SequentialAgentConfig{
    Name:      "SandboxPipeline",
    SubAgents: []Agent{analyzeAgent, codeTestLoop},
})
```

---

## 五、中断与恢复

参照 Eino ADK 的中断/恢复机制，通过 CheckPointStore 持久化执行状态，支持长任务暂停、恢复和跨实例迁移。

### 5.1 核心概念

```
Runner.Query(checkPointID="abc")    ──► 执行中 ──► Interrupt ──► 保存状态到 CheckPointStore
                                                                    │
Runner.Resume(checkPointID="abc")   ◄── 用户决策后 ◄────────────────┘
```

| 概念 | 说明 |
|------|------|
| **CheckPointID** | Runner 级别的唯一标识，串联"中断前"和"中断后"的多次运行 |
| **InterruptID** | 标识"在哪里发生了中断"，恢复时需传回同一个 ID |
| **CheckPointStore** | 状态持久化存储（Redis / MySQL），支持跨实例恢复 |
| **ResumeParams** | 恢复时传入的用户数据，如审批结果、补充信息 |

### 5.2 CheckPointStore 接口

```go
// core/checkpoint.go

type CheckPointStore interface {
    // Save 保存 CheckPoint 状态
    Save(ctx context.Context, cp *CheckPoint) error

    // Load 加载 CheckPoint 状态
    Load(ctx context.Context, checkPointID string) (*CheckPoint, error)

    // Delete 删除已完成/已过期的 CheckPoint
    Delete(ctx context.Context, checkPointID string) error
}

type CheckPoint struct {
    ID            string          // CheckPointID
    AgentPath     []string        // Agent 执行路径
    InterruptID   string          // 中断点 ID
    InterruptInfo interface{}     // 中断信息
    State         []byte          // 序列化的执行状态
    Status        CheckPointStatus
}

type CheckPointStatus string

const (
    CheckPointRunning    CheckPointStatus = "running"
    CheckPointInterrupted CheckPointStatus = "interrupted"
    CheckPointResumed    CheckPointStatus = "resumed"
    CheckPointCompleted  CheckPointStatus = "completed"
)
```

### 5.3 典型流程

```go
// 1. 首次执行
iter := runner.Query(ctx, input,
    WithCheckPointID("issue-42"),
)

for {
    event, ok := iter.Next()
    if !ok { break }

    if event.Action != nil && event.Action.Interrupted != nil {
        // 2. Agent 在 git commit 前中断
        interruptID := event.Action.Interrupted.ID
        fmt.Printf("Agent 请求确认: %v\n", event.Action.Interrupted.Info)

        // 3. 展示给用户，等待审批
        approved := askUserApproval()

        // 4. 恢复执行（可在不同进程/机器）
        iter, _ = runner.Resume(ctx, "issue-42", &ResumeParams{
            Targets: map[string]any{
                interruptID: &ApprovalResult{Approved: approved},
            },
        })
    }
}
```

**anserFlow 中的应用**：
- 沙箱编码中，`git commit` / `DB schema change` 前自动中断，等待人工审批
- 群聊调度中，Agent 发现需求不明确时中断，等待 IM 补充信息
- 长时间编码任务中服务重启，恢复后从断点继续

---

## 六、Human-in-the-Loop

参照 Eino ADK 的 HITL 框架，anserFlow 在沙箱编码场景中集成三种人工介入模式：

### 6.1 模式总览

| 模式 | 触发时机 | anserFlow 场景 |
|------|---------|---------------|
| **审批模式** | 工具调用前中断，等待确认 | `git push` / `DB migration` 执行前需人工审批 |
| **审查编辑模式** | 执行前审查并可原地编辑工具参数 | Agent 生成的代码在 `git commit` 前供人工 review 和修改 |
| **追问模式** | Agent 发现信息不足主动中断追问 | Issue 需求不明确时，Agent 在 IM 中追问澄清 |

### 6.2 审批模式示例

```go
// hitl/approval.go — 将任意 Tool 包装为可审批的版本

type ApprovableTool struct {
    tool.InvokableTool
    requireApproval func(input string) bool  // 判断是否需要审批
}

func (t *ApprovableTool) InvokableRun(ctx context.Context, input string) (string, error) {
    if t.requireApproval(input) {
        // 中断执行，等待审批
        return "", &InterruptError{
            InterruptID:   generateInterruptID(),
            InterruptInfo: fmt.Sprintf("即将执行: %s(%s)，是否确认？", t.Info().Name, input),
        }
    }
    return t.InvokableTool.InvokableRun(ctx, input)
}
```

### 6.3 审查编辑模式

Agent 生成代码后、`git commit` 前中断，展示 diff 给人工审查，人工可以：
- **批准**：直接提交
- **拒绝**：Agent 重新生成
- **编辑**：修改工具调用参数后继续执行

---

## 七、Runner 执行容器

参照 Eino ADK 的 `Runner` 设计，作为 Agent 的执行容器，统一管理事件流、中断恢复和 Token 追踪。

### 7.1 Runner 职责

```go
// core/runner.go

type RunnerConfig struct {
    Agent           Agent              // 被执行的 Agent
    CheckPointStore CheckPointStore    // 状态持久化
    EnableStreaming bool               // 是否启用流式输出
    MaxIterations   int                // 最大迭代次数（防无限循环）
}

type Runner struct {
    config    RunnerConfig
    tokenMgr  *TokenManager
}

func NewRunner(ctx context.Context, config RunnerConfig) *Runner

// Query 首次执行或从新起点执行
// input 携带完整的查询上下文（Query / Context / Mode / Tags 等）
func (r *Runner) Query(ctx context.Context, input *AgentInput, opts ...RunOption) *AsyncIterator[*AgentEvent]

// Resume 从中断点恢复执行
func (r *Runner) Resume(ctx context.Context, checkPointID string, params *ResumeParams) (*AsyncIterator[*AgentEvent], error)
```

### 7.2 Runner 工作流程

```
Runner.Query(query, checkPointID)
    │
    ├─► 1. 创建/加载 CheckPoint
    ├─► 2. 构建 AgentInput
    ├─► 3. Agent.Run(ctx, input) → AsyncIterator
    │       │
    │       ├─► AgentEvent{Output}   → 流式推送给调用方
    │       ├─► AgentEvent{Action}   → 记录 Token 用量
    │       └─► AgentEvent{Interrupted} → 保存 CheckPoint → 返回控制权给调用方
    │
    └─► 4. 完成后删除 CheckPoint

Runner.Resume(checkPointID, params)
    │
    ├─► 1. 加载 CheckPoint
    ├─► 2. 恢复执行上下文
    ├─► 3. 将 ResumeParams 注入中断点
    └─► 4. Agent.Run() 从中断点继续
```

### 7.3 与现有模块的关系

- **编排器**：GroupOrchestrator / SandboxOrchestrator 不再直接调用 `Agent.Run()`，而是通过 `Runner.Query()/Resume()` 管理完整生命周期
- **TokenManager**：Runner 在事件流消费过程中自动统计 Token 用量
- **CheckPointStore**：支持 InMemory（开发调试）/ Redis（生产环境）/ MySQL（持久化）三种实现

**anserFlow 场景对应**：
- 群聊调度：`Runner.Query(sessionID)` → 流式推送到 IM
- 沙箱执行：`Runner.Query(issueID)` → 中断 → `Runner.Resume(issueID, approval)` → 继续

---

## 八、五层记忆系统

### 8.1 分层架构

记忆在任务执行过程中持续沉淀，使 Agent 逐步形成稳定且高效的工作方式：

```
┌──────────────────────────────────────────────────┐
│  L0 — 元规则（Meta Rules）                        │
│  Agent 的基础行为规则和系统约束                     │
│  存储：L0-meta-rules.md（单文件）                  │
│  生命周期：永久，极少变                            │
│  谁写：人工定义                                    │
├──────────────────────────────────────────────────┤
│  L1 — 记忆索引（Insight Index）                    │
│  极简索引层，用于快速路由与召回                      │
│  存储：L1-index.md（路由表）                       │
│  生命周期：永久，随下层更新                         │
│  谁写：Agent 自动维护                              │
├──────────────────────────────────────────────────┤
│  L2 — 全局事实（Global Facts）                     │
│  在长期运行过程中积累的稳定知识                      │
│  存储：L2-facts/ 目录（多个 Markdown 文件）          │
│  生命周期：长期稳定，缓慢演进                        │
│  谁写：Agent 从执行中积累                           │
├──────────────────────────────────────────────────┤
│  L3 — 任务 Skills / SOPs                           │
│  完成特定任务类型的可复用流程                        │
│  存储：L3-skills/ 目录（可自改进）                  │
│  生命周期：中期，随实践优化                         │
│  谁写：Agent 自动生成 + 自改进                      │
├──────────────────────────────────────────────────┤
│  L4 — 会话归档（Session Archive）                  │
│  从已完成任务中提炼出的归档记录，用于长程召回         │
│  存储：L4-archive/ 目录（按月归档）                 │
│  生命周期：归档后只读                              │
│  谁写：任务完成后自动归档                           │
└──────────────────────────────────────────────────┘
```

### 8.2 项目级记忆目录

```
/var/lib/anserflow/projects/{project_id}/memory/
├── L0-meta-rules.md              # L0 元规则（单文件，简洁）
├── L1-index.md                    # L1 记忆索引（路由表）
├── L2-facts/                      # L2 全局事实
│   ├── README.md                  #   事实总览
│   ├── tech-stack.md              #   技术栈 & 依赖
│   ├── architecture.md            #   架构决策
│   ├── conventions.md             #   编码规范
│   └── api-contracts.md           #   API 约定
├── L3-skills/                     # L3 任务 Skills / SOPs
│   ├── README.md                  #   Skills 索引
│   ├── react-auth/
│   │   └── SKILL.md               #   可复用流程
│   ├── api-crud/
│   │   └── SKILL.md
│   └── ...
└── L4-archive/                    # L4 会话归档
    ├── README.md                  #   归档索引
    ├── 2026-05/
    │   ├── issue-42-auth-refactor.md
    │   ├── issue-58-token-refresh.md
    │   └── ...
    └── ...
```

### 8.3 各层内容示例

**L0 — 元规则**（`L0-meta-rules.md`）：

```markdown
# 元规则

## 行为约束
- 收到编码任务时，先阅读相关 L2 事实和 L3 Skills，再动手
- 不确定的 API 调用，先查 L2 事实中的 API 约定，不要猜测
- 每次执行完成后，检查是否有新的 L2 事实或 L3 Skill 需要更新

## 沟通风格
- 群聊讨论中简洁发言，每次不超过 3 段
- 发现方案风险时主动提出，不要默认执行

## 安全边界
- 永远不要删除其他 Agent 的代码，除非有明确指令
- 涉及数据库 schema 变更时，必须先讨论再执行
- 不在代码中硬编码密钥或 Token
```

**L1 — 记忆索引**（`L1-index.md`）：

```markdown
# 记忆索引

## 快速路由

| 关键词 | 路由目标 | 层级 |
|--------|---------|------|
| JWT / 认证 / Token | L2: api-contracts.md, L3: react-auth/ | L2+L3 |
| React / 前端 | L2: tech-stack.md | L2 |
| 数据库 / 迁移 | L2: architecture.md | L2 |
| CRUD / 增删改查 | L3: api-crud/ | L3 |
| 登录页 | L4: issue-42-auth-refactor.md | L4 |

## 检索策略
- 编码任务 → 先查 L3 Skills，无匹配则查 L2 事实
- 方案讨论 → 先查 L2 事实 + L4 归档中的历史决策
- 不确定 → L1 索引路由 → 按关键词定位具体文件
```

**L2 — 全局事实**（稳定知识，例如 `L2-facts/tech-stack.md`）：

```markdown
> 来源：项目初始化 + Agent 执行中持续更新
> 更新频率：低（月度级别）

# 技术栈

## 后端
- Go 1.24 + Gin + GORM
- 数据库：MySQL 8.0
- 缓存/队列：Redis 7.0（AOF 持久化）
- AI 框架：Eino（字节 CloudWeGo）
- 沙箱运行时：anserAgent（自研）

## 前端
- Next.js 14（SPA 静态导出）
- UI：shadcn/ui + Tailwind CSS
- 状态：Zustand + TanStack Query

## 沙箱
- Docker SDK for Go
- 运行时：anserAgent
```

**L3 — 任务 Skill/SOP**（可复用流程，例如 `L3-skills/react-auth/SKILL.md`）：

```markdown
---
name: react-auth
version: 2
success_rate: 0.92
use_count: 5
last_used: 2026-05-16
tags: [frontend, react, auth, jwt]
---

# React 认证模块 SOP

## 触发条件
需要实现登录/注册/Token 刷新的前端认证功能

## 标准流程
1. api.ts 添加 Token 拦截器（401 → 自动刷新）
2. use-auth.ts 实现认证状态（Zustand store）
3. ProtectedRoute 组件（路由守卫）
4. 登录页组件（表单 + Zod 校验）

## 已知陷阱
- JWT 过期需统一 401 拦截，不要每个请求单独处理
- bcrypt 在后端做，前端只传明文密码

## 改进历史
| 版本 | 日期 | 变更 |
|------|------|------|
| v1 | 2026-05-10 | 初始生成（从 Issue #42 提取） |
| v2 | 2026-05-16 | 修复：添加 Token 刷新逻辑（Issue #58 失败后改进） |
```

**L4 — 会话归档**（任务完成后提炼，例如 `L4-archive/2026-05/issue-42-auth-refactor.md`）：

```markdown
> 归档时间：2026-05-16
> 关联 Issue：#42
> 关联 Skill：react-auth (v2)

# Issue #42 — 登录模块重构

## 任务摘要
将 Session 认证迁移到 JWT + bcrypt

## 关键决策
- JWT 无状态，水平扩展友好
- bcrypt cost = 12

## 踩过的坑
- 前端未处理 401 → 需统一拦截器
- 时序攻击：需 ConstantTimeCompare

## 经验贡献
- 更新了 L3: react-auth（添加 Token 刷新步骤）
- 更新了 L2: api-contracts（添加 JWT 约定）
```

### 8.4 核心接口

```go
// memory/manager.go
type MemoryManager struct {
    store     WikiStore           // Markdown 文件读写
    retriever MemoryRetriever     // 检索器
}

// Recall 检索记忆（按层级路由）
func (m *MemoryManager) Recall(ctx context.Context, req RecallRequest) ([]Memory, error)

// Store 存储记忆（根据层级写入对应文件）
func (m *MemoryManager) Store(ctx context.Context, mem Memory) error

// SummarizeAndArchive 将会话总结归档到 L4
func (m *MemoryManager) SummarizeAndArchive(ctx context.Context, sessionID string) error

// UpdateIndex 同步更新 L1 索引
func (m *MemoryManager) UpdateIndex(ctx context.Context, projectID uint) error

type RecallRequest struct {
    AgentID   uint
    ProjectID uint
    Query     string            // 检索查询
    Layers    []MemoryLayer     // 指定检索层级
    Limit     int               // 返回条数
}

type Memory struct {
    ID        string
    Layer     MemoryLayer       // L0 / L1 / L2 / L3 / L4
    Content   string
    FilePath  string            // 来源 Markdown 文件路径
    Tags      []string
    Source    string            // session_id / issue_id / manual
    CreatedAt time.Time
}

type MemoryLayer string

const (
    LayerL0 MetaRules    MemoryLayer = "L0"  // 元规则
    LayerL1 InsightIndex MemoryLayer = "L1"  // 记忆索引
    LayerL2 GlobalFacts  MemoryLayer = "L2"  // 全局事实
    LayerL3 SkillSOP     MemoryLayer = "L3"  // 任务 Skills / SOPs
    LayerL4 SessionArchive MemoryLayer = "L4" // 会话归档
)
```

### 8.5 记忆读写时机

| 时机 | 操作 | 层级 |
|------|------|------|
| Agent 被调用 | Recall 相关记忆 | L1 路由 → L2 + L3 + L4 |
| Agent 调用中 | Store 中间结果到上下文 | 运行时内存（不持久化） |
| Agent 调用结束 | 回复内容记入会话 | 运行时内存 |
| /new 切换会话 | SummarizeAndArchive 当前会话 | L4（归档） |
| Issue 执行完成 | 执行经验总结写入文件 | L4（归档）+ L2（新事实） |
| 同类任务 ≥3 次 | MaybeCrystallize 生成新 SOP | L3（Skills） |
| Skill 执行失败 | Improve 失败 Skill | L3（Skills 自改进） |
| L2/L3/L4 更新后 | UpdateIndex 同步路由表 | L1（索引） |

### 8.6 更新频率与权限

| 层级 | 更新频率 | 谁能写 | 自动/人工 |
|------|---------|--------|----------|
| L0 元规则 | 极低（月度/季度） | 人工 | 人工 |
| L1 索引 | 跟随 L2-L4 更新 | Agent 自动 | 自动 |
| L2 全局事实 | 低（周度） | Agent 积累 + 人工审核 | 半自动 |
| L3 Skills | 中（每次执行后评估） | Agent 自动生成 + 自改进 | 自动 |
| L4 会话归档 | 高（每次任务完成） | Agent 自动归档 | 自动 |

---

## 九、Skills 自改进引擎

### 9.1 Skill 生命周期

```
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│  定义    │───►│  加载    │───►│  执行    │───►│  评估    │───►│  改进    │
│ (创建)   │    │ (注册)   │    │ (调用)   │    │ (打分)   │    │ (重写)   │
└─────────┘    └─────────┘    └─────────┘    └─────────┘    └─────────┘
     ▲                                              │              │
     └──────────────────── 结晶化（自动生成新 Skill）─┘              │
     ▲                                                             │
     └─────────────────── 改进后的 Skill 覆盖原版本 ─────────────────┘
```

### 9.2 Skill 自动生成规则

```go
// skill_engine/rules.go — Skill 自动生成规则定义

type SkillGenerationRule struct {
    Trigger         SkillTrigger   // 何时触发自动生成
    Strategy        GenStrategy    // 如何生成
    MinScore        float64        // 最低质量分（0-1）
    RequireApproval bool           // 是否需要人工审批
}

type SkillTrigger string

const (
    TriggerOnComplexTask   SkillTrigger = "on_complex_task"    // 复杂任务完成后
    TriggerOnRepeatedTask  SkillTrigger = "on_repeated_task"   // 同类任务 ≥3 次
    TriggerOnFailure       SkillTrigger = "on_failure"         // 执行失败后分析
    TriggerOnManual        SkillTrigger = "on_manual"          // 人工请求创建
)

type GenStrategy string

const (
    StrategyExtract       GenStrategy = "extract"              // 从执行日志提取模式
    StrategyFailure       GenStrategy = "failure_learning"     // 失败经验总结
    StrategyBestPractice  GenStrategy = "best_practice"        // 多次成功归纳最佳实践
)
```

**预定义自动生成规则**：

| 规则 ID | 触发条件 | 生成策略 | 说明 |
|---------|---------|---------|------|
| `R01` | 同类 Issue 执行 ≥3 次成功 | `best_practice` | 提取编码模式生成 Skill |
| `R02` | Issue 执行失败后重试成功 | `failure_learning` | 总结踩坑经验 |
| `R03` | 跨 ≥3 个文件的重构任务 | `extract` | 生成重构流程 Skill |
| `R04` | 单次执行 Token > 50k | `extract` | 高成本任务提取优化模式 |
| `R05` | 人工 `/skill save` 指令 | `extract` | 手动保存当前会话为 Skill |

### 9.3 Skill 结晶化（借鉴 GenericAgent）

```go
// skill_engine/generator.go

type SkillGenerator struct {
    chatModel model.ChatModel
    store     SkillStore
}

// Crystallize 将执行路径结晶为 Skill
func (g *SkillGenerator) Crystallize(ctx context.Context, record ExecutionRecord) (*Skill, error) {
    // 1. 检查是否满足结晶条件
    if !g.shouldCrystallize(record) {
        return nil, nil
    }

    // 2. 提取执行路径（从 agent_logs 中）
    path := g.extractExecutionPath(record)

    // 3. LLM 浓缩为 SOP（信息密度最大化）
    sop := g.condenseToSOP(ctx, path)

    // 4. 写入 L3 Skills 目录（Markdown 文件）
    skill := g.writeSkillFile(sop)

    // 5. 更新 L1 索引
    g.updateIndex(skill)

    return skill, nil
}

func (g *SkillGenerator) shouldCrystallize(record ExecutionRecord) bool {
    switch {
    case record.IsNewTaskType:
        return true
    case record.ToolCallCount >= 5:
        return true
    case record.TokenUsed > 30000:
        return true
    default:
        return false
    }
}
```

### 9.4 Skill 自改进

```go
// skill_engine/improver.go

type SkillImprover struct {
    chatModel model.ChatModel
    store     SkillStore
}

type ImproveRequest struct {
    SkillID    uint
    Execution  ExecutionRecord
    Error      *TaskError        // nil 表示成功但可优化
}

func (s *SkillImprover) Improve(ctx context.Context, req ImproveRequest) error {
    if !s.needsImprovement(req) {
        return nil
    }
    analysis := s.analyze(ctx, req)
    improved := s.generateImproved(ctx, req.SkillID, analysis)
    if err := s.validate(improved); err != nil {
        return err
    }
    return s.store.Save(ctx, improved)
}

func (s *SkillImprover) needsImprovement(req ImproveRequest) bool {
    switch {
    case req.Error != nil:
        return true                    // 执行报错 → 必须改进
    case req.Execution.IsOutdated():
        return true                    // 内容过时 → 改进
    case req.Execution.Score < 0.6:
        return true                    // 质量评分低 → 改进
    default:
        return false
    }
}
```

### 9.5 Skill 过时检测

| 维度 | 检测方式 | 示例 |
|------|---------|------|
| **API 版本** | 对比 Skill 引用的 SDK 版本与项目当前版本 | "go-github/v67" → 项目已升级到 v68 |
| **库依赖** | 检查 Skill 引用的库是否已废弃 | "request" → 项目已迁移到 "httpx" |
| **目录结构** | 检查 Skill 引用的路径是否仍存在 | "internal/handler/" → 已重构为 "internal/api/" |
| **编码规范** | 对比 Skill 中的规范与 L2 conventions | Skill 说 "var 命名" → 规范已改为 "camelCase" |
| **执行成功率** | 统计近 N 次执行的成功率 | 最近 5 次执行 3 次失败 → 标记为 broken |

---

## 十、Agent 主循环

### 10.1 Agent 统一接口

> 接口定义详见 [二、设计基础 — Agent 统一接口](#二设计基础--agent-统一接口)。anserAgent 作为 ChatModelAgent 实现该接口，Workflow Agents（Sequential/Parallel/Loop）也实现同一接口，使得编排器可以统一调度任意 Agent 类型。

### 10.2 anserAgent 实现

```go
// core/agent.go

type anserAgent struct {
    id         uint
    name       string
    role       string               // 角色人设
    chatModel  model.ChatModel      // Eino ChatModel
    memory     *memory.Manager      // 五层记忆
    skills     *skill_engine.Manager
    tools      *ToolRegistry        // Eino Tool 注册表
    analyzer   *analysis.Engine
}

func (a *anserAgent) Name(ctx context.Context) string { return a.name }
func (a *anserAgent) Description(ctx context.Context) string { return a.role }

// Run 实现 Agent 接口，返回异步事件流
func (a *anserAgent) Run(
    ctx context.Context,
    input *AgentInput,
) *AsyncIterator[*AgentEvent] {
    iter := NewAsyncIterator[*AgentEvent]()

    go func() {
        defer iter.Close()

        // 1. 五层记忆注入
        memories := a.memory.Recall(ctx, RecallRequest{
            AgentID: a.id, Query: input.Query, Limit: 10,
        })

        // 2. 构建 System Prompt（L0 元规则 + 角色人设）
        systemPrompt := a.buildSystemPrompt(memories)

        // 3. 加载 L3 Skills → Eino Tools
        tools := a.skills.LoadAsTools(ctx, a.id)

        // 4. 组装消息
        messages := a.assembleMessages(systemPrompt, memories, input)

        // 5. ReAct 循环：Reason → Act → Observe
        for {
            resp, err := a.chatModel.Generate(ctx, messages,
                model.WithTools(tools),
                model.WithCallbacks(a.tokenCallback()),
            )
            if err != nil {
                iter.Emit(&AgentEvent{Err: err})
                return
            }

            // 推送推理事件
            iter.Emit(&AgentEvent{
                AgentName: a.name,
                Output: &AgentOutput{
                    MessageOutput: &schema.Message{Role: schema.Assistant, Content: resp.Content},
                },
            })

            // 如果没有工具调用，结束循环
            if len(resp.ToolCalls) == 0 {
                // 后处理
                a.memory.Store(ctx, Memory{
                    Layer: LayerL4, Content: resp.Content, Tags: input.Tags,
                })
                // Skill 结晶化评估（异步，使用独立 context 防止父 ctx 取消后丢失）
                go a.skills.MaybeCrystallize(context.Background(), input, resp)
                return
            }

            // 6. 执行工具调用（Act）
            for _, tc := range resp.ToolCalls {
                result := a.executeTool(ctx, tc)
                iter.Emit(&AgentEvent{
                    AgentName: a.name,
                    Action: &AgentAction{
                        ToolCall: tc,
                        ToolResult: result,
                    },
                })
                // 工具结果追加到消息历史（Observation）
                messages = append(messages, result.AsMessage())
            }
        }
    }()

    return iter
}

type AgentInput struct {
    Query     string
    Context   []*schema.Message
    Mode      AgentMode           // orchestrate / execute
    IssueID   *uint
    SessionID string
    Tags      []string
}

type AgentMode string

const (
    ModeOrchestrate AgentMode = "orchestrate"  // 调度编排
    ModeExecute     AgentMode = "execute"      // 沙箱执行编排
)

type AgentEvent struct {
    AgentName string
    Output    *AgentOutput
    Action    *AgentAction
    Err       error
}

type AgentOutput struct {
    MessageOutput *schema.Message
}

type AgentAction struct {
    ToolCall   interface{}
    ToolResult interface{}
    // Interrupted 非 nil 时表示 Agent 在此处中断
    Interrupted *InterruptContext
}

type InterruptContext struct {
    ID   string      // 中断唯一 ID，恢复时需传回
    Info interface{} // 中断信息（供用户决策）
}
```

### 10.3 上下文构建（信息密度最大化）

```go
// core/context.go — 分层记忆注入，控制总 Token 预算

func (a *anserAgent) buildSystemPrompt(memories []Memory) string {
    var sb strings.Builder

    // L0 元规则（永远注入，~500 token）
    sb.WriteString(a.loadMetaRules())

    // L1 索引（路由表，~200 token）
    sb.WriteString(a.loadInsightIndex())

    // L2 全局事实（按相关性筛选，~2K token）
    facts := a.selectRelevantFacts(memories, 2000)
    sb.WriteString(formatFacts(facts))

    // L3 Skills（匹配到的 SOP，~3K token）
    skills := a.matchSkills(memories)
    sb.WriteString(formatSkills(skills))

    // L4 会话归档（最近相关的，~2K token）
    archives := a.searchArchives(memories, 2000)
    sb.WriteString(formatArchives(archives))

    // 总计约 ~8K token 系统提示，留 ~22K 给对话
    return sb.String()
}
```

### 10.4 调度编排模式（基于 Runner）

```go
// orchestrator/group_orchestrator.go

func (o *GroupOrchestrator) OnMessage(ctx context.Context, msg Message) {
    agent := o.agentRegistry.Get(msg.GroupID)

    input := &AgentInput{
        Query:     msg.Content.Text,
        Context:   o.getSessionHistory(msg.GroupID, msg.SessionID),
        Mode:      ModeOrchestrate,
        SessionID: msg.SessionID,
    }

    // 创建 Runner，配置 CheckPointStore
    // 注：Runner 为轻量对象，可按每次消息创建；若需跨消息复用中断恢复，
    // 应在 Orchestrator 级别缓存 Runner 实例并复用同一 CheckPointID
    runner := NewRunner(ctx, RunnerConfig{
        Agent:           agent,
        CheckPointStore: o.checkPointStore,
        EnableStreaming: true,
    })

    // 异步事件流消费
    iter := runner.Query(ctx, input, WithCheckPointID(msg.SessionID))
    for {
        event, ok := iter.Next()
        if !ok {
            break
        }
        if event.Err != nil {
            o.sendErrorMessage(msg.GroupID, event.Err)
            return
        }
        // 流式推送推理/工具调用事件到 IM
        if event.Output != nil && event.Output.MessageOutput != nil {
            o.streamToIM(msg.GroupID, event.Output.MessageOutput.Content)
        }
    }

    // 会话结束时归档到 L4
    if o.isSessionEnding() {
        o.memory.SummarizeAndArchive(ctx, msg.SessionID)
    }
}
```

### 10.5 沙箱执行编排模式（基于 Runner + 中断）

```go
// orchestrator/sandbox_orchestrator.go

func (s *SandboxOrchestrator) ExecuteIssue(ctx context.Context, issue *Issue) error {
    agent := s.agentRegistry.GetByID(issue.AgentID)

    // 加载项目级长期记忆（L2 + L3）
    memories := s.memory.Recall(ctx, RecallRequest{
        AgentID: issue.AgentID, ProjectID: issue.ProjectID,
        Query: issue.Title, Layers: []MemoryLayer{LayerL2, LayerL3},
    })

    input := &AgentInput{
        Query:   issue.BuildTaskPrompt(),
        Mode:    ModeExecute,
        IssueID: &issue.ID,
        Tags:    s.extractTags(issue),
    }

    checkPointID := fmt.Sprintf("issue-%d", issue.ID)

    runner := NewRunner(ctx, RunnerConfig{
        Agent:           agent,
        CheckPointStore: s.checkPointStore,
        EnableStreaming: true,
    })

    iter := runner.Query(ctx, input, WithCheckPointID(checkPointID))
    for {
        event, ok := iter.Next()
        if !ok {
            break
        }

        // 处理中断事件（如 commit 前审批）
        if event.Action != nil && event.Action.Interrupted != nil {
            interruptID := event.Action.Interrupted.ID
            // 将中断信息推送给用户，等待审批
            approval := s.waitForApproval(ctx, event.Action.Interrupted.Info)
            // 恢复执行
            iter, _ = runner.Resume(ctx, checkPointID, &ResumeParams{
                Targets: map[string]any{interruptID: approval},
            })
            continue
        }

        if event.Err != nil {
            // 执行失败 → 触发 Skill 自改进
            s.skillImprover.Improve(ctx, ImproveRequest{
                SkillID:   s.detectRelatedSkill(issue),
                Error:     wrapTaskError(event.Err),
            })
            return event.Err
        }

        // 记录执行事件供后续结晶化评估
        s.executionRecorder.Record(event)
    }

    // 执行成功 → 评估是否结晶为新 Skill
    s.skillGenerator.MaybeCrystallize(ctx, s.executionRecorder.Finalize())

    return nil
}
```

---

## 十一、与其他模块的集成

```
┌───────────────────────────────────────────────────────┐
│  Go 后端（Gin + Asynq）                                │
│                                                        │
│  Hub.OnMessage()                                       │
│       │                                                │
│       ▼                                                │
│  Runner.Query(mode=orchestrate, checkPointID)          │
│       │                                                │
│       ├── Agent.Run() ──► AsyncIterator[*AgentEvent]  │
│       │     ├── MemoryManager ──► memory/ (L0~L4)     │
│       │     ├── SkillManager  ──► L3-skills/ + MySQL  │
│       │     └── Eino ChatModel ──► LLM API            │
│       ├── CheckPointStore ──► Redis / MySQL           │
│       └── TokenManager ──► 用量追踪                   │
│                                                        │
│  Worker（Asynq）                                      │
│       │                                                │
│       ▼                                                │
│  Runner.Run(mode=execute, checkPointID)               │
│       │                                                │
│       ├── Agent.Run() ──► AsyncIterator[*AgentEvent]  │
│       │     ├── MemoryManager ──► L2 事实 + L3 Skills │
│       │     ├── SkillManager  ──► anser-coder 等      │
│       │     └── RuntimeAdapter ──► anserAgent 沙箱    │
│       ├── CheckPointStore ──► Redis / MySQL           │
│       └── HITL ──► 审批中断 / 审查编辑                │
└───────────────────────────────────────────────────────┘
```

| 模块 | anserAgent 关系 |
|------|----------------|
| **Eino ADK** | Agent 接口 + ChatModelAgent + Workflow Agents + Runner + 中断/恢复 |
| **Eino Graph** | 底层引擎，提供 ChatModel / Graph / Tool / Callbacks |
| **RuntimeAdapter** | 执行模式的行动层（anserAgent 沙箱） |
| **MemoryManager** | 五层记忆读写，Markdown Wiki 文件存储 |
| **SkillManager** | Skills 加载、自动生成、自改进 |
| **CheckPointStore** | 执行状态持久化，支持中断恢复（Redis / MySQL） |
| **Runner** | Agent 执行容器，管理事件流、中断、Token 追踪 |
| **GitManager** | 执行模式中调用 GitOps |
| **SandboxManager** | 执行模式中创建/管理沙箱容器 |
| **TokenManager** | 记录所有 LLM Token 用量 |
| **NotificationManager** | 调度模式中推送消息到 IM |

---

## 十二、数据库表

### 12.1 agent_checkpoints

```sql
CREATE TABLE agent_checkpoints (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    checkpoint_id VARCHAR(128) NOT NULL UNIQUE,  -- Runner 级别的唯一 CheckPoint ID
    agent_id BIGINT NOT NULL,
    project_id BIGINT NOT NULL,
    org_id BIGINT NOT NULL,
    agent_path JSON NOT NULL,                     -- Agent 执行路径（用于恢复定位）
    interrupt_id VARCHAR(128),                    -- 中断点 ID（关联 InterruptContext）
    interrupt_info TEXT,                          -- 中断信息（供用户决策）
    state_data MEDIUMTEXT NOT NULL,               -- 序列化的执行状态
    status ENUM('running','interrupted','resumed','completed','failed') DEFAULT 'running',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_agent_project (agent_id, project_id),
    INDEX idx_status (status),
    FOREIGN KEY (agent_id) REFERENCES agents(id),
    FOREIGN KEY (project_id) REFERENCES projects(id)
);
```

### 12.2 agent_memories

```sql
CREATE TABLE agent_memories (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    agent_id BIGINT NOT NULL,
    project_id BIGINT NOT NULL,
    org_id BIGINT NOT NULL,
    layer ENUM('L0','L1','L2','L3','L4') NOT NULL,
    title VARCHAR(256),
    content TEXT NOT NULL,                  -- 记忆内容摘要
    file_path VARCHAR(512),                 -- Markdown 文件路径（L2~L4）
    tags JSON,                              -- 标签
    source ENUM('auto_summary','skill_created','manual','execution') NOT NULL,
    session_id VARCHAR(64),
    issue_id BIGINT,
    relevance_score FLOAT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_agent_project (agent_id, project_id),
    INDEX idx_project_layer (project_id, layer),
    FOREIGN KEY (agent_id) REFERENCES agents(id),
    FOREIGN KEY (project_id) REFERENCES projects(id)
);
```

### 12.3 skill_generation_logs

```sql
CREATE TABLE skill_generation_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    agent_id BIGINT NOT NULL,
    project_id BIGINT NOT NULL,
    rule_id VARCHAR(16) NOT NULL,           -- R01~R05
    trigger_type ENUM('on_complex_task','on_repeated_task','on_failure','on_manual') NOT NULL,
    strategy ENUM('extract','failure_learning','best_practice') NOT NULL,
    source_issue_id BIGINT,
    skill_name VARCHAR(128),
    skill_version INT DEFAULT 1,
    quality_score FLOAT,
    status ENUM('generated','approved','rejected','improved') DEFAULT 'generated',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (agent_id) REFERENCES agents(id),
    FOREIGN KEY (project_id) REFERENCES projects(id)
);
```

---

## 十三、从 eino-* Skills 到 anserAgent 的迁移映射

原 eino-* Skills 的功能全部由 anserAgent + Eino ADK 接管：

| 原 Skill | 迁移到 | 说明 |
|----------|--------|------|
| ~~eino-discuss~~ | **L0 元规则** | 讨论行为约束内化为元规则 |
| ~~eino-backlog~~ | **L3 Skill** | /backlog 方案拆解作为可自改进的 SOP |
| ~~eino-optimizer~~ | **SkillImprover** | 提示词优化逻辑由自改进引擎接管 |
| ~~eino-planner~~ | **L3 Skill** | 任务编排 SOP，可自动生成 |
| ~~eino-* 硬编码编排~~ | **Workflow Agents + Runner** | 编排逻辑交由 Eino ADK Sequential/Parallel/Loop Agent + Runner 统一管理 |
| anser-coder | 保留 | 沙箱执行规范，不变 |

---

## 十四、当前阶段范围

| 功能 | Phase 1（当前） | [Phase 2](11-backlog.md) |
|------|:---:|:---:|
| Agent 统一接口（实现 Eino ADK Agent 接口） | ✅ | ✅ |
| anserAgent 内核（Eino ChatModelAgent ReAct） | ✅ | ✅ |
| Workflow Agents（Sequential / Parallel / Loop） | ✅ | ✅ |
| Runner 执行容器（事件流 + CheckPoint 管理） | ✅ | ✅ |
| 中断与恢复（CheckPointStore） | ✅ | ✅ |
| 五层记忆（L0~L4）+ Wiki 文件存储 | ✅ | ✅ |
| Skill 自动生成（R01~R05） | ✅ | ✅ |
| Skill 过时检测 + 自改进 | ✅ | ✅ |
| 调度编排 + 沙箱执行编排 | ✅ | ✅ |
| Human-in-the-Loop（审批中断） | ❌ | ✅ |
| 记忆语义检索（向量） | ❌ | ✅ |
| Multi-Agent 范式（Supervisor / Plan-Execute） | ❌ | ✅ |
| Skill 版本回滚 | ❌ | ✅ |
