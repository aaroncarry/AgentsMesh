---
name: worktree
description: |
  创建 Git worktree 用于隔离开发新功能或修复 bug。
  自动处理分支创建、worktree 设置、目录切换和开发环境初始化。
user-invocable: true
---

# Git Worktree 创建

创建独立的 worktree 用于并行开发，避免污染主工作目录。

## 使用流程

### 1. 确认参数

需要以下信息：
- **分支名称**: 新功能/修复的分支名（如 `feature/add-login`, `fix/user-auth`）
- **基础分支**: 从哪个分支创建（默认 `main`）
- **worktree 目录**: 放置位置（默认 `../<repo>-<branch>`）

### 2. 创建 Worktree

```bash
# 1. 获取最新代码
git fetch origin

# 2. 创建 worktree 和新分支
git worktree add -b <branch-name> <worktree-path> origin/<base-branch>

# 3. 进入 worktree 目录
cd <worktree-path>

# 4. 验证状态
git status
git log --oneline -3
```

### 3. 初始化开发环境

worktree 创建后，执行开发环境初始化脚本：

```bash
# 进入 deploy/dev 目录
cd deploy/dev

# 一键初始化（生成隔离配置 + 启动 Docker + 数据库迁移 + seed 数据）
./init-worktree.sh
```

脚本会自动：
- 根据 worktree/分支名生成隔离的 `.env` 配置（端口自动偏移，避免冲突）
- 启动所有 Docker 服务（PostgreSQL、Redis、MinIO、Backend、Web、Runner）
- 执行数据库迁移
- 初始化测试账号 seed 数据

### 4. 完成后输出

创建完成后，告知用户：

```
已创建 worktree:
- 路径: /Users/xxx/Works/AIO/AgentMesh-feature-user-auth
- 分支: feature/user-auth (基于 origin/main)

开发环境:
- 访问地址: http://localhost:<port>
- 测试账号: dev@agentmesh.local / devpass123
- Adminer: http://localhost:<adminer-port>
- MinIO: http://localhost:<minio-port>

常用命令:
- 查看日志: cd deploy/dev && docker compose logs -f backend
- 停止环境: cd deploy/dev && docker compose down
- 清理重建: cd deploy/dev && ./init-worktree.sh --clean && ./init-worktree.sh

完成开发后:
- 提交代码: git add . && git commit -m "..."
- 推送分支: git push -u origin feature/user-auth
- 清理 worktree: cd <主仓库> && git worktree remove <worktree-path>
```

## 完整示例

用户说："创建一个 worktree 开发用户认证功能"

执行：
```bash
# 创建 worktree
git fetch origin
git worktree add -b feature/user-auth ../AgentMesh-feature-user-auth origin/main
cd ../AgentMesh-feature-user-auth

# 初始化开发环境
cd deploy/dev
./init-worktree.sh
```

## 注意事项

- 分支名遵循约定：`feature/*`, `fix/*`, `refactor/*`, `docs/*`
- Worktree 目录默认放在仓库同级目录
- 如果分支已存在，使用 `git worktree add <path> <existing-branch>`
- 每个 worktree 的开发环境端口自动隔离，可并行运行多个
- 清理前确保所有更改已提交或推送
- 清理环境：`./init-worktree.sh --clean`
