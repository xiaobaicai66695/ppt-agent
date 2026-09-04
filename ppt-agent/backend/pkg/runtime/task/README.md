# Task runtime

任务代码按状态机职责导航。所有实现文件都在当前目录并属于同一 Go package，以保留 `TaskState` 未导出锁边界。

- `manager.go`、`state_boundary.go`：状态、锁和生命周期转换
- `sse_stream.go`：事件缓存、监听器、游标回放和 chunk 规范化
- `persistence_boundary.go`、`plan_helpers.go`：任务文件、数据库投影和恢复
- `manager.go`：会话流与继续消息
- `delivery.go`、`delivery_metadata.go`、`feedback.go`：文件就绪、交付元数据和反馈同步
