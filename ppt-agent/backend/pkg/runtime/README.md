# Runtime boundaries

运行时相关代码按职责集中在二级目录，包名和对外契约保持不变：

| 目录 | 责任 | 入口 |
| --- | --- | --- |
| `web` | Gin 路由、鉴权、请求解析、SSE/文件响应 | `web.NewServer` |
| `task` | 任务状态、SSE 回放、持久化、会话与交付 | `task.NewTaskManager` |
| `model` | provider 配置、模型工厂、fallback、限流、流清洗与压缩 | `utils.NewFallbackToolCallingChatModel` |

目录迁移只改变 import path，不改变 Go package name、HTTP/SSE/JSON 字段或历史任务文件兼容规则。新增运行时文件应放入对应边界目录，避免再次在 `pkg` 根目录堆叠。
