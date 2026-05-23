> 来源：`docs/plan/04-sandbox.md` 第 727 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱 / Agent 基础设施](../README.md) -> [一、AI Agent 框架：Eino + 自研封装](README.md) -> GitManager — Git 管理器
> 相邻：[上一篇](09-NotificationChannelManager-—-通知渠道管理器.md) · [下一篇](11-TokenManager-—-Token-配额管理器.md)
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### GitManager — Git 管理器

统一管理 Git 平台 API 和仓库操作，分为两个子接口：

| 子接口 | 职责 | 运行位置 | 平台相关性 |

|--------|------|----------|-----------|

| **GitPlatform** | 平台 REST API（Issue/PR/Repo） | Go 后端 Service 层 | 平台相关 |

| **GitOps** | 仓库操作（Clone/Fetch/Push/Commit） | Worker 沙箱内 | 平台无关 |

**设计原则**：

- `manager.Platform(platform)` → 返回对应平台 API 实现

- `manager.NewOps(containerID, workdir)` → 返回仓库操作实例（绑定容器）

- 业务代码不感知具体平台和底层 Git 实现

- Phase 2 可替换为 go-git 库实现（详见 [11-backlog.md](../../11-backlog/README.md)），上层无感知

**接口定义**：

```go

// internal/git/manager.go

package git

import (

    "context"

    "fmt"

)

// GitPlatform 平台 REST API 接口（平台相关，运行在 Go 后端 Service 层）

type GitPlatform interface {

    CreateIssue(ctx context.Context, repo, title, body string, labels []string) (issueID string, err error)

    CreatePR(ctx context.Context, repo, title, head, base, body string) (prURL string, err error)

    GetRepoInfo(ctx context.Context, repo string) (*RepoInfo, error)

    ListBranches(ctx context.Context, repo string) ([]string, error)

}

// GitOps 仓库操作接口（平台无关，运行在 Worker 沙箱内）

type GitOps interface {

    IsRepo(ctx context.Context) bool

    Clone(ctx context.Context, repoURL, branch, dest string) error

    FetchAll(ctx context.Context) error

    Checkout(ctx context.Context, branch string) error

    Pull(ctx context.Context, branch string) error

    Commit(ctx context.Context, message string, author Author) (string, error)

    Push(ctx context.Context) error

}

// Author 提交者信息

type Author struct {

    Name  string

    Email string

}

// GitManager Git 管理器（统一入口）

type GitManager struct {

    platforms       map[string]GitPlatform

    defaultPlatform string

}

func NewGitManager() *GitManager {

    return &GitManager{

        platforms:       make(map[string]GitPlatform),

        defaultPlatform: "github",

    }

}

// Register 注册平台 API 实现

func (m *GitManager) Register(platform string, p GitPlatform) {

    m.platforms[platform] = p

}

// Platform 获取平台 API 实现（默认返回 github）

func (m *GitManager) Platform(platform string) (GitPlatform, error) {

    if platform == "" {

        platform = m.defaultPlatform

    }

    p, ok := m.platforms[platform]

    if !ok {

        return nil, fmt.Errorf("unsupported git platform: %s", platform)

    }

    return p, nil

}

// NewOps 为指定容器创建仓库操作实例

func (m *GitManager) NewOps(containerID, workdir string) GitOps {

    return NewContainerGitOps(containerID, workdir)

}

```

**ContainerGitOps — 容器内 Shell 实现**：通过 `docker exec` 映射 GitOps 方法到 Shell 命令：

| GitOps 方法 | 容器内 Shell 命令 |

|------------|------------------|

| `IsRepo` | `test -d {workdir}/.git` |

| `Clone(repoURL,branch,dest)` | `git clone --branch {branch} {repoURL} {dest}` |

| `FetchAll` | `cd {workdir} && git fetch --all` |

| `Checkout(branch)` | `cd {workdir} && git checkout {branch}` |

| `Pull(branch)` | `cd {workdir} && git checkout {branch} && git pull` |

| `Commit(msg,author)` | `cd {workdir} && git add . && git commit -m "{msg}" --author="{name} <{email}>"` |

| `Push` | `cd {workdir} && git push` |

> Phase 2 可选替换为 `GoGitOps`（go-git 库实现，详见 [11-backlog.md](../../11-backlog/README.md)），上层通过 `GitOps` 接口无感知切换。

**初始化**：

```go

gitMgr := git.NewGitManager()

gitMgr.Register("github", &GitHubPlatform{token: cfg.GitHubToken})

// Phase 2: gitMgr.Register("gitea", &GiteaPlatform{...})

// Phase 2: gitMgr.Register("gitlab", &GitLabPlatform{...})

```
