> 来源：`docs/plan/04b-sandbox-runtime.md` 第 325 行
> 位置：[总目录](../../README.md) -> [AnserFlow - 沙箱执行运行时](../README.md) -> [三、扩展新运行时指南](README.md) -> 步骤 3: 更新 Dockerfile
> 相邻：[上一篇](02-步骤-2-替换初始化.md) · 下一篇：无
> 相关主题：[返回上级章节](README.md) · [返回文档入口](../README.md)

### 步骤 3: 更新 Dockerfile

```dockerfile
# 添加新运行时的安装步骤
RUN curl -fsSL https://claude.ai/install.sh | sh
```

Worker 代码无需修改，完全通过接口隔离。
