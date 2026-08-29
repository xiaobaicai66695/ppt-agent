# 2026-08-29 持久任务会话与 Skill 大版本发布

## 发布范围

- 所有首次输入先创建用户归属的 `conversation` 任务，`/api/messages` 始终返回并复用 `task_id`。
- 任务会话持久化并向路由器提供有界上下文；“介绍青甘大环线旅游项目”后的“你决定主题和风格吧”会保留原主题，路由为 `create/prepare_create`。
- `POST /api/tasks/:id/start` 将会话任务原地提升为生成任务，不再新建任务 ID。
- Dashboard/Home/Sidebar 显示并保留“对话中”阶段；前端后续消息和开始生成均使用同一个 ID。
- 收录路由上下文 benchmark，以及 Planner/Reviewer/Fixer、视觉素材物化与校验、`ppt-deck-planner` Skill 的本次工作树改动。

## 本地验证

- `go test ./pkg/agent/deck ./pkg/task ./pkg/web ./cmd/pptbench` 通过。
- `go build ./...` 和 Linux `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -trimpath` 通过。
- 前端 `npm test -- --run` 39 个测试与 `npm run build` 均通过。
- 使用工作区 Python 运行 `python -m unittest discover -s skills/ppt-deck-planner/tests`，30 个测试通过；组件渲染测试会生成并检查 PPTX 媒体包。
- `validate_visual_assets.py --help` 通过；路由上下文 benchmark 的 model + judge 结果为 `5.0/5`。

## 上线与冒烟

- 目标：`remote-dev:/ppt/ppt-agent`，后端二进制为 `/ppt/ppt-agent-linux`；同时发布前端 `dist` 和 Skill。
- 2026-08-29 15:55（Asia/Shanghai）新进程 PID 为 `338057`，二进制 SHA-256：`5b3cc45c1ec1d17527f4f0b9dbba5785b2e5758e62f47ecc95ed5a2a7ecfa61a`。
- 健康检查 `/api/health` 返回 200，启动日志包含 `mysql_connected`、`deck_planner_skill_ready dir=/ppt/skills`。
- Skill 同步到 `/ppt/skills/ppt-deck-planner` 与 `/ppt/ppt-agent/skills/ppt-deck-planner`，避免 Planner 的运行时 Skill 根目录与生成器目录脱节。
- 受保护 API 冒烟：首条青甘大环线描述获得任务 ID；“你决定主题和风格吧”返回同一 ID、`intent=create`、`action=prepare_create`；会话记录存在多条 turn，`/api/tasks/:id/start` 返回 202。冒烟任务随后取消并删除，远端发布临时包已清理。

## 已知边界

- 本次任务启动 smoke 只验证同 ID 交接，不等待完整模型渲染，以避免额外的长时模型费用；完整生成仍由现有 Planner -> Reviewer -> Renderer 链路负责。
