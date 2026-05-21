package runtime

// RuntimeAdapter 运行时适配器接口
// 封装不同 AI 编码工具的配置注入差异
// 保留此接口以便未来扩展新的运行时
type RuntimeAdapter interface {
	// Name 运行时标识
	Name() string

	// HomeDir 运行时在容器内的主目录（bind mount 目标）
	HomeDir() string

	// ConfigPath 配置文件写入路径（容器内）
	ConfigPath() string

	// RenderConfig 将通用配置渲染为该工具特有的配置格式
	RenderConfig(config map[string]interface{}) (string, error)

	// EnvMapping API Key → 环境变量映射
	EnvMapping(config map[string]interface{}, decryptedKey string) map[string]string

	// SkillsMountPath Skills 目录在容器内的路径
	SkillsMountPath() string

	// SessionPath 会话文件路径（用于事后 Token 汇总），空字符串表示不支持
	SessionPath() string

	// ExecuteCommand 生成执行命令
	ExecuteCommand(workdir, configPath, prompt string) string
}

// OutputParser 输出解析器接口
// 封装不同 AI 工具的 stdout 格式差异
type OutputParser interface {
	// ParseLine 解析单行 stdout 输出
	// 返回 nil 表示该行不是结构化事件（按纯文本处理）
	ParseLine(line []byte) *ParsedEvent

	// ParseSessionFile 解析会话文件（执行完成后调用）
	// 返回 nil 表示该运行时不支持 session 文件
	ParseSessionFile(content []byte) (*TokenSummary, error)
}

// ParsedEvent 解析后的单行事件
type ParsedEvent struct {
	Type       string            // "agent_log" / "token_usage" / "error"
	Content    string            // 文本内容（写入 issue_timeline）
	TokenUsage *TokenUsageDetail // Token 用量（非空时写入 token_tracker）
}

// TokenUsageDetail Token 用量明细
type TokenUsageDetail struct {
	InputTokens      int64
	OutputTokens     int64
	CacheReadTokens  int64
	CacheWriteTokens int64
}

// TokenSummary 会话文件 Token 汇总
type TokenSummary struct {
	TotalInput  int64
	TotalOutput int64
}
