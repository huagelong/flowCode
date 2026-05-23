> 来源：`docs/plan/04-sandbox.md` 第 173 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱 / Agent 基础设施](../README.md) -> [一、AI Agent 框架：Eino + 自研封装](README.md) -> Agent 运行时配置
> 相邻：[上一篇](02-Eino-初始化与配置.md) · [下一篇](04-ChatModel-调用示例.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### Agent 运行时配置

`agents.runtime_config` JSON 由绑定的运行时决定其 schema。`runtimes.config_schema` 定义了该运行时可配置的所有字段，前端根据 schema 动态生成表单：

**anserAgent 运行时配置示例**（`runtimes.config_schema` 驱动）：

```json

{

  "provider": "openai",

  "model": "gpt-4o",

  "agent": "build",

  "api_key_encrypted": "aes256:xxx",

  "max_iterations": 20,

  "thinking": true

}

```

**配置流转**：

```

Admin UI (Agent 编辑页)

│  ① 下拉选择运行时（anserAgent）

│  ② 前端根据 runtimes.config_schema 动态渲染配置表单

│  ③ 保存 → agents.runtime_config JSON

│

▼

Worker (沙箱启动时)

│  ① 读取 agents.runtime_id → 确定运行时（anserAgent）

│  ② 读取 agents.runtime_config → 填充模板变量

│  ③ 通过 RuntimeClient 接口与沙箱内 anserAgent 建立双向流连接：

│     - SandboxClient: 启动 Docker 容器，通过 ContainerAttach 双向通讯

│     - LocalClient:   启动本地子进程，通过 stdin/stdout 双向通讯

│     - 两种模式使用同一套 JSON Lines 通讯协议

│  ④ 发送任务 → 流式接收日志、状态、结果事件

│  ⑤ 任务结束 → 关闭连接，销毁容器/终止子进程

```
