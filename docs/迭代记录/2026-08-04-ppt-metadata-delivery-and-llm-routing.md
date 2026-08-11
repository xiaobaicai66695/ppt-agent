# 2026-08-04 PPT 元数据交付与 LLM 路由改造

## 迭代目标

解决页面已经全部生成但任务仍停留在“核对生成文件”的问题，同时移除首次生成中的 QA/自动修复流程，并让大模型直接决定任务意图与可执行 Agent 路由。本次事项为 `PPT-FLOW-001`，进入 OpenSpec change：`openspec/changes/ppt-agent-metadata-delivery-and-llm-routing/`。

## 已完成

- 意图分类改为单次 LLM 优先结构化决策，输出意图、领域、复杂度、页数、Agent、流水线、并发数、置信度和路由来源。
- Router 复用同一份分类结果，不再执行规则优先匹配或第二次语义分类；模型失败时只使用固定的 `deep / plan -> generate` 运行兜底。
- 首次生成的 旧多子代理架构 仅注册 SlideExecutor，永久跳过 Reviewer 和自动 Fixer；显式的交付后页面修改能力继续保留。
- 主 Agent prompt 移除 QA、自动修复、bash 批量转换和最终磁盘核对，最后一页返回后由任务管理器接管交付。
- TaskState 增加代码维护的 `DeliverySnapshot`，记录计划页数、已交付页数、声明文件和待处理任务；非零 N/N 是任务完成的唯一权威信号。
- 元数据达到 N/N 时，任务管理器使用内部成功原因结束剩余 Agent 控制循环，并发送确定性的完成消息，不再调用 LLM 检查磁盘。
- 元数据为空或不完整时保持失败语义；用户取消仍与内部交付完成取消严格区分。
- 前端状态文案改为元数据驱动的交付状态，不再将完成任务停留在“正在核对生成文件”。

## 验证与部署

- `go test ./...` 通过；`go build ./...` 通过。
- `go test -race ./pkg/task ./pkg/agent/learning` 通过。
- `npm test -- --run` 通过，共 13 项；`npm run build` 通过。
- `openspec validate ppt-agent-metadata-delivery-and-llm-routing --strict` 通过；`git diff --check` 通过。
- Linux 后端和前端已部署到 `/ppt/ppt-agent`，当前服务 PID `58715`，`/api/health` 返回 `{"status":"ok"}`。
- 在线生成任务 `4387279b-5773-467c-bbf5-cd20bef38d3b` 在页面产物完成后约 1.1 秒触发 `delivery_metadata_complete done=1 total=1`，API 终态为 completed，没有最终 LLM 文件检查。
- 当前版本路由实测输出 `agent_type=planner pipeline=plan,generate concurrency=5`；当前进程启动后的日志未出现 QA、Reviewer、Fixer 或 `intent_llm_fallback`。

## 关联内容

- `docs/issues/todo.md`
- `openspec/changes/ppt-agent-metadata-delivery-and-llm-routing/`
- `ppt-agent/backend/pkg/agent/intent/`
- `ppt-agent/backend/pkg/agent/router/engine.go`
- `ppt-agent/backend/pkg/agent/learning/engine.go`
- `ppt-agent/backend/pkg/agent/deck/agent.go`
- `ppt-agent/backend/pkg/prompts/planner/master_instruction.tmpl`
- `ppt-agent/backend/pkg/task/delivery_metadata.go`
- `ppt-agent/backend/pkg/task/manager.go`
- `ppt-agent/frontend/src/utils/workbench.ts`
