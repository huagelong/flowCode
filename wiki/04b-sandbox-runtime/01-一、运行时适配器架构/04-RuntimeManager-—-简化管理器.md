> ???`docs/plan/04b-sandbox-runtime.md` ? 102 ?
> ???[???](../../README.md) -> [AnserFlow - 沙箱执行运行时](../README.md) -> [一、运行时适配器架构](README.md) -> RuntimeManager — 简化管理器
> ???[???](03-当前实现：anserAgent.md) ? [???](05-Worker-调用示例（运行时无关）.md)
> ?????[??????](README.md) ? [??????](../README.md)

### RuntimeManager — 简化管理器

```go
// internal/runtime/manager.go

type RuntimeManager struct {
    adapter RuntimeAdapter
    parser  OutputParser
}

// NewRuntimeManager 创建运行时管理器（直接初始化 anserAgent）
func NewRuntimeManager() *RuntimeManager {
    return &RuntimeManager{
        adapter: &AnserAgentAdapter{},
        parser:  &AnserAgentParser{},
    }
}

// ResolveConfig 解析运行时配置
func (m *RuntimeManager) ResolveConfig(
    config map[string]interface{},
    decryptedKey string,
) (renderedConfig string, envVars map[string]string, err error) {
    renderedConfig, err = m.adapter.RenderConfig(config)
    if err != nil {
        return "", nil, err
    }
    
    envVars = m.adapter.EnvMapping(config, decryptedKey)
    return renderedConfig, envVars, nil
}
```
