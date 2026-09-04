# Runtime 模块边界重构基线

## 变更

OpenSpec：`openspec/changes/refactor-runtime-module-boundaries/`。

本次只做后端文件职责拆分，保持 HTTP API、SSE 事件、任务状态、数据库投影和模型 provider 行为不变。实施顺序为 Web → Task → Model。

## 当前基线

### 路由入口

- 鉴权：`/api/auth/{send-code,register,login,guest,set-password,logout,me}`
- 消息与草稿：`/api/messages`、`/api/plan-drafts{,/:id}`
- 任务：`/api/tasks`、`/api/tasks/:id`、`/api/tasks/:id/start`、`/api/tasks/:id/stream`、`/api/tasks/:id/files/:filename`、`/api/tasks/:id/thumb/:filename`、`/api/tasks/:id/cancel`、`/api/tasks/:id/feedback`、`/api/tasks/:id/continue`、`/api/tasks/:id/conversation`、`/api/tasks/:id/runtime-events/:event_id`
- 用户凭据：`/api/users/me/api-key`
- 模板、日志、管理：`/api/templates/layouts`、`/api/log-analyses{,/task/:task_id}`、`/api/admin/{stats,users,tasks,feedback,log-analyses}`
- 系统：`/api/health`、`/health/ready`、`/metrics`

### SSE 终态

`answer_end` 只结束说明文本；`complete`、`continue_complete`、`conversation_complete` 才是对应流的终态，`error` 保持错误事件语义。

### 测试基线

在拆分前运行：

```text
go test ./pkg/runtime/web ./pkg/runtime/task ./pkg/runtime/model ./pkg/agent/deck
ok   github.com/cloudwego/ppt-agent/pkg/web
ok   github.com/cloudwego/ppt-agent/pkg/task
ok   github.com/cloudwego/ppt-agent/pkg/agent/utils
ok   github.com/cloudwego/ppt-agent/pkg/agent/deck
```

## 文件边界约定

| 区域 | 文件职责 | 不负责 |
| --- | --- | --- |
| `pkg/runtime/web` | 请求解码、鉴权、服务调用、响应映射、路由装配 | Task 状态细节、模型创建、文件工作流实现 |
| `pkg/runtime/task` | 状态/锁、SSE、持久化、会话、交付对账 | HTTP 路由和 provider 配置 |
| `pkg/runtime/model` | 模型配置、provider 工厂、fallback、限流、流处理、观测、压缩 | 业务路由和任务数据库生命周期 |

三个运行时包均按职责后缀拆分实现文件；未创建只有 README 的子包，避免跨目录拆包造成未导出符号和 API 变化。

文件使用职责后缀（`*_handler.go`、`*_state.go`、`*_stream.go`、`*_persistence.go` 等），先保留原 package，避免为纯搬迁引入新依赖层。

## 2026-09-04 实施与上线证据

- 后端按职责迁移到 `pkg/runtime/web`、`pkg/runtime/task`、`pkg/runtime/model`；保留 JSON、HTTP、SSE 和历史任务文件兼容入口。
- 新增 `ppt/plan/download/sync` 直白命名入口、包责任注释及路由契约回归测试。
- 本地验证：`go test ./...`、`go build ./...`、Linux amd64 交叉构建均通过；前端 `npm run build` 通过。
- 部署目标：`remote-dev:/ppt/ppt-agent`，时间：2026-09-04；部署前备份目录为 `.deploy-backup-20260904-162831`，并保留 `frontend/dist.previous`。
- 新进程 PID：`2447429`，从 `/ppt/ppt-agent/backend` 启动以解析正确的静态资源路径，监听 `:8080`；启动日志显示 MySQL、Redis、Python 检查通过。
- 线上冒烟：`GET /api/health`、`GET /health/ready`、`GET /api/templates/layouts`、`GET /metrics` 均返回 200；根路径 `/` 返回 200，确认前端静态资源恢复。未创建高成本生成任务，未产生临时任务数据。
