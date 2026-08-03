# 2026-07-31 PPT Agent 迭代记录

## 今日已完成

- 建立并归档 OpenSpec change：`openspec/changes/archive/2026-07-31-ppt-agent-harness-runtime/`。
- 增加任务级 `RuntimeMeta`：阶段、耗时、工具调用、工具错误、同参重复、token、压缩、QA 计数、slide 进度、缺失文件和预算 warning。
- 将 runtime status 注入 DeepAgent、SlideExecutor、Reviewer、Fixer 的模型调用。
- 将 runtime metadata 接入 task manager、SSE `runtime_meta` 事件和 Dashboard 开发者状态栏。
- 增加离线 PPT eval fixture 和 artifact evaluator：`docs/eval/ppt_quality_cases.json`、`ppt-agent/scripts/eval_ppt_quality.py`。
- 归档后的 OpenSpec specs 已生成在 `openspec/specs/`。

## 本轮继续落地

- 按 XQ-001/XQ-002 确认范围建立任务本地 runtime event 轨迹：运行中追加 `work_dir/runtime_events.jsonl`，终态写入 `work_dir/runtime_report.json`，暂不做全局索引或历史冷回放。
- Dashboard 开发者状态栏增加当前任务 Timeline，直接消费 SSE `runtime_meta.recent_events`，展示事件类型、阶段、组件名、状态、耗时和摘要。
- 按 XQ-003 固化默认放弃在线 Reviewer/QA：`ENABLE_QA=true` 才启用；Fixer 保留为用户继续对话后的手动修复路径。
- 按 XQ-004/XQ-006 更新 Visual Designer 契约：标题/章节/视觉页推荐背景图，信息密集页默认 clean background；视觉元素优先使用本地图标、glyph、形状、卡片和图表 primitives。
- 为 12 个高频单页模板补充 `contract` 元数据：容量、必填字段、适用/避免场景、溢出策略、背景策略和视觉 primitives。
- OpenSpec changes：`openspec/changes/ppt-agent-runtime-event-log/`、`openspec/changes/visual-designer-contract-upgrade/` 均已通过 `--strict` 校验。

## 本轮追加落地：动态排版与本地素材库

- 建立 `openspec/changes/visual-designer-dynamic-assets/`，对应 `PPT-SKILL-002`。
- 新增 `visual_designer/assets/manifest.json`，登记 24 个核心 PNG 图标、6 张编辑型浅色背景、4 张 subtle pattern。
- 新增 `generators/asset_manager.py` 和 `generators/layout_intelligence.py`，用于素材选择、图标放置、内容密度判断、动态字号和对齐策略。
- 接入 `title_slide`、`section_divider`、`content_slide`、`card_grid`、`icon_grid`、`quote_slide`、`summary_slide`，少字内容页可切换到居中焦点布局，卡片/图标页可使用本地语义图标。
- 修复 `image_hero` 传入背景时 `colors` 未初始化的问题。
- 生成 review deck 并通过 LibreOffice PDF 转换、Poppler PNG 渲染和总览图检查。

## 新决策

- 为节省时间和模型成本，默认放弃在线 QA/Reviewer 流程。
- 质量改进重心从“生成后多模态 QA 修复”转向“生成前强契约 + 生成中可观测 + 生成后轻量本地验证”。
- QA 相关字段和历史状态暂不删除，作为兼容旧任务和未来手动启用的可选能力。

## 可继续推进的方向

### 1. Agent 可观测性升级

当前 `RuntimeMeta` 是状态快照，下一步应升级为可回放的执行轨迹：

- `runtime_events.jsonl`：追加式记录 ToolStarted、ToolFinished、ToolFailed、LLMStarted、LLMFinished、PhaseChanged、FileCreated、ManifestValidated、CompressionRan、BudgetWarningRaised。
- `runtime_report.json`：任务结束后输出阶段耗时、工具失败率、同参重复、缺失文件、token、压缩、slide 结果摘要。
- Dashboard Timeline：按 agent、phase、slide、error/warning 筛选执行过程。
- Slide Execution Ledger：按页记录生成尝试、工具调用、输出文件、失败原因、轻量验证结果。

### 2. Visual Designer Skill 升级

当前 `visual_designer` 的主要问题是 skill 文档、模板 JSON、generator 参数和 Agent 消费方式没有形成单一契约。下一步优先：

- 同步 `templates/single-page/*.json` 与 `references/generators.md`，让真实 generator 参数进入模板 schema。
- 为每个模板补 `capacity`、`required_fields`、`best_for`、`avoid_for`、`overflow_strategy`。
- 将 `content_plan` 升级为更强的结构化输入，尤其覆盖数据、来源、案例、图表和容量等级。
- 建立 layout selector，让 Agent 先按内容性质选布局，再填参数。
- 增加轻量本地渲染验证，优先检查文本溢出、重叠、缺失输出和 manifest 一致性。

## 暂不推进

- 不做默认在线 QA/Reviewer。
- 不做高成本 LLM-as-judge。
- 不急于接 OpenTelemetry/Jaeger/Tempo；先落本地 JSONL 和 report，降低接入成本。

## 关联事项

- `docs/issues/todo.md`
- `docs/research/2026-07-31-observability-skill-feasibility.md`
- `docs/xq-todo.md`
