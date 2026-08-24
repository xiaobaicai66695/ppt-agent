---
name: ppt-deck-planner
description: 指导 PPT Agent 规划 DeckSpec/tasks.json、选择 content_type、填写 description/content_plan/layout_variant/source，并通过组件化 Python generators 生成 PPT。
---

# PPT Deck Planner

本 Skill 的核心作用是指导 Agent 规划可执行的 `tasks.json`，不是让 LLM 手写视觉实现。

## 职责边界

主 Agent 负责：

- 选择合法 `content_type`，只使用 `references/slide_types.md` 和 `templates/component_contracts.json` 中的英文 id。
- 决定整套 `theme`、`template`、页数、页面顺序和每页标题。
- 填写 `description`、`content_plan.slide_intent`、`content_plan.components`、`layout_variant`、`source` 等字段。
- 判断内容是否需要拆页、合并或改用更合适的页面类型。
- 为数据页、案例页、图表页补充真实来源。
- 需要视觉素材时，把主题转换成可搜索的视觉主体，并把下载后的图片路径和署名写入 `image` 组件或 `visual_intent`。

SlideExecutor 负责：

- 读取 `tasks.json`。
- 优先消费 `content_plan.components`；旧 `elements/summary/description` 只作为兼容兜底。
- 保持 `palette=manifest.theme`，逐字符使用 `task.output_file`。
- 将 `layout_variant`、`source` 和图片字段传给 generator。

Python generators 负责：

- 使用 `component_layout.py` 消费语义组件并选择布局区域。
- 控制字号、行距、文本框、换行、垂直居中、内容带居中和溢出保护。
- 解析 `image.local_path` / `visual_intent.local_path`，嵌入真实图片。
- 绘制卡片、列表、长论述、图表、KPI、流程、时间线、表格和架构关系。

禁止：

- 在 tasks.json 中写坐标、字号、颜色、边距、透明度、卡片尺寸等视觉实现参数。
- 要求 LLM 估算文字几何尺寸或手写 `python-pptx` 绘图代码。
- 把 `layout_variant` 当作 `content_type`。
- 在新规划中使用旧本地背景主题、离线 asset id 或旧 full-deck/single-page 模板文件。

## 规划入口

规划前读取：

- `templates/component_contracts.json`：组件类型、content_type 容量和已实现 `layout_variant` 的单一契约。
- `references/slide_types.md`：合法页面类型说明。
- `references/generators.md`：只有在需要确认 generator 参数或生成器行为时读取。

不要读取已删除的旧目录：`templates/full-decks/`、`templates/single-page/`、`assets/`、`background_templates/`、`scripts/`。

## 页面类型选择

| 内容性质 | 优先 content_type |
|---------|-------------------|
| 封面 | `title_slide` |
| 目录 | `agenda` |
| 章节转换 | `section_divider` |
| 核心概念/普通要点 | `content_slide` |
| 图文叙事/场景说明 | `image_text` |
| 多个平等要点 | `card_grid` / `three_column` / `icon_grid` |
| 多维对比 | `comparison_table` / `two_column` |
| 时间顺序 | `timeline` |
| 流程步骤 | `process_flow` |
| 关键数字 | `stat_slide` |
| 多指标看板 | `kpi_dashboard` |
| 趋势、占比、对比数据 | `chart_slide` |
| 命名案例 | `example_detail` / `case_study` |
| 原理、架构、深度分析 | `deep_dive` |
| 金句、观点引用 | `quote_slide` |
| 总结收束 | `summary_slide` |

历史别名 `bar_chart`、`line_chart`、`pie_chart`、`doughnut_chart`、`table` 不能写入新的 tasks.json；统一用 `chart_slide` 或 `comparison_table`。

## 整体结构

整套 PPT 优先采用“观点 -> 论据 -> 推论/行动”的叙事结构，而不是连续罗列要点。

- 每个章节先用 1 页 `section_divider` 明确阶段主题；章节少于 2 个时可以省略分割页。
- 内容页先确定本页主观点，再选择 2-4 个论据组件支撑；论据可以是事实、数据、图片、案例、引用或表格。
- 深度说明页优先使用 `argument_block` 承载完整论述，再配列表、证据、KPI、图片或架构组件。
- 对比、选型、方案评估优先用 `comparison_table` 或 `two_column`，并用 `recommendation` 给出结论。
- 每 3-5 页安排一次节奏变化：章节页、图文页、数据页、案例页或总结页。
- 同一页只讲一个中心判断；多个判断并列时拆页或改成目录/框架页。

## 内容容量

Agent 只控制内容容量和拆页，具体排版适配由 generator 负责。

| 类型 | 推荐容量 | 超出时处理 |
|------|----------|------------|
| `content_slide` | 4-6 条 bullet，每条 45-85 字 | 拆成概述页 + 详解页，或改用 `deep_dive` |
| `card_grid` | 4-6 张卡片，body 80-140 字 | 拆成两页卡片，或提炼为 4 张重点卡 |
| `two_column` | 左右各 3-5 条，或 2-3 个结构化区块 | 改用 `comparison_table` 或拆页 |
| `image_text` | 300-450 字自然段，配 1 张真实图片 | 过长拆为两页图文叙事 |
| `argument_block` | 440-840 字完整论述 | 超过容量时拆成多页 |
| `kpi_dashboard` | 3-4 个 KPI | 多指标拆页或改 `chart_slide` |
| `chart_slide` | 1 个主图表，1-3 个 dataset | 多图表拆页 |
| `timeline` / `process_flow` | 4-6 个节点 | 超过 6 个拆页或按阶段聚合 |
| `summary_slide` | 3-5 条总结 | 合并相似结论或拆出行动页 |

