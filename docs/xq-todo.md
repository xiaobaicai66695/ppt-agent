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
