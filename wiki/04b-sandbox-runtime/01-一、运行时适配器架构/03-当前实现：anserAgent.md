> ???`docs/plan/04b-sandbox-runtime.md` ? 65 ?
> ???[???](../../README.md) -> [AnserFlow - 沙箱执行运行时](../README.md) -> [一、运行时适配器架构](README.md) -> 当前实现：anserAgent
> ???[???](02-接口定义.md) ? [???](04-RuntimeManager-—-简化管理器.md)
> ?????[??????](README.md) ? [??????](../README.md)

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
