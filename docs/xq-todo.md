# XQ 待商讨事项

本文件记录当前还不适合直接实施、需要用户和 Agent 共同确认的事项。

| ID | 议题 | 当前判断 | 待确认问题 |
| --- | --- | --- | --- |
| XQ-001 | Runtime events 存储位置 | 已确认：每个任务写入 `work_dir/runtime_events.jsonl`，结束时生成 `runtime_report.json` | 暂不做全局索引 |
| XQ-002 | Dashboard Timeline 首版范围 | 已确认：首版只展示当前任务的本地事件流 | 暂不做历史任务冷加载和跨任务对比 |
| XQ-003 | QA 与 Fixer 的关系 | 已确认：默认放弃 Reviewer/在线 QA；Fixer 保留为用户继续对话后的手动修复工具 | 自动 Fixer 不作为默认主链路 |
| XQ-004 | Visual Designer 背景图策略 | 已确认：title/section/hero 使用，信息密集页默认不用 | 暂不做全局强制背景 |
| XQ-005 | `content_plan` 强 schema 归属 | 倾向后端定义核心 schema，skill 模板声明容量和字段 | schema 是否需要由 OpenSpec 固化为长期契约 |
| XQ-006 | 图标/图片资产 | 已确认：先做本地图标和形状 primitives，不接外部图片搜索 | 图片检索/生成式图片资产后续再议 |
| XQ-007 | 质量评分方式 | 倾向轻量本地验证 + eval fixture，不默认 LLM judge | 是否需要保留可选的人工/LLM 评分入口 |
| XQ-008 | 主流程规划质量门禁 | 倾向新增 Planner -> Reviewer -> Refiner 的渲染前循环，规划未通过结构、容量、场景匹配和当前任务风格要求前不进入生成 | 循环上限、失败降级策略和是否固化为 OpenSpec change |
| XQ-009 | 组件级 DeckSpec | 倾向先在规划阶段引入组件级 `components`，渲染仍保持按页 worker pool；坐标、字号、颜色继续由 generator 控制 | 首批覆盖哪些 `content_type`，以及是否兼容现有 `content_plan.elements` |
| XQ-010 | 组件级 QA / 局部修复 | 倾向作为第二阶段能力，只对结构化组件做内容修订或缩短，不在首版重构为组件级独立渲染任务 | 是否需要先建设组件 ID、QA issue schema 和局部 refiner |
