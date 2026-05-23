> 来源：`docs/plan/04b-sandbox-runtime.md` 第 1 行
> 位置：[总目录](../README.md) -> AnserFlow - 沙箱执行运行时
> 相邻：[上一篇](../04-sandbox/README.md) · [下一篇](../06-agent/README.md)
> 相关主题：[上一份源文档](../04-sandbox/README.md) · [下一份源文档](../06-agent/README.md)

# AnserFlow - 沙箱执行运行时

> **编号说明**：`04b` 作为 `04-sandbox.md` 的**运行时补充**存在，覆盖 Docker 镜像构建、entrypoint、Go Docker SDK 调用和 RuntimeAdapter 接口实现。逻辑上属于沙箱模块的子领域，因内容较多（接口定义、镜像细节、扩展指南）而独立成文。阅读时建议先读 [04-sandbox](../04-sandbox/README.md) 了解沙箱整体设计，再读本文档了解运行时的具体实现细节。
>
> **职责边界**：本文档覆盖 Docker 沙箱运行时适配器模式。Agent 基础设施层（Eino 框架、状态机、通知、Git、Token）见 [04-sandbox.md](../04-sandbox/README.md)。
>
> **架构说明**：保留 RuntimeAdapter 和 OutputParser 接口抽象，方便未来扩展新的运行时（如 Claude Code、Codex CLI 等）。当前仅实现 anserAgent 一个运行时。

---

## 章节导航

- [一、运行时适配器架构](01-一、运行时适配器架构/README.md)
- [二、沙箱镜像与执行细节](02-二、沙箱镜像与执行细节/README.md)
- [三、扩展新运行时指南](03-三、扩展新运行时指南/README.md)
