# Frontend Code Examples（03-client.md 代码示例）

> 本文档集中存放 [03-client.md](../docs/plan/03-client.md) 的完整代码示例，按文档章节对应。

---

## 前端框架 — TanStack Query

```tsx
// hooks/use-issues.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'

export function useIssues(projectId: number) {
  return useQuery({
    queryKey: ['issues', projectId],
    queryFn: () => fetch(`/api/projects/${projectId}/issues`).then(r => r.json()),
    staleTime: 30 * 1000,  // 30s 内不重新请求
  })
}

export function useCreateIssue() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateIssueReq) =>
      fetch('/api/issues', { method: 'POST', body: JSON.stringify(data) }).then(r => r.json()),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: ['issues', vars.project_id] })
    },
  })
}
```

## 前端框架 — Zustand

```tsx
// stores/sidebar.ts
import { create } from 'zustand'

export const useSidebar = create<SidebarState>((set) => ({
  isOpen: true,
  toggle: () => set((s) => ({ isOpen: !s.isOpen })),
}))
```

## 前端框架 — React Hook Form + Zod

```tsx
// components/agent-form.tsx
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

const agentSchema = z.object({
  name: z.string().min(1, '名称不能为空').max(64),
  role_label: z.string().max(64),               // 自定义角色标签（如 PM / 前端 / 后端）
  system_prompt: z.string().max(200, '人设 1-2 句话即可，调度行为由 anserAgent 五层记忆驱动'),
  runtime_id: z.number().min(1, '请选择运行时'),
  // runtime_config 由前端根据 runtimes.config_schema 动态生成表单字段
})

type AgentFormData = z.infer<typeof agentSchema>

export function AgentForm() {
  const form = useForm<AgentFormData>({ resolver: zodResolver(agentSchema) })
  // ...
}
```

## 前端框架 — TanStack Table

```tsx
// features/issues/components/issue-table.tsx
const columns: ColumnDef<Issue>[] = [
  { accessorKey: 'title', header: '标题' },
  { accessorKey: 'status', header: '状态' },
  { accessorKey: 'priority', header: '优先级' },
]

<DataTable columns={columns} data={issues} />
```

## 前端框架 — Recharts

```tsx
// features/dashboard/components/issue-stats-chart.tsx
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'

const data = [
  { status: 'backlog', count: 12 },
  { status: 'todo', count: 8 },
  { status: 'in_progress', count: 5 },
  { status: 'in_review', count: 3 },
  { status: 'done', count: 20 },
]

<ResponsiveContainer width="100%" height={300}>
  <BarChart data={data}>
    <CartesianGrid strokeDasharray="3 3" />
    <XAxis dataKey="status" />
    <YAxis />
    <Tooltip />
    <Bar dataKey="count" fill="var(--primary)" radius={[4, 4, 0, 0]} />
  </BarChart>
</ResponsiveContainer>
```

## 前端框架 — lucide-react

```tsx
import { Plus, Trash2, Settings, Users, FolderKanban, Bot, MessageSquare } from 'lucide-react'

<Button><Plus className="mr-2 h-4 w-4" />创建</Button>
<Button variant="destructive"><Trash2 className="h-4 w-4" /></Button>
```

## 前端框架 — next-themes

```tsx
import { useTheme } from 'next-themes'

<Button onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}>
  <Sun className="dark:hidden" />
  <Moon className="hidden dark:block" />
</Button>
```

## 前端框架 — Framer Motion

```tsx
import { motion, AnimatePresence } from 'framer-motion'

<AnimatePresence>
  {isOpen && (
    <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0 }}>
      <Modal />
    </motion.div>
  )}
</AnimatePresence>
```

## 前端框架 — Sonner

```tsx
import { toast } from 'sonner'

toast.promise(createAgent(data), {
  loading: '创建 Agent 中...',
  success: 'Agent 创建成功',
  error: '创建失败',
})
```

## 前端框架 — date-fns

```tsx
import { format, formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

format(new Date(), 'yyyy-MM-dd HH:mm')
formatDistanceToNow(issue.created_at, { locale: zhCN, addSuffix: true }) // "3 小时前"
```

## 环境变量

```bash
# .env.local
NEXT_PUBLIC_API_BASE=http://localhost:8080/api
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
```

## 代码质量工具

```json
{
  "scripts": {
    "lint": "eslint . --ext .ts,.tsx",
    "format": "prettier --write .",
    "type-check": "tsc --noEmit"
  }
}
```

