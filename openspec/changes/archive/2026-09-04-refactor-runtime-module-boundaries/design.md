## Context

当前后端采用 Go 单包分层，但三个文件承担了过宽的职责：`pkg/web/handler.go` 同时处理鉴权、任务、会话、下载、凭据和管理接口；`pkg/task/manager.go` 混合任务状态、SSE、持久化、会话和交付对账；`pkg/agent/utils/model.go` 混合 provider 配置、模型工厂、fallback、限流、流式清洗和请求观测。

这些代码已经有稳定的 API、SSE、任务文件和模型兼容契约。本变更只调整文件与内部依赖边界，保留现有公开方法、路由注册、事件字段、锁语义和错误行为。工作树当前存在其他功能改动，实施时不得回滚或覆盖它们。

## Goals / Non-Goals

**Goals:**

- 让 Web、Task、Model 三条主线按单一职责组织，核心文件控制在可扫描的规模。
- 保持 HTTP/SSE、`TaskState` 生命周期、provider/fallback、敏感信息脱敏和数据库行为不变。
- 将共享 helper、转换函数和接口放在清晰的单向依赖边界上。
- 每个阶段都能独立运行已有聚焦测试，并可在失败时通过文件级回滚恢复。

**Non-Goals:**

- 不更换 Web 框架、Eino、模型 SDK、数据库或前端架构。
- 不修改 `tasks.json`、SSE 事件名/终态、API response shape、provider 配置语义或业务流程。
- 不在本变更中重写 Planner、生成器或数据库 schema；前端仅允许为任务状态标签和恢复入口做必要的最小适配。
- 不为了“整齐”新增万能 `utils`、大而全 service 或跨包循环依赖。

## Decisions

### 1. 先按运行时二级目录归类，再在目录内按文件拆分

将三条运行时主线集中到 `pkg/runtime/{web,task,model}` 二级目录，保留原有 Go package 名称与导出 API；目录迁移只更新 import path，不把同一 package 的源文件拆到多个目录。这样 Go 的未导出符号、现有测试和 `*Server`/`*TaskManager` 接收者都无需改变，行为回归面最小。

候选文件边界：

- `runtime/web/auth_handler.go`、`task_handler.go`、`conversation_handler.go`、`credential_handler.go`、`admin_handler.go`；`server.go` 只保留装配和路由注册。
- `runtime/task/state_boundary.go`、`sse_stream.go`、`persistence_boundary.go`、`delivery.go`；公共 `TaskInfo`/`TaskStatus`/`SSERichEvent` 保持在稳定类型文件。
- `runtime/model/model_runtime.go`、`model.go`、`runtime_meta.go`、`compressor.go`；模型配置、fallback、限流和流安全按职责命名。

### 2. 采用“消费者定义接口、实现留在原域”

只有跨职责协作确实需要替换或测试隔离时才抽取小接口，例如 Web 使用任务查询/创建接口、模型 fallback 使用统一调用器。简单的纯函数不为了接口而接口，避免形成新的抽象迷宫。

### 3. 以契约测试作为拆分护栏

每搬迁一组函数，先迁移原有测试，再补充边界测试：路由表和响应、SSE 游标回放与终态、任务锁与恢复、provider 顺序/fallback、流式工具调用清洗和日志脱敏。测试只验证行为，不依赖文件名或内部布局。

### 4. 按风险顺序落地

实施顺序固定为 Web → Task → Model。每阶段保持可编译提交点：先移动纯 helper，再移动 handler/状态方法，最后整理依赖和删除旧文件。三阶段完成后才做统一 Linux 构建和部署，避免线上同时混入半套边界重构。

### 5. 术语与渲染入口收敛

新增代码使用面向业务动作的命名：`ppt`、`plan`、`page`、`download`、`sync`、`read`、`write`。`deck`、`manifest`、`materialize`、`reconcile` 仅在兼容既有导出符号、历史文件或第三方术语时保留，并应在注释中说明含义。

固定渲染链不再通过通用 Graph 组织：单一入口按顺序调用“读取并校验计划、下载/修订图片、并发生成页面”。内部 worker pool 和失败聚合保持不变。任务 JSON 的字段名保持兼容；文件名迁移采用“新写新名、读兼容旧名”的双读策略，确认历史任务淘汰后才移除旧入口。

## Risks / Trade-offs

- [符号移动导致隐式依赖或初始化顺序变化] → 每次只移动同一职责组，使用 `go test` 和 `go test -run` 聚焦验证，并保留原接收者和初始化入口。
- [SSE/锁代码拆分后引入竞态] → 不改变 `TaskState.Mu` 的所有权；先为事件游标、监听器和 conversation stream 增加并发回归测试，再删除重复实现。
- [模型模块拆分时意外改变 fallback 或脱敏] → 先冻结 provider chain、timeout、error classification 和 metadata 测试样例，拆分只允许改变调用位置。
- [当前工作树有并行未提交改动] → 实施前后用 `git diff` 做文件级审计，不覆盖无关文件；若拆分目标与并行改动冲突，暂停并记录冲突点。
- [文件数量增加后导航成本上升] → 采用职责后缀命名、在 package README/架构文档维护入口映射，并避免按函数粒度拆成几十个文件。

## Migration Plan

1. 建立当前行为基线：记录路由列表、关键 SSE 事件、聚焦测试与 `go build` 结果。
2. Web 阶段：按 handler 组移动实现，保留 `Server` 接收者和 `server.go` 路由；运行 `go test ./pkg/web ./pkg/task`。
3. Task 阶段：拆出状态/SSE、持久化、会话和交付实现；运行任务、数据库、SSE 恢复相关测试。
4. Model 阶段：拆出配置、工厂、fallback、限流、流清洗和观测；运行 model/provider/retry/compressor 测试。
5. 全量聚焦验证、`go test ./...`、`go build ./...`，再构建 Linux 交付物并部署。
6. 线上确认新进程、端口、启动日志、`/api/health` 及低成本任务链路；若失败，按阶段恢复对应文件级备份，不回滚无关用户改动。

## Open Questions

- 是否在三阶段完成后再把稳定接口提取到新 package，还是维持原 package 的文件边界？默认维持原 package，除非循环依赖或测试替身明确需要。
- `handler.go` 中的日志分析和管理接口是否继续归 `web`，还是下一轮独立为运营模块？本变更先归入 `admin_handler.go`，不扩大范围。
