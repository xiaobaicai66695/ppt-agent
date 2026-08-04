# 2026-08-04 PPT 任务终态校验修复

## 问题

任务管理器在 Agent 流无错误返回后立即写入 `completed`，之后才检查 `tasks.json` 和页面文件。即使校验结果为 `0/23`、缺失 23 个文件，也只记录 warning，不会撤销完成状态；RuntimeMeta 还会提前记录 `phase=complete`，导致前端展示“任务完成”。

本次事项登记为 `PPT-STATUS-001`，按 direct 路线修复。

## 已完成

- 将任务终态判定移动到 manifest reconcile 和交付校验之后。
- 只有任务清单非空、所有页面状态完成、所有声明文件存在时才允许 `completed`。
- RuntimeMeta 增加终态守卫：`done_slides < total_slides` 或 `missing_files > 0` 时拒绝完成。
- 文件/manifest 与 RuntimeMeta 形成双门禁，任一显示未交付都转为 `failed` 并给出 `已交付 X/Y 页、缺少 N 个页面文件`。
- `phase=complete` 仅在双门禁通过后记录；失败与取消分别记录对应 phase。
- 失败终态不再被 Agent 首句“收到，开始准备”覆盖，前端会收到真实未交付原因。
- 空 tasks.json 也视为无效交付，不能形成 0/0 completed。

## 验证与部署

- 新增 0/23 manifest、RuntimeMeta 0/23、完整 23/23、空 manifest 四类回归测试。
- `go test ./pkg/task ./pkg/agent/deep ./pkg/web` 通过。
- `go build ./...` 通过。
- 远端聚焦测试与 Linux 构建通过。
- 已部署到 `/ppt/ppt-agent`，服务 PID `23035`，`/api/health` 返回 200，日志确认 MySQL 与模型初始化成功。

## 关联内容

- `docs/issues/todo.md`
- `ppt-agent/backend/pkg/task/manager.go`
- `ppt-agent/backend/pkg/task/manager_delivery_test.go`
- `ppt-agent/backend/pkg/agent/deep/types.go`
- `ppt-agent/backend/pkg/agent/deep/types_test.go`
