# Runtime boundaries

运行时相关代码按职责集中在三个主线；Web 主线进一步拆为 router、service、model
三个边界，避免路由表、业务操作和 JSON 结构体继续混在一起：

| 目录 | 责任 | 入口 |
| --- | --- | --- |
| `web` | Server 装配与兼容性 handler | `web.NewServer` |
| `web/router` | Gin 路由表、分组和 middleware 挂载 | `router.Register` |
| `web/service` | 任务生命周期应用服务 | `service.NewTaskService` |
| `web/model` | Web 请求/响应和路由 DTO | model structs |
| `task` | 任务状态、SSE 回放、持久化、会话与交付 | `task.NewTaskManager` |
| `model` | provider 配置、模型工厂、fallback、限流、流清洗与压缩 | `utils.NewFallbackToolCallingChatModel` |

文件拆分不改变顶层 `web.NewServer`、HTTP/SSE/JSON 字段或历史任务文件兼容规则。
新增 Web 代码应遵守 `router → service/model`、`service → model/task` 的单向依赖，
不要把 Gin handler 或数据库模型放回 `web/model`。
