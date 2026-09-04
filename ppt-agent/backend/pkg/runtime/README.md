# Runtime boundaries

运行时相关代码按职责集中在三个 Go package；每个 package 内再按职责拆成文件，避免把依赖未导出符号的实现强行拆成子 package：

| 目录 | 责任 | 入口 |
| --- | --- | --- |
| `web` | Gin 路由、鉴权、请求解析、SSE/文件响应 | `web.NewServer` |
| `task` | 任务状态、SSE 回放、持久化、会话与交付 | `task.NewTaskManager` |
| `model` | provider 配置、模型工厂、fallback、限流、流清洗与压缩 | `utils.NewFallbackToolCallingChatModel` |

文件拆分不改变 Go package name、HTTP/SSE/JSON 字段或历史任务文件兼容规则。新增运行时文件应放入对应 package，并使用 `*_handler.go`、`*_stream.go`、`*_persistence.go` 等职责后缀。
