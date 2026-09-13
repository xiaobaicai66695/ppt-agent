# 2026-09-13 会话历史执行轨迹持久化

- 提交：`15bdb34 fix: persist conversation execution history`
- 范围：历史 PPT 会话回放。

## 改动

- 新增 `conversation_trace_events`，仅持久化已发送到工作台的安全思考摘要、工具调用/结果和关键进度；不保存模型私有推理。
- `/api/tasks/:id/conversation` 按时间合并会话消息与持久化轨迹，前端恢复可折叠的思考、工具调用/结果卡；对旧任务从 `runtime_event_records` 最佳努力回放工具与阶段卡。
- 删除任务时同步删除轨迹记录。

## 验证与发布

- 本地：`go test ./pkg/runtime/task ./pkg/runtime/web ./pkg/db`、`go build ./...`、`npm test -- --run src/utils/conversationTimeline.test.ts`、`npm run build` 均通过。
- UI 静态审计：Premium strict audit 0 findings；`DESIGN.md` lint 0 errors（保留 5 条既有 orphaned-token warnings）。
- 部署：`remote-dev:/ppt/ppt-agent`，替换后端源码、skills、Linux 二进制和前端 `dist`；保留远端 `backend/.env` 与 `weboutput`。
- 新进程：PID `1404404`，`/api/health` 返回 200，启动日志确认 MySQL 连接和单实例锁正常。
- 线上低成本冒烟：临时检索对话的历史接口返回 `message, thought, tool_call, tool_result`；数据库写入 4 条安全轨迹记录。临时任务、消息、轨迹和工作目录均已清理。

## 遗留边界

- 本次发布前已完成的旧任务没有 `conversation_trace_events`，只能从已有运行事件恢复工具和阶段摘要；无法补回当时未持久化的完整中间文本。
