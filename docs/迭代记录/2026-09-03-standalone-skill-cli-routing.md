# 2026-09-03 独立 PPT Skill 图片 CLI 路由

## 目标

避免通用 Agent 在已经加载 `ppt-deck-planner` 的情况下，因为宿主同时提供图像生成能力而自行生成背景图，跳过已配置的图片搜索与下载工具。

## 交付

- `SKILL.md` 将独立 Agent 在 `visual_policy.mode="required"` 下读取 `references/unsplash-cli.md` 从按需建议提升为必经步骤。
- 增加明确的图片工具路由：独立 Agent 必须先使用随 skill 提供的 `unsplash` CLI 搜索、下载并回填素材；未认证、网络失败或无结果时不得静默使用 `imagegen`、虚构本地路径或伪造来源。
- 只有用户明确要求 AI 生成的插画、概念图或艺术图时，才允许图像生成；该素材不能替代每页 required 背景。其他宿主工具只有实际搜索并下载 Unsplash 图片、同时写回路径、来源与署名时才可等价替代 CLI。

## 验证与上线

- 本地：`node scripts/unsplash.mjs --help` 正常输出 `auth`、`fetch` 命令；`python -m unittest discover -s skills/ppt-deck-planner/tests -v` 通过 39 项；`git diff --check` 通过。
- `skill-creator` 的 `quick_validate.py` 未能运行，原因是当前默认 Python 环境缺少 `PyYAML`；这不是 skill 内容或生成器失败，且生成器回归已通过。
- 运行时 skill 已同步到 `remote-dev:/ppt/ppt-agent/skills/ppt-deck-planner` 与 `/ppt/skills/ppt-deck-planner`，两个入口及 CLI 参考文件的 SHA256 与本地一致。
- 服务从 `/ppt/ppt-agent/backend` 以 `/ppt/ppt-agent-linux -mode web -addr :8080` 重启，进程 PID `1932266`，监听 `:8080`；内网和公网 `http://124.220.22.162:8080/api/health` 均返回 200。
- 远端 `python3` 严格 DeckSpec 预检与最小整套渲染通过；临时 `/tmp/ppt-skill-cli-routing-smoke.pptx` 已删除。远端运行环境不安装 Node.js，CLI 可发现性已在本地 Node 22.13.1 环境验证；线上 Web 服务不依赖 Node.js。
