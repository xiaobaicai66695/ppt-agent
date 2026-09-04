# 架构决策索引

这里记录仍约束当前实现的长期决策；一次性故障修复和发布细节不在此重复。

| 决策 | 适用范围 |
| --- | --- |
| [001 Agent Harness 与可观测性](./001-agent-harness-observability.md) | 运行事件、Timeline、上下文压缩与诊断数据。 |
| [002 PPT Deck Planner 与视觉质量](./002-visual-designer-quality.md) | DeckSpec、组件契约、图片物化和 Python 生成器。 |
| [003 生成编排与交付闭环](./003-generation-delivery-flow.md) | Planner/Reviewer/Fixer、提交、渲染和文件对账。 |
| [004 前端工作台与交付体验](./004-frontend-workbench-delivery.md) | 统一会话、SSE、预览、工作台呈现。 |
| [005 运维、治理与可靠性交付](./005-ops-governance-reliability.md) | 验证、部署、冒烟、运行证据与清理。 |
| [006 对话、检索与用户可见轨迹](./006-conversation-retrieval-boundary.md) | 会话任务、意图路由、资料检索和安全可见性。 |
| [007 多模型与账户凭据边界](./007-model-provider-credential-boundary.md) | provider profile、账户 Key、fallback 和敏感信息保护。 |
