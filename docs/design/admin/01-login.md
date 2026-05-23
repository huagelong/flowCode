# Admin 登录 / 注册

> 路由：`/admin/login`，`/admin/register`
> 共享页面，详见 [shared/00-auth-pages.md](../shared/00-auth-pages.md)

---

## 页面差异

Admin 登录页与 Client 登录页共享同一组件，仅以下参数不同：

| 属性 | Admin | Client |
|------|-------|--------|
| 登录后跳转 | `/admin/dashboard` | `/client/chat` |
| Logo 文案 | "AnserFlow Admin" | "AnserFlow" |
| 注册入口链接 | `/admin/register` | `/client/register` |

---

## 布局

```
┌───────────────────────────────────────────────────────────┐
│                                                           │
│                    ┌─────────────────┐                    │
│                    │   AnserFlow     │                    │
│                    │   Admin         │                    │
│                    │                 │                    │
│                    │  ┌───────────┐  │                    │
│                    │  │ Email     │  │                    │
│                    │  └───────────┘  │                    │
│                    │  ┌───────────┐  │                    │
│                    │  │ Password  │  │                    │
│                    │  └───────────┘  │                    │
│                    │                 │                    │
│                    │  [  登 录  ]    │                    │
│                    │                 │                    │
│                    │  ─── 或者 ───   │                    │
│                    │                 │                    │
│                    │  [GitHub 登录]  │                    │
│                    │                 │                    │
│                    │  没有账号？注册  │                    │
│                    └─────────────────┘                    │
│                                                           │
└───────────────────────────────────────────────────────────┘
```

- 居中卡片布局，`max-w-sm` (384px)
- 背景色 `bg-background`，卡片 `bg-card shadow-sm rounded-xl`
- 完整设计规范见 [shared/00-auth-pages.md](../shared/00-auth-pages.md)
