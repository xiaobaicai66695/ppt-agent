# 幻灯片类型体系

本文档列出 `tasks.json` 唯一允许使用的 `content_type`。页面细节写入 `content_plan.components`，不要为相近排版另建类型；`layout_variant` 仅在对应类型明确支持时使用。

## 合法 content_type 清单

### 结构引导类
- `title_slide`: 标题页
- `agenda`: 目录页
- `section_divider`: 章节分隔页

### 内容陈述类
- `content_slide`: 普通内容页
- `quote_slide`: 金句/引言页
- `image_text`: 图文混排页

### 对比与并列类
- `card_grid`: 卡片阵列
- `comparison_table`: 对比表格页

### 流程与关系类
- `timeline`: 时间轴
- `kanban`: 看板进度页
- `brand_focus`: 品牌价值聚焦页

### 数据与图表类
- `kpi_dashboard`: 指标看板
- `chart_slide`: 图表专页，图表类型通过 `content_plan.chart_type` 或生成器参数 `chart_type` 表达

- `swot_analysis`: SWOT 分析页

## 禁止使用的历史别名

以下 id 不是合法 `content_type`，不能写入 `tasks.json` 或模板 JSON：

- `bar_chart`、`line_chart`、`pie_chart`、`doughnut_chart`: 改用 `chart_slide`，并在 `content_plan.chart_type` 中写 `bar`、`line`、`pie`、`doughnut`
- `table`: 改用 `comparison_table`
- `toc_slide`: 改用 `agenda`
- `smart_layout`: 改用 `content_slide`、`card_grid` 或其它现有布局
- `three_column`、`icon_grid`: 改用 `card_grid`
- `process_flow`: 改用 `timeline`，以 `process_step` 组件表达
- `stat_slide`: 改用 `kpi_dashboard`
- `case_study`、`example_detail`、`deep_dive`: 改用 `image_text`；架构说明可使用 `architecture_box` 组件
- `summary_slide`: 改用 `content_slide`

## 内容深度决策树

| 场景 | 类型 |
|------|------|
| 封面、章节、目录 | `title_slide` / `section_divider` / `agenda` |
| 介绍概念、总结、原理说明 | `content_slide` |
| 展示案例效果 | `image_text` |
| 多个平等要点 | `card_grid` |
| 多维对比 | `comparison_table` |
| 时间顺序 | `timeline` |
| 流程步骤 | `timeline`（使用 `process_step`） |
| 数据展示 | `chart_slide` / `kpi_dashboard` |
| 无特殊特征 | `content_slide` |
