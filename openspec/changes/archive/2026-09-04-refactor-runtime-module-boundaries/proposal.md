## Why

PPT Agent 的主要运行行为已经稳定，但 `web/handler.go`、`task/manager.go` 和 `agent/utils/model.go` 分别承载了过多互不相同的职责，阅读、定位故障和安全修改的成本持续上升。现在进行只保持行为不变的边界重构，可以在继续迭代前消除历史叠加造成的耦合，而不改变 API、SSE、任务契约或模型选择语义。

## What Changes

- 将 Web 层按鉴权、任务、会话、交付、凭据和管理接口拆成职责明确的 handler 文件；路由注册和 HTTP 错误语义保持不变。
- 将 TaskManager 按状态/SSE、生命周期、持久化、会话和交付同步拆分；保留现有任务状态机、事件游标、锁边界和恢复行为。
- 将模型工具按配置解析、provider 工厂、fallback/限流、流式清洗、请求观测和上下文压缩拆分；保留现有 provider、超时、fallback 顺序和敏感信息保护。
- 收敛代码、Agent 与工具的内部术语：新代码优先使用 `ppt`、`plan`、`page`、`download`、`sync`、`read`、`write` 等直白名称；保留既有 JSON 字段、HTTP/SSE 字段和历史文件读取兼容，避免一次性破坏任务恢复。
- 将固定的“校验计划 → 准备图片 → 并发生成页面”渲染链收敛为可直接阅读的顺序入口；保留并发 worker、图片修订和现有错误契约，不为三步固定流程保留额外 Graph 包装。
- 将跨包共享的接口和纯转换函数放在最小、单向依赖的边界上，禁止通过新增万能 `utils` 或循环依赖重新聚合职责。
- 按拆分后的模块迁移现有测试，并为每一阶段补充边界级回归测试。
- 不改变公开 HTTP API、SSE 事件名称和终态规则、JSON 字段、数据库 schema 或部署方式；任务文件名迁移须保留旧 `tasks.json`/`tasks.draft.json` 的读取兼容。

## Capabilities

### New Capabilities

- `web-handler-boundaries`: Web 路由按业务职责组织，保持对外 HTTP/SSE 契约不变。
- `task-runtime-boundaries`: 任务状态、事件流、持久化、会话和交付职责可独立阅读和测试。
- `model-runtime-boundaries`: 模型供应商解析、fallback、限流、流处理和观测职责可独立演进。

### Modified Capabilities

无。此变更只重组内部实现，不修改既有能力的外部需求。

## Impact

- 后端：`ppt-agent/backend/pkg/web`、`pkg/task`、`pkg/agent/utils`，以及 `main.go` 中的路由/服务装配。
- 测试：对应 package 的 Go 单测、SSE/任务恢复测试、模型 fallback 测试。
- 交付：不新增运行时依赖；完成后需要 `go test ./...`、`go build ./...`、前端构建和现有最小线上冒烟。
- 本次重构不触及前端组件、Python 生成器和数据库迁移，除非验证发现仅为编译引用调整。
