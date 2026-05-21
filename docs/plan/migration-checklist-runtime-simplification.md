# 运行时精简迁移检查清单

## ✅ 已完成

### 代码层面
- [x] 创建 `internal/runtime/adapter.go`（接口定义）
- [x] 创建 `internal/runtime/anseragent.go`（AnserAgent 实现）
- [x] 创建 `internal/runtime/manager.go`（简化管理器）
- [x] 删除 `internal/runtime/registry.go`
- [x] 删除 `internal/runtime/opencode_adapter.go`
- [x] 删除 `internal/runtime/opencode_parser.go`
- [x] 删除 `internal/runtime/hermes_adapter.go`
- [x] 删除 `internal/runtime/hermes_parser.go`

### 文档层面
- [x] 重写 `docs/plan/04b-sandbox-runtime.md`（精简架构说明）
- [x] 创建 `docs/plan/migration-notes-runtime-simplification.md`（迁移说明）

---

## 🔄 待完成

### 文档更新

- [ ] 更新 `docs/plan/04-sandbox.md`
  - [ ] 第 131 行：`由运行时（anserflow agent）在 Docker 沙箱完成`
  - [ ] 第 177 行：删除 opencode 配置示例，替换为 anserAgent
  - [ ] 第 205 行：`下拉选择运行时（anserAgent）`
  - [ ] 第 217 行：`读取 agents.runtime_id → 确定运行时（anserAgent）`
  - [ ] 第 221 行：`与沙箱内 anserAgent 建立双向流连接`
  - [ ] 第 255 行：`Runtime 默认（anserAgent）`
  - [ ] 第 927 行：`Source string // "agent" | "anseragent"`
  - [ ] 第 1091-1103 行：删除 `anserAgent Tool 与 opencode Tool 对比` 表格
  - [ ] 第 1219 行：`Eino 在将人工提示词注入 anserAgent 之前`
  - [ ] 第 1235 行：`沙箱内的代码生成完全由 anserAgent 完成`
  - [ ] 第 1247 行：`| **anserAgent 执行** | Docker 沙箱内 |`
  - [ ] 第 1261 行：`source="anseragent" → anseragent_prompt_tokens`
  - [ ] 第 1275 行：`#### anserAgent 执行阶段 — 双通道采集`
  - [ ] 第 1395 行：`│   ├── anserflow (Go 二进制) / git / bash`
  - [ ] 第 1424-1426 行：`docker exec anserflow agent run`
  - [ ] 第 1533-1539 行：Step 5 执行命令更新

- [ ] 更新 `docs/plan/06-agent.md`
  - [ ] 第 200 行：`- 运行时：anserAgent（自研）`
  - [ ] 第 696 行：`└── RuntimeAdapter ──► anserAgent 沙箱`
  - [ ] 第 703 行：`| **RuntimeAdapter** | 执行模式的行动层（anserAgent） |`

- [ ] 更新 `docs/plan/02-api.md`
  - [ ] 第 355 行：`注入 anserAgent 配置 + Agent 人设`
  - [ ] 第 357 行：`anserAgent run 执行编码`
  - [ ] 第 358 行：`anserAgent 检查结果`
  - [ ] 第 402 行：`worktree 保留，anserAgent 进程终止`
  - [ ] 第 403 行：`| anserAgent 成功 |`
  - [ ] 第 404 行：`| anserAgent 失败 |`
  - [ ] 第 1696 行：`注入到 anserAgent 执行提示词`
  - [ ] 第 1798 行：`→ 注入配置 + 关联Issue上下文 → anserAgent run`
  - [ ] 第 1822 行：`④ anserAgent 自检查 → PR`

- [ ] 更新参考代码示例
  - [ ] `reference/backend-code-examples.md` 第 250 行
  - [ ] `reference/sandbox-code-examples.md` 相关示例

### Docker 镜像更新

- [ ] 更新 `docker/sandbox/Dockerfile`
  - [ ] 移除 Node.js、npm、Python 安装
  - [ ] 添加 anserflow 二进制复制
  - [ ] 优化镜像大小（目标 <100MB）

- [ ] 更新 `docker/sandbox/entrypoint.sh`
  - [ ] 替换 `opencode run` 为 `anserflow agent run`
  - [ ] 更新配置注入逻辑

### 数据库迁移

- [ ] 执行 SQL 迁移脚本

```sql
-- 1. 废弃旧的运行时
UPDATE runtimes 
SET is_active = 0, deprecated_at = NOW() 
WHERE name IN ('opencode', 'hermes');

-- 2. 插入 anserAgent 运行时
INSERT INTO runtimes (
    name, 
    display_name, 
    config_schema, 
    execute_template,
    home_dir,
    config_path,
    skills_mount_path,
    session_path,
    is_builtin,
    is_active,
    created_at
) VALUES (
    'anseragent',
    'AnserAgent (内置)',
    '{
        "provider": {
            "type": "string",
            "enum": ["openai", "anthropic", "deepseek"],
            "required": true
        },
        "model": {"type": "string", "required": true},
        "max_iterations": {"type": "number", "default": 20},
        "thinking": {"type": "boolean", "default": true}
    }',
    '/usr/local/bin/anserflow agent run --workdir {{.workdir}} --config {{.config_path}} --format json',
    '/home/sandbox/.anseragent',
    '/home/sandbox/.anseragent/config.yaml',
    '/home/sandbox/.anseragent/skills',
    '/home/sandbox/.anseragent/sessions/*.jsonl',
    1,
    1,
    NOW()
);

-- 3. 更新 agents 表的 runtime_id 引用
UPDATE agents 
SET runtime_id = (SELECT id FROM runtimes WHERE name = 'anseragent')
WHERE runtime_id IN (SELECT id FROM runtimes WHERE name IN ('opencode', 'hermes'));
```

