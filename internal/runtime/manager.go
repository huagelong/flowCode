package runtime

// RuntimeManager 运行时管理器
// 简化管理器，直接初始化 anserAgent 实现
// 保留接口以便未来扩展新的运行时
type RuntimeManager struct {
	adapter RuntimeAdapter
	parser  OutputParser
}

// NewRuntimeManager 创建运行时管理器
func NewRuntimeManager() *RuntimeManager {
	return &RuntimeManager{
		adapter: &AnserAgentAdapter{},
		parser:  &AnserAgentParser{},
	}
}

// GetAdapter 获取运行时适配器
func (m *RuntimeManager) GetAdapter() RuntimeAdapter {
	return m.adapter
}

// GetParser 获取输出解析器
func (m *RuntimeManager) GetParser() OutputParser {
	return m.parser
}

// ResolveConfig 解析运行时配置
func (m *RuntimeManager) ResolveConfig(
	config map[string]interface{},
	decryptedKey string,
) (renderedConfig string, envVars map[string]string, err error) {
	// 渲染配置
	renderedConfig, err = m.adapter.RenderConfig(config)
	if err != nil {
		return "", nil, err
	}

	// 环境变量映射
	envVars = m.adapter.EnvMapping(config, decryptedKey)

	return renderedConfig, envVars, nil
}
