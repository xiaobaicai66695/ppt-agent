# 幻灯片类型体系

本文档列出 `tasks.json` 中允许使用的唯一合法 `content_type`。规划、前端大纲、full-deck 模板和 SlideExecutor 都必须使用这些英文 id；图表形态、表格含义、视觉风格等细节应写入 `content_plan`，不要发明新的 `content_type`。

`layout_variant` 用于表达同一 `content_type` 下的具体排法，例如 `title_slide` 的 `photo_full_bleed_center` 或 `image_text` 的 `left_photo`。它不是新的页面类型，不能写入 `content_type` 字段。

## 合法 content_type 清单

### 结构引导类
- `title_slide`: 标题页
- `agenda`: 目录页
- `section_divider`: 章节分隔页
- `summary_slide`: 总结页

### 内容陈述类
- `content_slide`: 普通内容页
- `quote_slide`: 金句/引言页
- `image_text`: 图文混排页
- `image_hero`: 视觉冲击页

### 对比与并列类
- `two_column`: 双栏对比
- `three_column`: 三栏并列
- `card_grid`: 卡片阵列
- `comparison_table`: 对比表格页
- `icon_grid`: 图标网格页

### 流程与关系类
- `timeline`: 时间轴
- `process_flow`: 步骤流程图
- `kanban`: 看板进度页
- `region_map`: 区域版图页
- `brand_focus`: 品牌价值聚焦页

### 数据与图表类
- `stat_slide`: 关键数字页
- `kpi_dashboard`: 指标看板
- `chart_slide`: 图表专页，图表类型通过 `content_plan.chart_type` 或生成器参数 `chart_type` 表达

### 内容叙事类
- `example_detail`: 实例详解页（命名案例 + 背景 + 方案 + 数据 + 启示）
- `deep_dive`: 深入详解页（双栏：左原理右代码/架构/数据）
- `case_study`: 案例研究页（Context -> Problem -> Solution -> Results）
- `swot_analysis`: SWOT 分析页

## 禁止使用的历史别名

以下 id 不是合法 `content_type`，不能写入 `tasks.json` 或模板 JSON：

- `bar_chart`、`line_chart`、`pie_chart`、`doughnut_chart`: 改用 `chart_slide`，并在 `content_plan.chart_type` 中写 `bar`、`line`、`pie`、`doughnut`
- `table`: 改用 `comparison_table`
- `toc_slide`: 改用 `agenda`
- `smart_layout`: 改用 `three_column`、`deep_dive`、`card_grid` 或其它现有布局

## 内容深度决策树

| 场景 | 类型 |
|------|------|
| 封面、章节、目录、总结 | `title_slide` / `section_divider` / `agenda` / `summary_slide` |
| 介绍概念 | `example_detail` 或 `content_slide` |
| 讲解原理 | `deep_dive` |
| 展示案例效果 | `case_study` |
| 多个平等要点 | `card_grid` / `three_column` / `icon_grid` |
| 多维对比 | `two_column` / `comparison_table` |
| 时间顺序 | `timeline` |
| 流程步骤 | `process_flow` |
| 数据展示 | `chart_slide` / `kpi_dashboard` / `stat_slide` |
| 无特殊特征 | `content_slide` |
