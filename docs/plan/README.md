# AnserFlow 文档索引

## 文档关系图

```
docs/
├── ddl.sql                      ← 数据库 DDL 建表语句 + 种子数据
└── plan/
    ├── README.md                  ← 本文档索引
    ├── 01-admin.md                ← 后台管理：配置体系 + 嵌入 SPA
    ├── 02-api.md                  ← API / 后端：框架、数据模型、路由
    ├── 03-client.md               ← 客户端前端：IM 界面、时间线、国际化
    ├── 04-sandbox.md              ← Agent 基础设施：Eino 框架、状态机、通知、Git、Token
    ├── 04b-sandbox-runtime.md     ← 沙箱执行运行时：SandboxManager、RuntimeManager
    ├── 06-agent.md                ← anserAgent 智能体系统
    ├── 07-architecture.md         ← 系统愿景、技术栈、约束范围
    ├── 08-deployment.md           ← 部署指南
    ├── 09-ci-cd.md                ← CI/CD 工作流
    ├── 10-roadmap.md              ← 开发路线图 + 数据库迁移
    └── 11-backlog.md              ← 远期 backlog 规划
```

## 阅读顺序建议

| 阅读顺序 | 文件 | 目的 |
|---------|------|------|
| 1 | `07-architecture.md` | 先了解系统整体愿景、技术栈选型与当前约束 |
| 2 | `01-admin.md` | 后台管理体系：配置、热更新、嵌入 |
| 3 | `02-api.md` | API 层：数据模型、WebSocket、路由 |
| 4 | `03-client.md` | 前端架构：IM 界面、组件、i18n |
| 5 | `04-sandbox.md` | Agent 基础设施：Eino 框架、ChatModel、状态机、Manager 层 |
| 5b | `04b-sandbox-runtime.md` | 沙箱运行时：SandboxManager、RuntimeManager 适配器 |
| 6 | `06-agent.md` | anserAgent：五层记忆、Skill 自改进 |
| 7 | `09-ci-cd.md` | CI/CD：GitHub Actions 工作流 |
| 8 | `08-deployment.md` | 部署：单机/分布式方案 |
| 9 | `10-roadmap.md` | 路线图：L1-L4 交付计划 |
| 11 | `11-backlog.md` | 远期规划（Phase 2） |

## 引用说明

- 所有 API 路由、数据模型、CLI 命令均为目标架构设计，未必全部在仓库中落地
- 参考代码骨架位于 `reference/workflow-backend-skeleton/`
- 前端完整代码示例位于 `reference/frontend-code-examples.md`
- 数据库 DDL 位于 `docs/ddl.sql`
- 各文件中标注的 **Phase 1** 内容属于 L1-L4 路线图范围，**Phase 2** 为远期规划

> 最后更新：2026-05-21
