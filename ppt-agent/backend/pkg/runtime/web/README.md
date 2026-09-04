# Web layer

Web 层现在按请求边界分为三个 package，顶层 `web` 只保留兼容性的
`Server` 装配入口和既有 handler 实现：

| Package | 责任 | 入口 |
| --- | --- | --- |
| `web/router` | Gin 路由表、分组、中间件挂载和 HTTP handler 适配 | `router.Register` |
| `web/service` | 任务创建、会话任务启动、重复请求抑制和大纲校验 | `service.NewTaskService` |
| `web/model` | 请求、响应、路由结果和健康检查 DTO；不依赖 Gin | 结构体与契约常量 |

`web.NewServer` 仍是对外兼容入口：它创建运行时依赖，将 handler 函数注入
`router.Handlers`，再由 `router.Register` 注册完整 API。旧包级类型通过 type
alias 指向 `web/model`，因此 API、SSE 事件和现有测试无需迁移。

依赖方向固定为 `router → service/model`，`service → model/task`；`model`
不反向依赖 `web`，避免循环依赖和把 Gin 类型渗透到业务层。
