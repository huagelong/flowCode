> 来源：`docs/plan/04b-sandbox-runtime.md` 第 65 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱执行运行时](../README.md) -> [一、运行时适配器架构](README.md) -> 当前实现：anserAgent
> 相邻：[上一篇](02-接口定义.md) · [下一篇](04-RuntimeManager-—-简化管理器.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### 当前实现：anserAgent

```go
// internal/runtime/anseragent.go

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
        workdir, configPath, prompt,
    )
}
```