## description 与 components

`description` 应说明本页要表达的核心内容，而不是视觉实现说明。

应写：

- 本页结论。
- 命名实体、事实、数字、时间和来源。
- 字段之间的结构关系，例如“左栏讲原理，右栏放案例数据”。
- 图表页的 `labels`、`datasets`、`chart_type`。
- KPI 页的 `value`、`label`、`delta`、`baseline`。

不应写：

- 字号、坐标、形状宽高、透明度、阴影参数。
- “手动加图片占位框”“画装饰线”“用 python-pptx 绘制”等实现指令。
- `{要点1}`、`某公司`、`若干数据` 等占位或匿名内容。

常用组件类型：

| 组件类型 | 用途 |
|----------|------|
| `headline` / `subheadline` | 主论点和副标题 |
| `argument_block` | 大段论述正文 |
| `paragraph` / `text_block` | 正文叙事或短文本块 |
| `list` / `numbered_list` / `bullet_list` / `evidence_list` | 通用列表、有序列表、要点列表或证据列表 |
| `feature_card` / `fact_card` / `key_point` / `insight` | 能力卡片、事实卡片、重点块和洞察 |
| `recommendation` / `risk_item` / `opportunity_item` / `decision_item` | 建议、风险、机会和决策项 |
| `case_snapshot` / `toc_item` | 案例摘要和目录项 |
| `kpi_metric` / `stat` / `number_callout` | KPI 或关键数字 |
| `chart` / `table` / `comparison_matrix` | 图表、表格和对比矩阵 |
| `timeline_node` / `process_step` / `milestone` | 时间线节点、流程步骤或里程碑 |
| `image` / `map` / `diagram` / `architecture_box` / `arrow` | 图片、地图、图解、架构模块和关系箭头 |
| `quote_block` / `callout` | 引用或强调信息 |
| `source_note` | 来源说明 |

## 图片搜索规划

当前 Planner 不依赖多模态能力查看候选图，因此必须先把页面视觉意图转换成可搜索词。

| `asset_purpose` | 适用内容 | 查询写法 |
|-----------------|----------|----------|
| `background` | 整页氛围和文字承载 | 视觉代理 + 环境/光线 + `wide landscape` + `clean negative space` |
| `scene` | PPT 中插入的真实实例 | 具体对象 + 动作 + 环境/地点 |
| `evidence` | 案例或事实的实景证据 | 行业/对象 + 可见设施或动作；事实来源仍写入 `source` |
| `decorative` | 辅助装饰 | 主题实体 + 简洁构图 |

例如“低空经济”不直接作为背景查询词：背景可转换为 `aerial city skyline at blue hour, wide landscape, clean negative space`，实景可转换为 `delivery drone flying above urban neighborhood`。

调用 `search_images(download=true)` 后，必须把选中图片的 `local_path`、`image_url`、`preview_url`、`source_url` 和 `attribution` 写回对应 `image` 组件或 `visual_intent`。图片保存路径应在当前 PPT 任务工作目录内。

顶层 `task.background` 是历史本地背景字段，新规划保持为空。外部图片用途写在 `visual_intent.asset_purpose` 或 `image.asset_purpose` 中。

## layout_variant

`layout_variant` 只填写 `templates/component_contracts.json` 中当前 `content_type` 明确列出的值。空 `variants` 表示生成器按组件密度自适应排版。

当前稳定支持：

| content_type | layout_variant |
|--------------|----------------|
| `section_divider` | `number_sidebar` |

## 计划审查

Planner 写入 DeckSpec 前必须调用 `review_tasks_manifest`。若未通过，按 critique 修订草稿并重新 review，最多 3 轮。

`reviewer_status.issues[].code` 使用固定枚举：

- `intent_mismatch`
- `profile_overfit`
- `weak_narrative`
- `low_information_density`
- `overload_capacity`
- `invalid_component_schema`
- `missing_data_or_fact`
- `layout_mismatch`
- `missing_background_image`

通过质量门后设置 `reviewer_status.locked=true`。质量门包括：用户意图匹配、用户画像没有跨场景过拟合、页数合理、`content_type` 合法、组件合法、组件数量不超过容量、信息密度不空不爆、图片规划与主题一致。

## 变更纪律

- 常规页面生成必须复用 `skills/ppt-deck-planner/generators` 包。
- 只调整 tasks.json 规划规则时，优先修改本文件和主 Agent prompt，不改 generator。
- 修改 generator 函数签名时，同步 `references/generators.md`、SlideExecutor prompt 和相关测试。
- 视觉实现细节沉淀到 `generators/base.py`、具体 generator 或 `references/generators.md`，不要混入本文件。
