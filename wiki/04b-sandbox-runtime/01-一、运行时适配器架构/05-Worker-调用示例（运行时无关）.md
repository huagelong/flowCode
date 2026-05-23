> 来源：`docs/plan/04b-sandbox-runtime.md` 第 135 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱执行运行时](../README.md) -> [一、运行时适配器架构](README.md) -> Worker 调用示例（运行时无关）
> 相邻：[上一篇](04-RuntimeManager-—-简化管理器.md) · [下一篇](../02-二、沙箱镜像与执行细节/README.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### Worker 调用示例（运行时无关）

```go
// internal/worker/executor.go

// 初始化
runtimeMgr := runtime.NewRuntimeManager()
adapter := runtimeMgr.GetAdapter()
parser := runtimeMgr.GetParser()

// 解析配置
renderedConfig, envVars, _ := runtimeMgr.ResolveConfig(
    agentConfig,
    decryptedAPIKey,
)

// 写入配置文件到容器
configPath := adapter.ConfigPath()
writeToContainer(containerID, configPath, renderedConfig)

// 构建执行命令
cmd := adapter.ExecuteCommand(
    issue.Workdir(),
    configPath,
    issue.TaskPrompt(),
)

// 设置环境变量
cmd.Env = buildEnv(envVars)

// 执行并解析 stdout
stdout, _ := cmd.StdoutPipe()
go parseAgentOutput(stdout, parser, issue.ID)

cmd.Run()
```

---
