> 来源：`docs/plan/04b-sandbox-runtime.md` 第 252 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱执行运行时](../README.md) -> [二、沙箱镜像与执行细节](README.md) -> Go Docker SDK
> 相邻：[上一篇](02-entrypoint.sh.md) · [下一篇](04-运行时数据目录.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### Go Docker SDK

```go
// internal/sandbox/ — Docker SDK 核心操作

func ensureProjectContainer(ctx, project, runtime) (string, error)
//   复用已有容器 / 重启停止容器 / 创建新容器（1GB/2CPU/AutoRemove:false）
//   首次 git clone → /workspace/main（worktree 基准）

func createWorktree(ctx, containerID, issue) error
func removeWorktree(ctx, containerID, issueID) error
func execAgent(ctx, containerID, issue, task) error
func destroyProjectSandbox(ctx, projectID, containerID) error
```
