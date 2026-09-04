# 长期迭代方向

> 本文件只保留仍在进行、待验证或需要长期追踪的事项。已完成能力基线见 [`done.md`](./done.md)，长期决策见 [`docs/decisions/`](../decisions/README.md)。

## 方向清单

- **Agent Harness 与可观测性**：持续完善 RuntimeMeta、可回放轨迹、预算与熔断、结构化 handoff、工具结果协议和用户可感知的实时状态。
- **PPT 视觉质量与 PPT Deck Planner 能力**：持续建设组件契约、容量控制、动态排版、图文混排、素材验证和低 AI 感评估体系。
- **生成编排与交付闭环**：持续优化路由、并行渲染、任务恢复、文件对账和线上最小烟测。
- **前端创作与交付工作台**：持续改善统一会话、模板/编排、过程可观测性、预览、错误分态和跨尺寸工作台体验。
- **评估、工程治理与运行可靠性**：持续建设质量评估、可重复部署、数据安全和运行故障恢复能力。
- **多模型供应商兼容层**：持续抽象 provider profile、账户级 Key、fallback 链、并发限制与模型能力差异。

## 当前执行事项

| ID | 事项 | Size | Route | Status | 计划时间 | 产出/下一步 |
| --- | --- | --- | --- | --- | --- | --- |
| PPT-TRACE-010 | 修复闲聊首轮完成后内存态未释放、第二轮仍返回“上一条对话仍在生成” | small | direct | pending | 2026-09-03 | 聚焦 `TaskState.BeginConversationStream`、`FinishConversationStream` 与冷加载恢复边界；线上首轮已持久化但第二轮仍会 409，需补回归后再发布。 |
| PPT-RESILIENCE-001 | 统一模型瞬时错误分类、流错误上抛与可恢复暂停状态 | medium | direct | implementation complete, deployment pending | 2026-09-04 | 本地后端聚焦测试、`go build ./...` 和前端构建已通过；当前工作树含未验证的无关运行改动，需隔离或确认一并发布后，完成 Linux 构建、部署与最小恢复链路烟测。 |
