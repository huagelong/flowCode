# AnserFlow - 沙箱代码示例

> 本文档包含 `docs/plan/04-sandbox.md` 中涉及的实现级代码示例。
> 文档中通过链接引用这些代码，保持规划文档的精炼。

## Eino Agent 框架

来源：[04-sandbox.md §一、AI Agent 框架](../docs/plan/04-sandbox.md#一ai-agent-框架eino--自研封装)

### Eino 初始化与配置

```go
// internal/agent/core/agent.go
package core

import (
    "github.com/cloudwego/eino/components/model"
    "github.com/cloudwego/eino/compose"
)

type AgentConfig struct {
    Provider  string `json:"provider"`   // openai, anthropic, etc.
    Model     string `json:"model"`      // gpt-4, claude-3, etc.
    APIKey    string `json:"api_key"`
    MaxTokens int    `json:"max_tokens"`
    Temp      float64 `json:"temperature"`
}

func NewAgent(cfg AgentConfig) (*Agent, error) {
    chatModel, err := model.NewOpenAIChatModel(ctx, &model.OpenAIConfig{
        APIKey: cfg.APIKey,
        Model:  cfg.Model,
    })
    if err != nil {
        return nil, err
    }
    
    graph := compose.NewGraph[string, string]()
    graph.AddChatModelNode("llm", chatModel)
    
    return &Agent{
        config:    cfg,
        graph:     graph,
        chatModel: chatModel,
    }, nil
}
```

### Agent 运行时配置

```json
{
  "agent_id": "agent_dev_001",
  "provider": "openai",
  "model": "gpt-4-turbo",
  "max_tokens": 8192,
  "temperature": 0.7,
  "tools": ["git_ops", "file_editor", "shell_executor"],
  "skills": ["react-best-practices"],
  "memory_layers": ["L0", "L1", "L2", "L3"]
}
```

### ChatModel 调用示例

```go
// internal/agent/core/agent.go
func (a *Agent) Run(ctx context.Context, prompt string) (string, error) {
    resp, err := a.chatModel.Generate(ctx, []*model.Message{
        {Role: model.System, Content: a.buildSystemPrompt()},
        {Role: model.User, Content: prompt},
    })
    if err != nil {
        return "", err
    }
    
    // 记录 Token 使用
    a.tokenManager.RecordUsage(resp.Usage)
    
    return resp.Message.Content, nil
}
```

## Tool / Skill 抽象

来源：[04-sandbox.md §一、AI Agent 框架](../docs/plan/04-sandbox.md#tool--skill-抽象)

### Tool 接口定义

```go
// internal/agent/tools/interface.go
package tools

type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (string, error)
}
```

### GitOps Tool 实现

```go
// internal/agent/tools/git_ops.go
type GitOpsTool struct {
    workdir string
}

func (t *GitOpsTool) Name() string {
    return "git_ops"
}

func (t *GitOpsTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    action := args["action"].(string)
    
    switch action {
    case "clone":
        return t.clone(ctx, args["url"].(string), args["branch"].(string))
    case "commit":
        return t.commit(ctx, args["message"].(string))
    case "push":
        return t.push(ctx)
    default:
        return "", fmt.Errorf("unknown action: %s", action)
    }
}

func (t *GitOpsTool) clone(ctx context.Context, url, branch string) (string, error) {
    cmd := exec.CommandContext(ctx, "git", "clone", "-b", branch, url, t.workdir)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("git clone failed: %w, output: %s", err, string(output))
    }
    return "Repository cloned successfully", nil
}
```

## Skill 两层继承

来源：[04-sandbox.md §一、AI Agent 框架](../docs/plan/04-sandbox.md#skill-两层继承沙箱执行时)

### 基础 Skill 抽象

```go
// internal/agent/skill/base.go
type Skill struct {
    Name        string
    Version     string
    Description string
    SOP         string // Standard Operating Procedure
}

func (s *Skill) Execute(ctx context.Context, input string) (string, error) {
    // 1. 加载 SOP
    // 2. 构建提示词
    // 3. 调用 LLM
    // 4. 返回结果
    return s.runSOP(ctx, input)
}
```

### 沙箱执行时 Skill

```go
// internal/agent/skill/sandbox_skill.go
type SandboxSkill struct {
    Skill
    Tools []tools.Tool
}

func (s *SandboxSkill) Execute(ctx context.Context, input string) (string, error) {
    // 沙箱技能可访问文件系统、Shell 等工具
    for _, tool := range s.Tools {
        // 根据 SOP 决定调用哪些工具
        if s.shouldUseTool(tool.Name()) {
            result, err := tool.Execute(ctx, buildToolArgs(input))
            if err != nil {
                return "", err
            }
            input = result // 工具输出作为下一步输入
        }
    }
    return s.runSOP(ctx, input)
}
```

## IssueStatusManager 状态机

来源：[04-sandbox.md §一、AI Agent 框架](../docs/plan/04-sandbox.md#issuestatusmanager--issue-状态机管理器)

### 状态机定义

```go
// internal/status/issue_state.go
package status

type IssueStatus string

const (
    StatusBacklog    IssueStatus = "backlog"
    StatusTodo       IssueStatus = "todo"
    StatusInProgress IssueStatus = "in_progress"
    StatusInReview   IssueStatus = "in_review"
    StatusDone       IssueStatus = "done"
    StatusBlocked    IssueStatus = "blocked"
)

// 合法状态流转
var validTransitions = map[IssueStatus][]IssueStatus{
    StatusBacklog:    {StatusTodo, StatusBlocked},
    StatusTodo:       {StatusInProgress, StatusBacklog, StatusBlocked},
    StatusInProgress: {StatusInReview, StatusTodo, StatusBlocked},
    StatusInReview:   {StatusDone, StatusTodo, StatusInProgress},
    StatusDone:       {StatusTodo}, // 返工
    StatusBlocked:    {StatusTodo, StatusBacklog},
}

func CanTransition(from, to IssueStatus) bool {
    allowed, ok := validTransitions[from]
    if !ok {
        return false
    }
    for _, s := range allowed {
        if s == to {
            return true
        }
    }
    return false
}
```

### 状态变更服务

```go
// internal/status/issue_status_manager.go
func (m *IssueStatusManager) Transition(ctx context.Context, issueID int64, newStatus IssueStatus, reason string) error {
    issue, err := m.repo.GetByID(ctx, issueID)
    if err != nil {
        return err
    }
    
    if !CanTransition(IssueStatus(issue.Status), newStatus) {
        return fmt.Errorf("invalid transition: %s → %s", issue.Status, newStatus)
    }
    
    oldStatus := issue.Status
    issue.Status = string(newStatus)
    issue.UpdatedAt = time.Now()
    
    // 记录状态变更日志
    m.logStatusChange(ctx, issueID, oldStatus, string(newStatus), reason)
    
    return m.repo.Save(ctx, issue)
}
```

## NotificationChannelManager

来源：[04-sandbox.md §一、AI Agent 框架](../docs/plan/04-sandbox.md#notificationchannelmanager--通知渠道管理器)

### 通知渠道接口

```go
// internal/notification/channel.go
type NotificationChannel interface {
    Send(ctx context.Context, recipient string, notification Notification) error
    Name() string
}
```

### WebSocket Push 实现

```go
// internal/notification/ws_channel.go
type WSChannel struct {
    manager *ws.Manager
}

func (c *WSChannel) Send(ctx context.Context, recipient string, notification Notification) error {
    userID, err := strconv.ParseInt(recipient, 10, 64)
    if err != nil {
        return err
    }
    
    msg := ws.WSMessage{
        Type:    "notification",
        Channel: fmt.Sprintf("user:%d", userID),
        Payload: notification,
    }
    
    return c.manager.Broadcast(msg)
}
```

### 浏览器通知

```go
// internal/notification/browser_channel.go
func (c *BrowserChannel) Send(ctx context.Context, recipient string, notification Notification) error {
    // 通过 WebSocket 推送浏览器 Notification API 调用指令
    payload := map[string]interface{}{
        "type": "browser_notification",
        "title": notification.Title,
        "body":  notification.Body,
        "icon":  notification.Icon,
    }
    return c.wsClient.Send(recipient, payload)
}
```

## GitManager

来源：[04-sandbox.md §一、AI Agent 框架](../docs/plan/04-sandbox.md#gitmanager--git-管理器)

### ContainerGitOps 实现

```go
// internal/git/container_git_ops.go
type ContainerGitOps struct {
    containerID string
    exec        *docker.Exec
}

func (g *ContainerGitOps) Clone(ctx context.Context, url, branch, workdir string) error {
    cmd := []string{"git", "clone", "-b", branch, url, workdir}
    exitCode, output, err := g.exec.Run(ctx, g.containerID, cmd)
    if err != nil || exitCode != 0 {
        return fmt.Errorf("git clone failed: exit_code=%d, output=%s", exitCode, output)
    }
    return nil
}

func (g *ContainerGitOps) Commit(ctx context.Context, workdir, message string) error {
    cmds := [][]string{
        {"git", "-C", workdir, "add", "."},
        {"git", "-C", workdir, "commit", "-m", message},
    }
    
    for _, cmd := range cmds {
        exitCode, output, err := g.exec.Run(ctx, g.containerID, cmd)
        if err != nil || exitCode != 0 {
            return fmt.Errorf("git command failed: %v, exit_code=%d", cmd, exitCode)
        }
    }
    return nil
}

func (g *ContainerGitOps) Push(ctx context.Context, workdir string) error {
    cmd := []string{"git", "-C", workdir, "push"}
    exitCode, output, err := g.exec.Run(ctx, g.containerID, cmd)
    if err != nil || exitCode != 0 {
        return fmt.Errorf("git push failed: exit_code=%d, output=%s", exitCode, output)
    }
    return nil
}
```

### Worktree 管理

```go
// internal/git/worktree.go
func (g *ContainerGitOps) CreateWorktree(ctx context.Context, repoDir, worktreeName, branch string) error {
    cmd := []string{
        "git", "-C", repoDir,
        "worktree", "add",
        filepath.Join(repoDir, "_worktrees", worktreeName),
        "-b", branch,
    }
    
    exitCode, output, err := g.exec.Run(ctx, g.containerID, cmd)
    if err != nil || exitCode != 0 {
        return fmt.Errorf("create worktree failed: %s", output)
    }
    return nil
}

func (g *ContainerGitOps) RemoveWorktree(ctx context.Context, repoDir, worktreeName string) error {
    cmd := []string{
        "git", "-C", repoDir,
        "worktree", "remove",
        filepath.Join(repoDir, "_worktrees", worktreeName),
    }
    
    exitCode, output, err := g.exec.Run(ctx, g.containerID, cmd)
    if err != nil || exitCode != 0 {
        return fmt.Errorf("remove worktree failed: %s", output)
    }
    return nil
}
```

## TokenManager

来源：[04-sandbox.md §一、AI Agent 框架](../docs/plan/04-sandbox.md#tokenmanager--token-配额管理器)

### Token 配额管理

```go
// internal/agent/token_manager.go
type TokenManager struct {
    db       *gorm.DB
    issueID  int64
    limit    int64
    used     int64
}

func (m *TokenManager) RecordUsage(usage model.TokenUsage) error {
    m.used += usage.TotalTokens
    
    // 写入数据库
    record := model.AgentLog{
        IssueID:    m.issueID,
        PromptTokens: usage.PromptTokens,
        CompletionTokens: usage.CompletionTokens,
        TotalTokens: usage.TotalTokens,
        CreatedAt:  time.Now(),
    }
    return m.db.Create(&record).Error
}

func (m *TokenManager) IsExceeded() bool {
    return m.used >= m.limit
}

func (m *TokenManager) GetRemaining() int64 {
    return m.limit - m.used
}
```

## anserAgent Tool 系统

来源：[04-sandbox.md §一、AI Agent 框架](../docs/plan/04-sandbox.md#anseragent-tool-系统skill-与系统通信)

### Tool 注册与调度

```go
// internal/agent/core/tool_registry.go
type ToolRegistry struct {
    tools map[string]tools.Tool
}

func (r *ToolRegistry) Register(tool tools.Tool) {
    r.tools[tool.Name()] = tool
}

func (r *ToolRegistry) Execute(ctx context.Context, toolName string, args map[string]interface{}) (string, error) {
    tool, ok := r.tools[toolName]
    if !ok {
        return "", fmt.Errorf("tool not found: %s", toolName)
    }
    return tool.Execute(ctx, args)
}
```

### 方案拆解规范 Tool

```go
// internal/agent/tools/spec_breakdown.go
type SpecBreakdownTool struct{}

func (t *SpecBreakdownTool) Name() string {
    return "spec_breakdown"
}

func (t *SpecBreakdownTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    requirement := args["requirement"].(string)
    
    // 调用 LLM 拆解需求
    prompt := fmt.Sprintf(`
将这个需求拆解为可执行的 Issue：

需求：%s

输出格式：
- [ ] Issue 标题（角色标签）描述
- 依赖关系
- 预估工时
`, requirement)
    
    return llm.Generate(ctx, prompt)
}
```
