# 2026-09-02 Benchmark 扩充与交付反馈上线记录

## 目标

- 将 `router`、`planner`、`reviewer`、`fixer` 的 `test` 与 `validation` benchmark 均扩充到每类 10 条，并增加离线覆盖、唯一 ID 和 test/validation 隔离检查。
- PPT 完成后，在 Dashboard 左下角提供 1–5 分评分和可选建议；反馈仅允许任务 owner 对已完成任务提交，并支持更新。
- 同步包含本轮已提交的审批控制、SSE 回放游标修复和图文正文阅读微调。

## 本地验证

- `go test ./...`
- `go test ./cmd/pptbench ./pkg/task ./pkg/web`
- `go build ./...`（Linux 交付二进制已生成）
- 前端 `npm test`：41 项通过；`npm run build` 通过
- Python 生成器：38 项测试通过，逐文件 `py_compile` 通过
- `openspec validate expand-benchmark-and-collect-deck-feedback --type change --strict`
- `git diff --check`

Benchmark 数量检查结果：`test` 与 `validation` 中四类均为 10 条；请求不重复，ID 非空且全局唯一。

## 线上发布

- 目标：`ssh remote-dev`，运行目录 `/ppt/ppt-agent`。
- 后端：`/ppt/ppt-agent-linux`，工作目录 `/ppt/ppt-agent/backend`，启动参数 `-mode web -addr :8080`。
- 前端：`/ppt/ppt-agent/frontend/dist`；运行时 skill：`/ppt/ppt-agent/skills/ppt-deck-planner`。
- 旧版本备份：`/ppt/ppt-agent/deploy-backups/20260902-feature-release`。
- 发布后最终进程 PID `1688027`，`:8080` 正常监听；启动日志包含 `mysql_connected`，无立即失败。

## 线上冒烟

- `GET /api/health`：200，`{"status":"ok"}`。
- `GET /health/ready`：200，MySQL、LibreOffice、Python 均为 `ok`。
- `GET /api/templates/layouts`：200。
- `GET /` 与 `GET /dashboard`：均为 200，返回当前前端入口。
- 当前 Dashboard bundle 请求为 200，包含 `feedback`、`评分`、`建议` 文案。
- 未登录 `PUT /api/tasks/does-not-exist/feedback`：401（确认反馈路由存在且受鉴权保护）。
- 未创建真实任务、未写入测试反馈；因没有可安全使用的线上测试账号，本轮未执行 owner 登录后的真实提交/更新冒烟，功能由本地 API/任务测试覆盖。

## 清理与回滚

- 冒烟未产生任务、PPT 或缩略图数据；发布临时 staging 文件已在复测后清理。
- 回滚时可使用备份目录中的旧二进制、前端 dist 和 skill；备份目录保留，不包含凭据记录。