### Worker 代码更新

- [ ] 更新 `internal/worker/executor.go`
  - [ ] 使用 `runtime.NewRuntimeManager()` 替代旧的 Registry
  - [ ] 更新执行命令构建逻辑
  - [ ] 更新环境变量注入逻辑

- [ ] 更新 `internal/worker/token_tracker.go`
  - [ ] 将 `Source` 字段从 `"opencode"` 改为 `"anseragent"`

### 测试验证

- [ ] 单元测试
  - [ ] `go test ./internal/runtime/...`（适配器接口测试）
  - [ ] `go test ./internal/worker/...`（Worker 执行测试）

- [ ] 集成测试
  - [ ] 构建 Docker 镜像：`docker build -t anserflow/sandbox:latest -f docker/sandbox/Dockerfile .`
  - [ ] 验证二进制：`docker run --rm anserflow/sandbox:latest anserflow agent --version`
  - [ ] 验证配置注入：启动容器检查配置文件
  - [ ] 验证 stdout 解析：执行测试任务检查日志

- [ ] Token 追踪验证
  - [ ] 实时通道：检查 stdout JSON 解析
  - [ ] 事后通道：检查 session 文件解析
  - [ ] 双通道去重：验证 `max(实时, 事后)` 逻辑

### 文档验证

- [ ] 检查所有文档中的 opencode/hermes 引用
  ```bash
  grep -r "opencode\|hermes" docs/ reference/ --include="*.md"
  ```
- [ ] 确认所有引用已更新为 anserAgent/anserflow agent
- [ ] 验证架构图和流程图一致性
- [ ] 检查代码示例与实际实现一致

---

## 📊 进度统计

| 类别 | 已完成 | 待完成 | 总计 |
|------|--------|--------|------|
| **代码文件** | 7 | 3 | 10 |
| **文档文件** | 2 | 6 | 8 |
| **测试用例** | 0 | 3 | 3 |
| **总进度** | **9/14 (64%)** | **5/14 (36%)** | **14** |

---

## 🎯 预期收益

### 代码简化
- 删除 5 个文件（~800 行代码）
- 新增 3 个文件（~300 行代码）
- **净减少 ~500 行（-62%）**

### 镜像优化
- 移除 Node.js、npm、Python
- 镜像大小：**400MB → 50MB（-87.5%）**
- 构建时间大幅缩短

### 维护成本
- 运行时实现：3 套 → 1 套
- 文档复杂度：大幅降低
- 调试难度：降低（统一二进制）

### 扩展性保留
- ✅ RuntimeAdapter 接口保留
- ✅ OutputParser 接口保留
- ✅ 未来加新运行时只需实现接口

---

## 📝 注意事项

1. **向后兼容**
   - 确保 `runtimes` 表迁移不影响现有 Agent
   - Agent 执行时自动使用新的 anserAgent 运行时

2. **环境变量**
   - anserAgent 使用 `ANSERAGENT_API_KEY`
   - 同时保留 `OPENAI_API_KEY` 等 provider 特定变量

3. **stdout 格式**
   - 确保 anserAgent 输出 JSON Lines 格式
   - Parser 能正确处理非 JSON 行（降级为日志）

4. **Session 文件**
   - 确认 anserAgent 会话文件路径
   - 验证事后 Token 汇总逻辑

5. **Docker 权限**
   - 容器内用户 `sandbox` (uid 1000)
   - 确保有执行 `/usr/local/bin/anserflow` 的权限

---

## 🚀 部署步骤

### Phase 1: 代码合并
```bash
git add internal/runtime/
git commit -m "refactor: 精简运行时架构，保留接口抽象"
```

### Phase 2: 文档更新
```bash
git add docs/plan/
git commit -m "docs: 更新运行时架构文档"
```

### Phase 3: 数据库迁移
```bash
# 在测试环境执行 SQL
mysql -u root -p anserflow < migrations/runtime_simplification.sql
```

### Phase 4: 构建镜像
```bash
docker build -t anserflow/sandbox:v2.0 -f docker/sandbox/Dockerfile .
docker push anserflow/sandbox:v2.0
```

### Phase 5: 部署验证
```bash
# 部署到测试环境
# 创建测试 Issue
# 验证执行流程
# 检查 Token 追踪
```

### Phase 6: 生产部署
```bash
# 灰度发布
# 监控错误率
# 逐步全量
```

---

## 📞 联系方式

如有问题，请联系架构团队或查看：
- `docs/plan/migration-notes-runtime-simplification.md` — 详细迁移说明
- `docs/plan/04b-sandbox-runtime.md` — 最新架构文档
