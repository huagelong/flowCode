package runtime

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

// AnserAgentAdapter anserAgent 运行时适配器实现
type AnserAgentAdapter struct{}

func (a *AnserAgentAdapter) Name() string {
	return "anseragent"
}

func (a *AnserAgentAdapter) HomeDir() string {
	return "/home/sandbox/.anseragent"
}

func (a *AnserAgentAdapter) ConfigPath() string {
	return "/home/sandbox/.anseragent/config.yaml"
}

// RenderConfig 渲染 anserAgent 配置（YAML 格式）
func (a *AnserAgentAdapter) RenderConfig(config map[string]interface{}) (string, error) {
	tmpl := `provider: {{.provider}}
model: {{.model}}
max_iterations: {{.max_iterations}}
thinking: {{.thinking}}
`
	t, err := template.New("anseragent_config").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := t.Execute(&buf, config); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// EnvMapping API Key 环境变量映射
func (a *AnserAgentAdapter) EnvMapping(config map[string]interface{}, decryptedKey string) map[string]string {
	env := make(map[string]string)

	// anserAgent 统一使用 ANSERAGENT_API_KEY
	env["ANSERAGENT_API_KEY"] = decryptedKey

	// 根据 provider 设置额外的环境变量
	if provider, ok := config["provider"].(string); ok {
		switch provider {
		case "openai":
			env["LLM_PROVIDER"] = "openai"
			env["OPENAI_API_KEY"] = decryptedKey
		case "anthropic":
			env["LLM_PROVIDER"] = "anthropic"
			env["ANTHROPIC_API_KEY"] = decryptedKey
		case "deepseek":
			env["LLM_PROVIDER"] = "deepseek"
			env["DEEPSEEK_API_KEY"] = decryptedKey
		}
	}

	return env
}

func (a *AnserAgentAdapter) SkillsMountPath() string {
	return "/home/sandbox/.anseragent/skills"
}

func (a *AnserAgentAdapter) SessionPath() string {
	return "/home/sandbox/.anseragent/sessions/*.jsonl"
}

// ExecuteCommand 生成执行命令
func (a *AnserAgentAdapter) ExecuteCommand(workdir, configPath, prompt string) string {
	return fmt.Sprintf(
		"/usr/local/bin/anserflow agent run --workdir %s --config %s --prompt %q --format json",
		workdir,
		configPath,
		prompt,
	)
}

// AnserAgentParser anserAgent 输出解析器实现
type AnserAgentParser struct{}

// ParseLine 解析 anserAgent stdout 输出
func (p *AnserAgentParser) ParseLine(line []byte) *ParsedEvent {
	// anserAgent 输出格式示例：
	// {"type": "log", "content": "正在分析需求..."}
	// {"type": "token_usage", "input": 1500, "output": 800}
	// {"type": "file_created", "path": "src/login.tsx"}

	var event struct {
		Type     string `json:"type"`
		Content  string `json:"content"`
		Input    int64  `json:"input"`
		Output   int64  `json:"output"`
		FilePath string `json:"path"`
	}

	if err := json.Unmarshal(line, &event); err != nil {
		// 非 JSON 行，按纯文本日志处理
		return &ParsedEvent{
			Type:    "agent_log",
			Content: string(line),
		}
	}

	switch event.Type {
	case "log":
		return &ParsedEvent{
			Type:    "agent_log",
			Content: event.Content,
		}
	case "token_usage":
		return &ParsedEvent{
			Type: "token_usage",
			TokenUsage: &TokenUsageDetail{
				InputTokens:  event.Input,
				OutputTokens: event.Output,
			},
		}
	case "file_created", "file_modified", "file_deleted":
		return &ParsedEvent{
			Type:    "agent_log",
			Content: formatFileEvent(event.Type, event.FilePath),
		}
	case "error":
		return &ParsedEvent{
			Type:    "error",
			Content: event.Content,
		}
	default:
		return nil
	}
}

// ParseSessionFile 解析 anserAgent 会话文件
func (p *AnserAgentParser) ParseSessionFile(content []byte) (*TokenSummary, error) {
	// anserAgent session 文件格式：每行一个 JSON 对象，包含 token_usage 字段
	var summary TokenSummary
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		var sessionEntry struct {
			TokenUsage struct {
				Input  int64 `json:"input"`
				Output int64 `json:"output"`
			} `json:"token_usage"`
		}

		if err := json.Unmarshal([]byte(line), &sessionEntry); err != nil {
			continue
		}

		summary.TotalInput += sessionEntry.TokenUsage.Input
		summary.TotalOutput += sessionEntry.TokenUsage.Output
	}

	return &summary, nil
}

func formatFileEvent(eventType, filePath string) string {
	switch eventType {
	case "file_created":
		return fmt.Sprintf("📝 创建文件: %s", filePath)
	case "file_modified":
		return fmt.Sprintf("✏️ 修改文件: %s", filePath)
	case "file_deleted":
		return fmt.Sprintf("🗑️ 删除文件: %s", filePath)
	default:
		return fmt.Sprintf("文件操作: %s - %s", eventType, filePath)
	}
}
