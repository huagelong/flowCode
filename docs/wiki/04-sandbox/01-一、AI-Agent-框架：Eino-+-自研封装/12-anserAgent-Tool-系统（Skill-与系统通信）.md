> 来源：`docs/plan/04-sandbox.md` 第 953 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱 / Agent 基础设施](../README.md) -> [一、AI Agent 框架：Eino + 自研封装](README.md) -> anserAgent Tool 系统（Skill 与系统通信）
> 相邻：[上一篇](11-TokenManager-—-Token-配额管理器.md) · [下一篇](13-backlog-与-todo-指令识别.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### anserAgent Tool 系统（Skill 与系统通信）

Skill 不只是 Markdown 文档，每个 Skill 声明一组可调用 Tool。anserAgent 调度 LLM 时，LLM 决定调用哪个 Tool → 执行对应的 Go handler → 操作数据库/群聊/通知。

**Skill 定义格式**（YAML frontmatter + Markdown）：

```markdown

---

name: issue-backlog

description: /backlog 方案拆解规范，将群聊讨论产出为一个 Issue

tools:

  - create_issue     # 创建 Issue（调用 IssueService）

  - read_issues      # 读取已有 Issue（防重复）

  - send_message     # 向群聊发送消息

is_builtin: true

---

# 方案拆解规范

## 触发条件

收到 /backlog 指令时调用 create_issue 工具。

## 创建规则

- title: 简洁的功能描述（<50字）

- description: 技术方案概述 + 验收标准

- priority: P0=核心路径 P1=重要功能 P2=增强

- 调用 read_issues 检查是否已存在相同 Issue

- 创建成功后调用 send_message 通知群聊

```

**Tool 注册与调度**：

```go

// internal/agent/tool/registry.go

type ToolRegistry struct {

    tools map[string]ToolHandler

}

type ToolHandler func(ctx context.Context, params json.RawMessage) (string, error)

func NewRegistry(services *Services) *ToolRegistry {

    r := &ToolRegistry{tools: make(map[string]ToolHandler)}

    // 注册 anserAgent 可调用的所有 Tool

    r.Register("create_issue",   services.IssueService.CreateFromAgent)

    r.Register("read_issues",    services.IssueService.ListByProject)

    r.Register("send_message",   services.WS.SendToGroup)

    r.Register("read_timeline",  services.TimelineRepo.FindByIssue)

    r.Register("change_status",  services.IssueService.UpdateStatus)

    r.Register("find_agent",     services.AgentRepo.FindByID)

    return r

}

// GetToolsSchema 生成 OpenAI Function Calling 格式的 tools 定义

func (r *ToolRegistry) GetToolsSchema(skillNames []string) []ToolDef {

    // 根据 Skill 声明的 tools 列表，返回对应的 Function 定义

}

```

```go

// internal/agent/tool/dispatch.go

func (d *Dispatcher) Execute(ctx context.Context, llmOutput string, agent *model.Agent) error {

    // ① 解析 LLM 输出的 JSON: {"tool": "create_issue", "params": {...}}

    var call ToolCall

    json.Unmarshal([]byte(llmOutput), &call)

    // ② Casbin 校验（Agent 是否有权限调用此 Tool）

    if !d.enforcer.Enforce(agent, call.Tool) {

        return fmt.Errorf("Agent %s 无权调用 %s", agent.Name, call.Tool)

    }

    // ③ 执行 Tool → 写入 DB / 发送 WS

    result, err := d.registry.Invoke(ctx, call.Tool, call.Params)

    // ④ 注入 agent_logs 记录

    d.logRepo.Create(ctx, &model.AgentLog{

        AgentID: agent.ID,

        Action:  call.Tool,

        Input:   call.Params,

        Output:  json.RawMessage(result),

        Status:  "success",

    })

    // ⑤ 返回结果给 LLM 继续上下文

    return err

}

```

**anserAgent Tool 与沙箱内 Tool 对比**：

| | anserAgent Tool | 沙箱内 Tool |

|------|---------|------------|

| 运行位置 | Go 后端进程 | Docker 沙箱内 |

| 操作对象 | 系统数据（Issue/消息/时间线） | 代码文件（read/write/bash） |

| 权限 | Casbin RBAC | 沙箱隔离 |

| 注册方式 | `registry.Register(name, handler)` | anserAgent 内置 |

| 典型调用 | `create_issue` / `send_message` | `read` / `write` / `bash` |
