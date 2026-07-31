# Agent 可观测与 Visual Designer Skill 改造预研

## 背景

当前 PPT Agent 的质量问题主要来自三条链路：

- 执行过程不可回放：只看到状态和日志片段，难以定位是哪一页、哪个 agent、哪个工具拖慢或失败。
- Skill 契约不稳定：`SKILL.md` 要求高，但 `templates/single-page/*.json` 和 generator 参数不同步，Agent 无法稳定填充高级字段。
- 在线 QA 成本高：多模态 QA/Reviewer 增加时间和模型成本，用户已决定默认放弃该流程。

## 目标与非目标

目标：

- 在不引入外部观测系统的前提下，提升 Agent 执行过程可追踪性。
- 将 PPT 质量控制前移到 schema、layout selector、容量预算和本地验证。
- 保留 QA 作为显式手动能力，不作为默认路径。

非目标：

- 不默认运行多模态 QA。
- 不引入新的数据库表或外部 tracing 服务。
- 不一次性重写所有 generator。
- 不用高成本 LLM judge 作为默认验证。

## 现状判断

### 可观测性

已完成 `RuntimeMeta` 快照和 SSE 状态栏，但它更像仪表盘，不是完整飞行记录。缺少：

- 可回放事件序列。
- 每个 LLM/tool 调用的 span 关系。
- 每页 slide 的执行账本。
- 任务结束后的诊断报告。

### Visual Designer

主要缺口：

- `SKILL.md` 要求 kicker、lede、实例、量化数据、来源，但部分模板 JSON 只暴露 title/bullets/cards。
- `generators.md` 和实际 generator 参数存在漂移。
- 缺少每个布局的容量声明，Agent 不知道何时拆页或换布局。
- 视觉系统只有 palette，缺少完整 design tokens 和 layout density。
- 背景图片策略偏粗，不适合所有信息页。

## 方案选项

### 方案 A：继续强化 QA/Reviewer

优点：理论上能发现视觉问题。

缺点：成本高、耗时长、受多模态模型稳定性影响，且问题发生在生成之后，修复链路可能继续引入新问题。

结论：不推荐作为默认路线。

### 方案 B：本地 Runtime Event Log + Report

优点：成本低、可渐进实现、崩溃后仍保留现场，能直接支撑前端 Timeline 和失败诊断。

缺点：需要定义事件 schema，并处理日志体积和敏感信息截断。

结论：可行，推荐优先做。

### 方案 C：OpenTelemetry 全链路 tracing

优点：标准化，后续可接 Grafana/Tempo/Jaeger。

缺点：引入复杂度较高，当前项目本地调试和单任务场景暂不需要。

结论：先不做，等 JSONL/report 稳定后再考虑 exporter。

### 方案 D：Visual Designer 契约升级

优点：把质量控制前移，能减少生成后修复成本；与用户放弃 QA 的决策一致。

缺点：需要同步模板 JSON、generator 文档、prompt 和前端展示，属于 large 变更。

结论：可行，应走 OpenSpec。

## 推荐方案

1. 直接执行：默认关闭 QA 流程，仅当 `ENABLE_QA=true` 且配置允许时启用。
2. OpenSpec 变更一：`ppt-agent-runtime-event-log`，实现 runtime JSONL、report、slide ledger 和 Dashboard Timeline。
3. OpenSpec 变更二：`visual-designer-contract-upgrade`，统一模板 schema、capacity、layout selector 和轻量本地验证。
4. 暂缓：OpenTelemetry、LLM-as-judge、自动图片生成/检索、所有 generator 重写。

## 风险与待确认问题

- 事件日志是否写入每个 task work_dir，还是集中写到全局 logs 目录。
- 前端 Timeline 首版是否只展示当前任务，是否需要支持历史任务冷加载。
- `content_plan` 强 schema 应由后端定义，还是跟随 skill 的 JSON schema 定义。
- 背景图片是否默认关闭，只在 title/section/hero 类页面启用。
- 是否保留 Fixer 流程，还是和 QA 一起降级为手动继续模式。

## 结论

- 放弃默认 QA 是可行的，且符合当前时间/成本目标。
- 可观测性升级和 skill 契约升级都可行，但均属于跨模块 large 变更，应登记后优先走 OpenSpec。
- 拿不准的问题已记录到 `docs/xq-todo.md`，待共同确认后再实施。

## 关联事项

- TODO: `docs/issues/todo.md#PPT-QA-001`
- TODO: `docs/issues/todo.md#PPT-OBS-001`
- TODO: `docs/issues/todo.md#PPT-SKILL-001`
