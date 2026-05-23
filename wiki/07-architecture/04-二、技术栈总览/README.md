> ???`docs/plan/07-architecture.md` ? 67 ?
> ???[???](../../README.md) -> [AnserFlow — 系统架构概述](../README.md) -> 二、技术栈总览
> ???[???](../03-目标架构与参考骨架差异清单/README.md) ? [???](../05-参考代码映射.md)
> ?????[??????](../README.md) ? [目标架构与参考骨架差异清单](../03-目标架构与参考骨架差异清单/README.md) ? [参考代码映射](../05-参考代码映射.md)

## 二、技术栈总览

```
┌──────────────────────────────────────────┐
│ 前端       Next.js 14 SPA (static export)│
│           shadcn/ui + Tailwind CSS        │
│           TanStack Query (数据请求)       │
│           Zustand (客户端状态)            │
│           React Hook Form + Zod (表单)    │
│           next-intl (国际化)              │
├──────────────────────────────────────────┤
│ 后端框架   Gin                           │
│ ORM        GORM                          │
│ 数据库     MySQL 8.0+                    │
│ CLI        Cobra                         │
│ 静态嵌入   embed (Go 1.16+)              │
│ 配置       Viper                         │
│ 日志       Zap                           │
│ 校验       go-playground/validator       │
│ 权限       Casbin (RBAC)                 │
├──────────────────────────────────────────┤
│ 缓存/广播  Redis                         │
│ 实时通信   Gorilla WebSocket             │
│           + Redis Pub/Sub (分布式)        │
│ 任务队列   Asynq (基于 Redis)            │
├──────────────────────────────────────────┤
│ AI Agent   Eino (字节跳动 CloudWeGo)     │
│            + 自研业务封装层               │
│ 沙箱       Docker SDK for Go            │
├──────────────────────────────────────────┤
│ Skills     手动编写 + ZIP 导入           │
│ 认证       JWT + OAuth2 (GitHub)         │
│ 邮件       gomail (SMTP)                 │
│ 邀请       分享链接 + 邮箱               │
│ 国际化     next-intl (前端)               │
│           go-i18n (后端)                  │
│ API文档    Swagger (swaggo/swag)         │
│ 跨域       gin-contrib/cors              │
└──────────────────────────────────────────┘
```

> **anserAgent**：AnserFlow 自研 AI Agent 系统，五层记忆（L0-L4）驱动，集成 Eino 编排框架，支持 Skill 引擎与自改进。详见 [06-agent.md](../../06-agent/README.md)。

## ?????

- [选型理由](01-选型理由.md)
- [系统功能模块总览](02-系统功能模块总览.md)
