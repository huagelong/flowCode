> ???`docs/plan/04b-sandbox-runtime.md` ? 1 ?
> ???[???](../README.md) -> AnserFlow - 沙箱执行运行时
> ???[???](../04-sandbox/README.md) ? [???](../06-agent/README.md)
> ?????[??????](../04-sandbox/README.md) ? [??????](../06-agent/README.md)

# AnserFlow - 沙箱执行运行时

> **职责边界**：本文档覆盖 Docker 沙箱运行时适配器模式。Agent 基础设施层（Eino 框架、状态机、通知、Git、Token）见 [04-sandbox.md](../04-sandbox/README.md)。
>
> **架构说明**：保留 RuntimeAdapter 和 OutputParser 接口抽象，方便未来扩展新的运行时。当前仅实现 anserAgent 一个运行时。

---

## ????

- [一、运行时适配器架构](01-一、运行时适配器架构/README.md)
- [二、沙箱镜像与执行细节](02-二、沙箱镜像与执行细节/README.md)
- [三、扩展新运行时指南](03-三、扩展新运行时指南/README.md)
