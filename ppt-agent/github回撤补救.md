# GitHub 回撤补救措施

## 事件概述

2026-05-21 提交 `fdf62fb` 时，因 `.gitignore` 配置不完整，导致以下文件被推送至 GitHub：

| 文件 | 风险 | 处理 |
|------|------|------|
| `ppt-agent.zip` (83MB) | 无敏感内容，但不应在仓库中 | amend 移除 + gitignore |
| `.claude/settings.local.json` | 本地 IDE 配置，无密钥 | git rm --cached + gitignore |
| `frontend/node_modules/nanoid/.claude/settings.local.json` | node_modules 残留 | git rm --cached |

`.env` 文件**未**进入 git 历史，但百度云仍检测到 API Key 泄漏，需轮换密钥。

---

## 已执行的补救步骤

### 1. 补全 `.gitignore`

```gitignore
# Env
.env                    # ← 之前缺失，已补
.env.local
.env.*.local

# Archives
*.zip                   # ← 新增
*.tar.gz
*.7z

# Test output
test/                   # ← 新增

# Build artifacts
*.tsbuildinfo           # ← 新增

# Claude local config
.claude/                # ← 新增
```

### 2. 从仓库移除敏感文件

```bash
# 移除 zip 和本地配置
git rm --cached ppt-agent.zip .claude/settings.local.json
git rm --cached frontend/node_modules/nanoid/.claude/settings.local.json

# 更新 .gitignore
git add .gitignore

# amend 合并到上一个提交
git commit --amend --no-edit

# 推送
git push origin master
```

### 3. 追加清理提交

```bash
# 确认所有敏感文件已移除
git commit -m "chore: 移除敏感文件，补全 gitignore (.env, *.zip, .claude/)"
git push origin master
```

### 4. 验证远程仓库清洁

```bash
# 确认远程无敏感文件
git ls-tree -r origin/master --name-only | grep -E "\.zip|\.env$|settings\.local"
# 输出：空（无敏感文件）
```

---

## 需要你手动执行

### ⚠️ 轮换百度云 API Key

即使 `.env` 未进入 git，曾推送的日志中可能包含请求头信息。安全起见：

1. 登录百度云控制台
2. 进入 API Key 管理页面
3. **撤销旧 Key**，生成新 Key
4. 更新服务器上 `backend/.env` 中的 `ARK_API_KEY` 为新值
5. 重启 ppt-agent 服务

### 检查服务器

```bash
# 确认 .env 不在 web 可访问路径
curl -I http://your-server:8080/.env
# 应该返回 404

# 确认 nginx/apache 没有错误代理 .env 文件
```

### GitHub 安全设置

1. 进入 GitHub 仓库 → Settings → Secrets and variables
2. 确认 Actions secrets 中没有过期密钥
3. 考虑启用 GitHub secret scanning（Settings → Code security）

---

## 预防措施（已落地）

| 措施 | 状态 |
|------|------|
| `.gitignore` 覆盖 `.env` / `*.zip` / `.claude/` | ✅ |
| `commit --amend` 移除敏感文件 | ✅ |
| git history clean | ✅ |
| API Key 轮换 | ⚠️ 待执行 |
| `.env` 模板文件 `.env.example`（不含真实 key） | 建议添加 |

### 建议添加 `.env.example`

```bash
# backend/.env.example（提交到 git，不含真实值）
ARK_API_KEY=your_api_key_here
ARK_MODEL=your_model_name
ARK_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
MYSQL_DSN=root:password@tcp(127.0.0.1:3306)/myapp?charset=utf8mb4
COZELOOP_API_TOKEN=
COZELOOP_WORKSPACE_ID=
PYTHON_BIN=/root/pptx_env/bin/python
AGENT_MODE=planner
```

---

## 时间线

| 时间 | 操作 |
|------|------|
| 2026-05-21 | 提交 `ea4e242` 含 zip/settings |
| 2026-05-21 | amend → `fdf62fb`，移除 zip/settings |
| 2026-05-21 | 追加 cleanup commit → `3c4681c` |
| 2026-05-21 | 验证远程清洁 |
| ⏳ 待执行 | 百度云 Key 轮换 |
