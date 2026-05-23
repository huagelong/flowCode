> 来源：`docs/plan/04b-sandbox-runtime.md` 第 102 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱执行运行时](../README.md) -> [一、运行时适配器架构](README.md) -> RuntimeManager — 简化管理器
> 相邻：[上一篇](03-当前实现：anserAgent.md) · [下一篇](05-Worker-调用示例（运行时无关）.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

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
