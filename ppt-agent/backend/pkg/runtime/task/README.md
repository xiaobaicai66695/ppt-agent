# Task runtime

任务代码按状态机职责导航。当前文件保持在同一 Go package 以保留 `TaskState` 未导出锁边界；子目录作为稳定的迁移索引。

- `state/`：状态、锁和生命周期转换
- `sse/`：事件缓存、监听器、游标回放和 chunk 规范化
- `persistence/`：任务文件、数据库投影和恢复
- `conversation/`：会话流与继续消息
- `delivery/`：文件就绪、交付元数据和反馈同步