## @anserflow/shared-ui

```tsx
// 引用示例
import { StatusBadge, AgentAvatar, useWebSocket, apiClient } from '@anserflow/shared-ui'
import type { Issue, WSMessage, PaginatedResponse } from '@anserflow/shared-ui'
```

---

## 国际化 — next-intl 配置

### 安装

```bash
npm install next-intl
```

### next.config.ts

```ts
// admin/next.config.ts
import createNextIntlPlugin from 'next-intl/plugin'

const withNextIntl = createNextIntlPlugin('./src/i18n/request.ts')

const nextConfig = {
  basePath: '/admin',
  output: 'export',
  distDir: 'dist',
}

export default withNextIntl(nextConfig)
```

### request.ts（i18n 入口）

```ts
// admin/src/i18n/request.ts
import { getRequestConfig } from 'next-intl/server'

// 静态导出模式：locale 从客户端 localStorage / navigator.language / 用户设置获取
// URL 不含 locale 段，所有页面统一走 /admin/* 路径
export default getRequestConfig(async () => {
  // 静态导出时使用默认 locale，实际切换在客户端完成
  return {
    locale: 'zh-CN',
    messages: (await import(`../../messages/zh-CN.json`)).default,
  }
})
```

### Locale 检测与切换

```ts
// admin/src/lib/locale.ts
export function detectLocale(): string {
  // 优先级：已登录用户设置 > localStorage > 浏览器语言 > 默认 zh-CN
  const stored = localStorage.getItem('anserflow-locale')
  if (stored) return stored

  const browserLang = navigator.language
  if (browserLang.startsWith('zh')) return 'zh-CN'
  if (browserLang.startsWith('en')) return 'en-US'

  return 'zh-CN'
}

export function setLocale(locale: string) {
  localStorage.setItem('anserflow-locale', locale)
  // 如果已登录，同步到后端 users.locale
  // window.location.reload() 或触发 next-intl 的 setLocale
}
```

### LanguageSwitcher 组件

```tsx
// 语言切换组件（无 URL 变化，纯客户端切换）
import { useRouter } from 'next/navigation'

export function LanguageSwitcher() {
  const switchTo = (locale: string) => {
    setLocale(locale)
    window.location.reload() // 重新加载页面以切换翻译包
  }

  return (
    <select onChange={(e) => switchTo(e.target.value)} value={detectLocale()}>
      <option value="zh-CN">中文</option>
      <option value="en-US">English</option>
    </select>
  )
}
```

### next-intl 组件使用

```tsx
// admin/src/app/dashboard/page.tsx
import { useTranslations } from 'next-intl'

export default function DashboardPage() {
  const t = useTranslations('Nav')
  const tIssue = useTranslations('Issue')

  return (
    <div>
      <h1>{t('dashboard')}</h1>           {/* "仪表盘" 或 "Dashboard" */}
      <span>{tIssue('noResults')}</span>  {/* "暂无 Issue" 或 "No Issues" */}
    </div>
  )
}
```

### 日期/数字本地化

```tsx
import { useFormatter } from 'next-intl'

const format = useFormatter()

// 日期
format.dateTime(issue.createdAt, {
  year: 'numeric', month: 'long', day: 'numeric'
})
// zh-CN → "2026年5月13日"
// en-US → "May 13, 2026"

// 相对时间
format.relativeTime(issue.createdAt)
// zh-CN → "3小时前"
// en-US → "3 hours ago"
```

### 翻译文件 — zh-CN

```json
// admin/messages/zh-CN.json
{
  "Nav": {
    "dashboard": "仪表盘",
    "agents": "智能体",
    "projects": "项目",
    "skills": "技能",
    "settings": "设置"
  },
  "Issue": {
    "title": "Issue 标题",
    "status": "状态",
    "priority": "优先级",
    "assignee": "负责人",
    "create": "创建 Issue",
    "noResults": "暂无 Issue"
  },
  "Common": {
    "save": "保存",
    "cancel": "取消",
    "delete": "删除",
    "confirm": "确认",
    "loading": "加载中...",
    "error": "出错了"
  }
}
```

### 翻译文件 — en-US

