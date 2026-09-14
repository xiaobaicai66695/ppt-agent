# 2026-09-14 历史会话查询索引上线记录

## 事项

- 提交：`34ea5bc`（`perf: index historical conversation queries`）
- 目的：降低历史会话、可见执行轨迹和任务侧栏按时间读取时的排序成本。

## 变更

- `conversation_messages` 新增 `idx_conversation_messages_task_timestamp_id (task_id, timestamp, id)`。
- `conversation_trace_events` 新增 `idx_conversation_trace_task_timestamp_id (task_id, timestamp, id)`。
- `task_records` 新增 `idx_task_records_user_created (user_id, created_at)`。

旧单列索引暂不删除，待持续收集线上执行计划和负载证据后再单独治理。

## 验证与上线

- 本地：`go test ./pkg/db/... ./pkg/runtime/web`、`go build ./...` 均通过。
- Linux 交付物：`GOOS=linux GOARCH=amd64 go build` 成功，SHA-256 为 `0246DF15E0E666B58C204E97E522DD07C994BDAEDA3D721F0BC77E79D887F0EC`。
- 部署目标：`remote-dev:/ppt/ppt-agent`；保留远端 `backend/.env` 与 `weboutput`，替换服务端源码、skills 和 Linux 二进制。
- 2026-09-14 17:49（Asia/Shanghai）新进程 PID `1716477` 启动，端口 `8080` 正常监听；本机和公网 `/api/health` 均返回 200。
- 线上 `EXPLAIN`：会话消息和轨迹查询分别命中新的 `task_id,timestamp,id` 索引；任务列表命中 `user_id,created_at`，显示 `Backward index scan; Using index`。

## 风险与后续

- 索引仅降低查询和排序成本；极长历史仍可能受完整轨迹 payload 的序列化、传输和前端渲染影响。若监控仍显示打开会话卡顿，再单独实施时间线分页与延迟加载。
