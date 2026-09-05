# 2026-09-05 PPT 视觉契约与渲染修复上线记录

## 事项

- ID：`PPT-RENDER-20260905`
- 计划时间：2026-09-05
- 完成时间：2026-09-05
- 范围：修复 Go Planner/Reviewer 与 `ppt-planner` Python 生成器在组件容量、`visual_policy` 和前景图片物化上的契约漂移，防止单页渲染前的全量清单校验阻断整套 PPT。

## 实现与提交

- `82bc02d fix: align PPT visual validation contracts`
  - Go 侧页面组件上限与 `skills/ppt-planner/templates/component_contracts.json` 保持一致：`timeline=6`、`kpi_dashboard=4`，并同步其他受影响页面类型。
  - `visual_policy.mode=required` 自动补齐 `required_roles:["background"]` 与覆盖全部页面的 `min_image_pages`；工具 Schema 显式暴露相应字段。
  - 前景 `image` 组件必须提供 `local_path` 或 `asset_query`；素材服务未能物化前景图片时在素材阶段失败，不再写入残缺清单后才由 Python 生成器拒绝。
- `b05c1b4`、`76b1be4`：在 `AGENTS.md` 固化部署规程：每次交付前清空本次服务端代码目录并重新上传，保留 `backend/.env` 与 `weboutput`。

## 本地验证

- `go test ./pkg/agent/ppt`：通过。
- `python -m unittest skills/ppt-planner/tests/test_validate_visual_assets.py`：4 项通过。
- `GOOS=linux GOARCH=amd64 go build -o <temp>/ppt-agent-linux .`：通过，生成 Linux AMD64 交付物。

## 部署与线上冒烟

- 目标：`remote-dev:/ppt/ppt-agent`，服务端口 `:8080`。
- 部署前停止旧进程；清空 `/ppt/ppt-agent/backend` 的全部代码内容及 `/ppt/ppt-agent/skills`，仅保留运行配置 `backend/.env` 和用户任务目录 `weboutput`；随后上传当前后端、skill 与 Linux 二进制。
- 新进程：PID `2947356`，工作目录 `/ppt/ppt-agent/backend`，启动日志确认 `ppt_planner_skill_ready`、MySQL 连接、单写锁、Redis 初始化和 `server_starting`。
- 冒烟：本机 `/api/health` 返回 200，`/api/templates/layouts` 返回 200，首页返回 200；公网 `http://124.220.22.162:8080/api/health` 返回 `{"status":"ok"}`。

未创建模型生成任务，避免为发布冒烟产生额外模型消耗或污染用户任务列表；渲染契约由上述 Go/Python 聚焦测试覆盖。