```json
// admin/messages/en-US.json
{
  "Nav": {
    "dashboard": "Dashboard",
    "agents": "Agents",
    "projects": "Projects",
    "skills": "Skills",
    "settings": "Settings"
  },
  "Issue": {
    "title": "Issue Title",
    "status": "Status",
    "priority": "Priority",
    "assignee": "Assignee",
    "create": "Create Issue",
    "noResults": "No Issues"
  },
  "Common": {
    "save": "Save",
    "cancel": "Cancel",
    "delete": "Delete",
    "confirm": "Confirm",
    "loading": "Loading...",
    "error": "Something went wrong"
  }
}
```

### Go 后端翻译管理

```bash
# Go 后端翻译管理
goi18n extract           # 从 Go 源码提取待翻译消息 → translate.zh-CN.json
goi18n merge active.*.json translate.*.json  # 合并新增 key
```

---

## 客户端 — IssueDetail 组件

```tsx
// features/issues/components/issue-detail.tsx

export function IssueDetail({ issueId }: { issueId: number }) {
  // 加载 Issue 元信息
  const { data: issue } = useQuery({
    queryKey: ['issue', issueId],
    queryFn: () => fetch(`/api/issues/${issueId}`).then(r => r.json()),
  })

  return (
    <div className="flex flex-col h-full">
      {/* 头部：标题 + 状态 + 优先级 + 负责人 */}
      <IssueHeader issue={issue} />

      {/* 工具栏：编辑/暂停/停止 + 日志筛选 */}
      <IssueToolbar issueId={issueId} status={issue?.status} />

      {/* 时间线：核心区域，占满剩余空间 */}
      <IssueTimeline issueId={issueId} />

      {/* 底部：追加提示词输入框（in_progress / paused / todo 状态显示） */}
      <PromptInput issueId={issueId} status={issue?.status} />

      {/* Token 消耗统计（折叠显示） */}
      <TokenUsageSummary issueId={issueId} />
    </div>
  )
}
```

## 客户端 — IssueTimeline 组件

```tsx
// features/issues/components/issue-timeline.tsx

export function IssueTimeline({ issueId }: { issueId: number }) {
  const bottomRef = useRef<HTMLDivElement>(null)
  const [filter, setFilter] = useState<string>('all')

  // ① 历史数据加载（分页）
  const { data: timeline } = useQuery({
    queryKey: ['issue-timeline', issueId],
    queryFn: () =>
      fetch(`/api/issues/${issueId}/timeline?page=1&size=100`).then(r => r.json()),
  })

  // ② WebSocket 实时追加
  const queryClient = useQueryClient()
  useEffect(() => {
    const unsub = ws.subscribe(`issue:${issueId}`, (msg) => {
      if (msg.type === 'agent_log' || msg.type === 'status_change') {
        queryClient.setQueryData(
          ['issue-timeline', issueId],
          (old) => [...(old || []), {
            id: Date.now(),           // 临时 ID，刷新后由服务端 ID 替换
            source: 'agent',          // 或 msg.source
            event_type: msg.type,
            content: msg.text || msg.hint,
            created_at: new Date(msg.ts * 1000).toISOString(),
          }]
        )
      }
    })
    return () => unsub()
  }, [issueId])

  // ③ 自动滚动到底部
  useEffect(() => { bottomRef.current?.scrollIntoView({ behavior: 'smooth' }) }, [timeline])

  // ④ 前端筛选
  const filtered = (timeline || []).filter(item => {
    if (filter === 'all') return true
    if (filter === 'agent') return item.source === 'agent'
    if (filter === 'system') return item.source === 'system'
    if (filter === 'human') return item.event_type === 'human_prompt'
    return true
  })

  return (
    <div className="flex-1 overflow-y-auto p-3 space-y-1">
      {filtered.map(item => (
        <TimelineItem key={item.id} item={item} />
      ))}
      <div ref={bottomRef} />
    </div>
  )
}

// 单条时间线
function TimelineItem({ item }: { item: TimelineEvent }) {
  const icon = getActionIcon(item)       // 根据 event_type + action 映射图标
  const style = getEventStyle(item)       // 根据 event_type 映射样式类
  return (
    <div className={`flex items-start gap-2 text-sm py-0.5 ${style}`}>
      <span className="text-xs text-muted-foreground w-12 shrink-0">
        {formatTime(item.created_at)}       {/* "12:05" */}
      </span>
      <span className="text-xs text-muted-foreground w-14 shrink-0">
        {item.source}                         {/* "agent" / "system" / "张三" */}
      </span>
      <span className="shrink-0">{icon}</span>
      <span className="flex-1">{item.content}</span>
    </div>
  )
}
```
