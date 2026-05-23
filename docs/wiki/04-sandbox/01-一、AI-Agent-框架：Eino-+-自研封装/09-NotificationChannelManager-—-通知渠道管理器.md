> 来源：`docs/plan/04-sandbox.md` 第 671 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱 / Agent 基础设施](../README.md) -> [一、AI Agent 框架：Eino + 自研封装](README.md) -> NotificationChannelManager — 通知渠道管理器
> 相邻：[上一篇](08-IssueStatusManager-—-Issue-状态机管理器.md) · [下一篇](10-GitManager-—-Git-管理器.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### NotificationChannelManager — 通知渠道管理器

`NotificationService` 中 WS 推送、浏览器通知、群聊系统消息、邮件通知四种渠道的触发逻辑硬编码在多个方法里，通过 `ChannelManager` 统一分发。

**设计原则**：

- 统一入口：`manager.Notify(event, payload)` → 自动分发到用户开通的渠道

- 新增渠道只需注册新 Channel，不改业务代码

- 用户通知偏好统一查询

**核心接口**：

```go

type Event string

const (

    EventIssueAssigned     Event = "issue_assigned"

    EventIssueStatusChange Event = "issue_status_changed"

    EventAgentCompleted    Event = "agent_completed"

    EventMention           Event = "mention"

    EventNewDM             Event = "new_dm"

)

type Channel interface {

    Name() string

    Send(ctx context.Context, userID uint, payload *NotifyPayload) error

}

type ChannelManager struct {

    channels  []Channel

    userPrefs UserPreferenceService

}

// Register 注册渠道 → Notify 查询偏好后按渠道分发

// NotifyGroup 直接写入 messages 表（不受用户偏好控制）

```

**渠道注册**：`WebSocketChannel`、`EmailChannel`、`BrowserChannel` 三种渠道在初始化时注册。
