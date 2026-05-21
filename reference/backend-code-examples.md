# AnserFlow - 后端代码示例

> 本文档包含 `docs/plan/02-api.md` 中涉及的实现级代码示例。
> 文档中通过链接引用这些代码，保持规划文档的精炼。

## OAuth2 — GitHub 第三方登录

来源：[02-api.md §框架补充说明](../docs/plan/02-api.md#oauth2--第三方登录github)

### GitHub Callback Handler

```go
// internal/handler/auth.go
func (h *AuthHandler) GitHubCallback(c *gin.Context) {
    code := c.Query("code")
    // 1. code → access_token
    token, _ := h.oauth.Exchange(ctx, code)
    // 2. access_token → GitHub user info
    ghUser, _ := h.oauth.GetUser(ctx, token)
    // 3. 查找或创建用户
    user := h.userRepo.FindOrCreateByGitHub(ghUser)
    // 4. 生成 JWT
    jwtToken, _ := h.jwtService.Generate(user.ID)
    // 5. 重定向到前端（URL 参数携带 JWT）
    c.Redirect(http.StatusFound,
        fmt.Sprintf("/admin/dashboard?token=%s", jwtToken))
}
```

## 国际化（i18n）错误码体系

来源：[02-api.md §框架补充说明](../docs/plan/02-api.md#后端go-i18n-错误码映射实现)

### Gin 错误响应中间件

```go
// internal/middleware/i18n_error.go
package middleware

import (
    "net/http"
    "github.com/BurntSushi/toml"
    "github.com/nicksnyder/go-i18n/v2/i18n"
    "golang.org/x/text/language"
    "github.com/gin-gonic/gin"
)

var bundle *i18n.Bundle

func InitI18n() {
    bundle = i18n.NewBundle(language.English)
    bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
    bundle.MustLoadMessageFile("locales/active.en.toml")
    bundle.MustLoadMessageFile("locales/active.zh-CN.toml")
}

// APIError 统一错误响应结构
type APIError struct {
    Code    string `json:"code"`    // 错误码（如 ERR_ISSUE_NOT_FOUND）
    Message string `json:"message"` // 当前 locale 的文案
}

// RespondError 返回国际化错误响应
func RespondError(c *gin.Context, statusCode int, errorCode string, templateData map[string]interface{}) {
    // 确定 locale：优先用户设置 → Accept-Language 头 → 默认 en
    locale := getUserLocale(c)
    localizer := i18n.NewLocalizer(bundle, locale)

    msg, err := localizer.Localize(&i18n.LocalizeConfig{
        MessageID:    errorCode,
        TemplateData: templateData,
    })
    if err != nil {
        msg = errorCode // fallback 到错误码本身
    }

    c.JSON(statusCode, APIError{Code: errorCode, Message: msg})
    c.Abort()
}

func getUserLocale(c *gin.Context) string {
    // 优先级：已登录用户 users.locale > Accept-Language > 默认 en
    if userLocale, exists := c.Get("user_locale"); exists {
        return userLocale.(string)
    }
    acceptLang := c.GetHeader("Accept-Language")
    if strings.HasPrefix(acceptLang, "zh") {
        return "zh-CN"
    }
    return "en"
}
```

### 翻译文件示例

**中文翻译** (`locales/active.zh-CN.toml`)：

```toml
[ERR_ISSUE_NOT_FOUND]
other = "Issue #{{.IssueID}} 不存在"

[ERR_VALIDATION_FAILED]
other = "请求参数校验失败: {{.Detail}}"

[ERR_PERMISSION_DENIED]
other = "权限不足，无法执行此操作"

[ERR_ORG_LIMIT_EXCEEDED]
other = "组织 {{.OrgName}} 并发 Agent 数已达上限 ({{.Max}})，请等待或提升限额"

[ERR_SANDBOX_TIMEOUT]
other = "Docker 沙箱执行超时（超过 {{.Timeout}} 秒）"
```

**英文翻译** (`locales/active.en.toml`)：

```toml
[ERR_ISSUE_NOT_FOUND]
other = "Issue #{{.IssueID}} not found"

[ERR_VALIDATION_FAILED]
other = "Validation failed: {{.Detail}}"

[ERR_PERMISSION_DENIED]
other = "Permission denied"

[ERR_ORG_LIMIT_EXCEEDED]
other = "Organization {{.OrgName}} has reached the max concurrent agent limit ({{.Max}}). Please wait or upgrade."

[ERR_SANDBOX_TIMEOUT]
other = "Docker sandbox execution timed out (exceeded {{.Timeout}} seconds)"
```

## WebSocket 分布式架构

来源：[02-api.md §一、分布式架构设计](../docs/plan/02-api.md#一分布式架构设计)

### WebSocket 消息协议

```json
{
  "type": "message",
  "channel": "group:42",
  "seq": 12345,
  "payload": {
    "id": "msg_uuid",
    "sender_id": 101,
    "content": "Hello",
    "created_at": "2026-05-21T10:00:00Z"
  }
}
```

### Redis Pub/Sub 订阅实现

```go
// internal/ws/manager.go
func (m *Manager) subscribeToChannels(channels []string) {
    pubsub := m.redis.Subscribe(ctx, channels...)
    defer pubsub.Close()

    ch := pubsub.Channel()
    for msg := range ch {
        var wsMsg WSMessage
        if err := json.Unmarshal([]byte(msg.Payload), &wsMsg); err != nil {
            continue
        }
        m.broadcastToChannel(wsMsg.Channel, wsMsg)
    }
}
```

### 消息持久化写入

```go
// internal/ws/persist.go
func (m *Manager) persistMessage(msg WSMessage) error {
    record := model.Message{
        Channel:   msg.Channel,
        Seq:       msg.Seq,
        SenderID:  msg.Payload.SenderID,
        Content:   msg.Payload.Content,
        CreatedAt: msg.Payload.CreatedAt,
    }
    return m.db.Create(&record).Error
}
```

### 遗漏消息补偿查询

```go
// internal/ws/recovery.go
func (m *Manager) getMissedMessages(channel string, lastSeq int64) ([]WSMessage, error) {
    var messages []model.Message
    err := m.db.Where("channel = ? AND seq > ?", channel, lastSeq).
        Order("seq ASC").
        Limit(500).
        Find(&messages).Error
    
    if err != nil {
        return nil, err
    }
    
    var wsMessages []WSMessage
    for _, msg := range messages {
        wsMessages = append(wsMessages, toWSMessage(msg))
    }
    return wsMessages, nil
}
```

## 任务队列（Asynq）

来源：[02-api.md §一、分布式架构设计](../docs/plan/02-api.md#12-任务队列方案)

### Issue 执行任务入队

```go
// internal/queue/producer.go
func (p *Producer) EnqueueIssueExecution(issueID int64) error {
    payload := map[string]interface{}{
        "issue_id": issueID,
        "priority": "high",
    }
    
    task, err := asynq.NewTask("issue:execute", convert.ToJSON(payload))
    if err != nil {
        return err
    }
    
    _, err = p.client.Enqueue(task, asynq.Queue("critical"))
    return err
}
```

### 任务处理器

```go
// internal/queue/handler.go
func (h *IssueExecutionHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
    var payload struct {
        IssueID int64 `json:"issue_id"`
    }
    if err := json.Unmarshal(task.Payload(), &payload); err != nil {
        return err
    }
    
    // 1. 更新 Issue 状态为 in_progress
    // 2. 分配沙箱容器
    // 3. 启动 anserAgent 执行
    // 4. 等待完成并更新状态
    
    return h.executionService.Run(ctx, payload.IssueID)
}
```

### 调度器对 paused 状态的处理

```go
// internal/queue/scheduler.go
func (s *Scheduler) handlePausedIssue(ctx context.Context) error {
    // 查询所有 paused 状态的 Issue
    var issues []model.Issue
    s.db.Where("status = ? AND paused_at < ?", "paused", time.Now().Add(-1*time.Hour)).
        Find(&issues)
    
    for _, issue := range issues {
        // 自动恢复执行
        if err := s.resumeIssue(ctx, issue.ID); err != nil {
            s.log.Error("Failed to resume issue", "issue_id", issue.ID, "error", err)
        }
    }
    return nil
}
```

### 重试次数限制

```sql
-- 当 retry_count >= 3 时自动回退到 backlog
UPDATE issues 
SET status = 'backlog', 
    updated_at = NOW()
WHERE status = 'in_progress' 
  AND retry_count >= 3
  AND updated_at < NOW() - INTERVAL 30 MINUTE;
```

## RBAC 权限管理

来源：[02-api.md §二、核心数据模型](../docs/plan/02-api.md#20-角色与权限管理rbac)

### Casbin 策略配置

```conf
# config/rbac_model.conf
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

### Casbin 中间件集成

```go
// internal/middleware/casbin.go
func CasbinMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := c.MustGet("user").(*model.User)
        obj := c.Request.URL.Path
        act := c.Request.Method
        
        // 检查权限
        allowed, err := enforcer.Enforce(user.ID, obj, act)
        if err != nil || !allowed {
            RespondError(c, http.StatusForbidden, "ERR_PERMISSION_DENIED", nil)
            return
        }
        
        c.Next()
    }
}
```

## 邮件服务

来源：[02-api.md §二、核心数据模型](../docs/plan/02-api.md#24-邮件服务)

### gomail 发送邮件

```go
// internal/service/email.go
package service

import (
    "gopkg.in/gomail.v2"
)

type EmailService struct {
    dialer *gomail.Dialer
}

func NewEmailService(host string, port int, username, password string) *EmailService {
    d := gomail.NewDialer(host, port, username, password)
    d.SSL = true
    return &EmailService{dialer: d}
}

func (s *EmailService) SendInvite(toEmail, inviteURL, orgName string) error {
    m := gomail.NewMessage()
    m.SetHeader("From", "noreply@anserflow.com")
    m.SetHeader("To", toEmail)
    m.SetHeader("Subject", fmt.Sprintf("加入 %s 组织", orgName))
    m.SetBody("text/html", fmt.Sprintf(`
        <p>您被邀请加入 <strong>%s</strong></p>
        <p><a href="%s">点击接受邀请</a></p>
    `, orgName, inviteURL))
    
    return s.dialer.DialAndSend(m)
}
```
