---
name: visual_designer
description: 指导 PPT Agent 规划 tasks.json、选择 content_type、填写 description/content_plan/layout_variant/background/source，并通过既有 Python generators 生成 PPT。
---

# Visual Designer

## 职责边界

本 Skill 的核心作用是指导 Agent **规划任务清单和生成器参数语义**，不是指导 LLM 手写视觉实现。

### 主 Agent 负责

- 选择合法 `content_type`，只使用 `references/slide_types.md` 中的英文 id。
- 决定整套 `theme`、`template`、页数、页面顺序和每页标题。
- 填写 `description`、`content_plan`、`layout_variant`、`background`、`source` 等 tasks.json 字段。
- 判断内容是否需要拆页、合并、改用更合适的页面类型。
- 为需要事实的数据页、案例页、图表页补充真实来源。

### SlideExecutor 负责

- 读取 tasks.json。
- 将 `description/content_plan` 转换为对应 generator 的 keyword 参数。
- 保持 `palette=manifest.theme`，逐字符使用 `task.output_file`。
- 将 `background`、`layout_variant`、`source` 等已规划字段原样传给支持它们的 generator。

### Python generators 负责

- 字号、行距、文本框宽高、换行、垂直居中、内容带居中。
- 背景图片解析、亮度处理、磨砂玻璃蒙版、文字对比度。
- 本地图标/图片/纹理选择，缺图时的默认图片或省略策略。
- 卡片、图表、KPI、流程、时间线等具体视觉绘制。

### 禁止混淆

- 不要在 tasks.json 或 prompt 中写具体字号、坐标、文本框宽度、蒙版透明度。
- 不要要求 LLM 估算文字几何尺寸；只控制内容长度和字段结构。
- 不要让 Agent 手写 `python-pptx` 绘图代码；所有页面必须复用 `generators` 包。
- 不要把 `layout_variant` 当作 `content_type`。

---

## tasks.json 规划原则

### 内容优先级

1. 用户当前主题、显式大纲、显式模板/配色/背景选择。
2. 当前任务识别出的受众、场景、交付目的。
3. 同领域历史偏好或用户确定性资料。
4. 模板默认结构。

模板示例内容只表达页面位置和叙事节奏，不能带入最终 PPT。

### 主题与配色

- 整套 PPT 只有一个顶层 `theme`，必须是 `references/palettes.md` 中的合法 palette id。
- `background` 只表示背景图片主题或具体图片引用，不得据此改写 `theme`。
- 禁止在 `description` 中写“使用 36pt 标题、蓝色渐变、深色蒙版”等实现细节。
- 如果用户显式指定品牌色或单位风格，将其写成语义约束，例如“保持正式政务语气”“突出单位名称”，不要写 CSS 或 RGB 指令。

### 页面类型选择

布局由内容性质驱动，而不是机械按要点数量决定。

| 内容性质 | 优先 content_type |
|---------|-------------------|
| 封面 | `title_slide` |
| 目录 | `agenda` |
| 章节转换 | `section_divider` |
| 核心概念/普通要点 | `content_slide` |
| 图文叙事/产品或场景说明 | `image_text` |
| 多个平等要点 | `card_grid` / `three_column` / `icon_grid` |
| 多维对比 | `two_column` / `comparison_table` |
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

---

## 内容容量与拆页

**Agent 只负责控制内容容量和拆页，具体排版适配由 generator 负责。**

### 推荐容量

| 类型 | 推荐容量 | 超出时处理 |
|------|----------|------------|
| `content_slide` | 4-6 条 bullet，每条 25-35 中文字 | 拆成概述页 + 详解页，或改用 `deep_dive` |
| `card_grid` | 4-6 张卡片，header 8-15 字，body 40-60 字 | 拆成两页卡片，或提炼为 4 张重点卡 |
| `two_column` | 左右各 3-5 条，或 2-3 个结构化区块 | 改用 `comparison_table` 或拆页 |
| `image_text` | 300-450 字自然段，不拆 bullet | 过长拆为两页图文叙事 |
| `kpi_dashboard` | 固定最多 4 个 KPI | 多指标拆成多页或改 `chart_slide` |
| `chart_slide` | 1 个主图表，1-3 个 dataset | 多图表拆页 |
| `timeline` / `process_flow` | 4-6 个节点 | 超过 6 个拆页或按阶段聚合 |
| `summary_slide` | 最多 4 条总结 | 合并相似结论 |

### 拆页判断

- 一个页面需要同时承载“背景、方案、数据、启示”且文字明显超过模板容量时，拆为多页。
- 内容不足时合并，避免为了页数制造空洞页面。
- 内容过多时拆分，避免把多个主题塞到一页。
- 拆页由主 Agent 在 tasks.json 中增加独立 `TaskItem` 完成；不要要求 generator 自动分页。
- 拆出的 `task_id`、`page_index`、`output_file` 必须保持唯一且有序，可使用 `3.1_架构总览.pptx`、`3.2_核心模块详解.pptx` 这类文件名。

---

## description 与 content_plan

### description 写法

`description` 是给 SlideExecutor 提取参数的自然语言契约，应包含页面要表达的核心内容，而不是视觉实现说明。

应写：

- 该页结论是什么。
- 包含哪些命名实体、事实、数字、时间、来源。
- 各字段之间的结构关系，例如“左栏讲原理，右栏放案例数据”。
- 图表页的数据 labels、datasets、chart_type。
- KPI 页的 value、label、delta、baseline。

不应写：

- 字号、坐标、形状宽高、透明度、阴影参数。
- “手动加图片占位框”“画一条装饰线”“用 python-pptx 绘制”等实现指令。
- `{要点1}`、`{小节标题}`、`某公司`、`若干数据` 等占位或匿名内容。

### content_plan 推荐字段

`content_plan` 用来让结构化内容更稳定，优先写业务语义：

```json
{
  "summary": "本页核心结论",
  "sections": [
    {"title": "背景", "items": ["..."]},
    {"title": "方案", "items": ["..."]}
  ],
  "chart_type": "bar",
  "data": {
    "labels": ["Q1", "Q2"],
    "datasets": [{"name": "收入", "values": [120, 180]}]
  },
  "visual_intent": {
    "role": "supporting_photo",
    "asset_query": "企业服务场景",
    "preferred_variant": "right_photo",
    "image_position": "right"
  }
}
```

`visual_intent` 只表达“需要什么视觉语义”，不是绘图参数。

可用 `role` 示例：

- `hero_photo`
- `supporting_photo`
- `chart`
- `cards`
- `icon`
- `map`
- `process`
- `timeline`

---

## 背景图片规划

背景图片由主 Agent 在 tasks.json 的 `background` 字段中规划，生成器负责实际解析和渲染。

### 使用规则

- 标题页、章节页、视觉冲击页、金句页、总结页：默认可使用背景图片。
- 图文叙事页、内容要点页：可使用背景图片，但正文容量较高时仍优先保证可读性。
- 图表、表格、KPI、强数据页：优先留空 `background`，使用干净背景。
- 相邻视觉页尽量不要重复同一张具体背景图。
- 如果后端或模板已分配具体图片引用，原样保留。

### 可用背景主题

| 主题标识 | 主题名称 | 适用场景 |
|----------|----------|----------|
| `party_government` | 党政办公 | 党建、政府汇报 |
| `ink_wash_mountain` | 水墨山水 | 中国风、文艺 |
| `vintage_chinese` | 复古中国风 | 传统文化 |
| `minimalist_blue` | 简约蓝白 | 商务、科技 |
| `snowy_mountain` | 雪山风景 | 自然、户外 |
| `artistic` | 艺术涂鸦 | 艺术、创意、时尚 |

### 字段形式

- 主题 id：`"party_government"`
- 主题内具体图片引用：`"party_government/images/3.jpg"`
- 不使用背景：`""`

不要在 `description` 中要求“压暗背景”“增加蒙版”“降低透明度”；这些由 generator 统一处理。

---

## layout_variant 与视觉意图

`layout_variant` 是同一 `content_type` 下的版式候选，只能写 generator 已支持的值。

当前优先使用：

| content_type | layout_variant 示例 |
|--------------|---------------------|
| `title_slide` | `photo_full_bleed_center` / `photo_full_bleed_left` / `editorial_split` |
| `section_divider` | `photo_band` / `number_sidebar` / `quiet_title` |
| `image_text` | `left_photo` / `right_photo` / `photo_strip` |
| `card_grid` | `equal_grid` / `featured_card_plus_grid` / `masonry_cards` |

其他页面如源码未显式支持 `layout_variant`，不要写该字段；可以只写 `content_plan.visual_intent` 表达语义。

---

## 本地素材与图片

- 默认优先使用本地素材库，不默认做外部图片搜索或生成式图片。
- `assets/manifest.json` 登记图标、图片、纹理等素材元数据。
- `background_templates/manifest.json` 登记主题背景。
- `image_text` 如果用户提供真实本地路径或 `asset:<photo-id>`，可写入 `content_plan.visual_intent` 或生成参数中的 `image_path` 语义；不要编造路径。
- 没有明确图片时，任务规划只需写 `asset_query` 或留空，具体默认图片由 generator 选择。
- 禁止写“图片占位”“虚线框”“待替换图片”等半成品文案。

素材维护和 generator 级别验证见 `references/generators.md` 与 `README.md`。

---

## 内容充实度标准

核心原则：内容页必须尽量包含具体、命名、带数字的事实。封面、目录、章节分隔、引用、总结等结构/叙事页按模板语义处理，不为了凑字段硬塞数据。

### 内容页优先包含

| 元素 | 要求 | 示例 |
|------|------|------|
| `kicker` | 分类标签，说明页面语义 | `核心技术 · CNN架构`、`案例 · 金融风控` |
| `lede` | 一句话结论，适合内容页/详解页 | `三次架构迭代将推理延迟从 320ms 压缩到 18ms` |
| 命名实体 | 公司、系统、人物、论文、政策、地点等真实名称 | `蚂蚁 AlphaRisk`，不是 `某金融公司` |
| 量化数据 | 数值、单位、时间、对比基准 | `准确率从 76% 提升至 99.99%` |
| 来源 | 数据页和事实密集页必须有来源 | `来源: 国家统计局 2025 | https://...` |

### 实例页 example_detail

```text
kicker: 实例 · {领域}
title: {命名案例}: {一句话总结}
context_block: 1-2 句行业/问题背景
solution_block: 2-3 句技术方案，含具体技术栈或方法
metrics: 至少 2-3 个量化指标，带单位和对比基准
takeaway: 1 句可迁移启示
```

### 深入页 deep_dive

左侧写原理、流程、关键设计决策；右侧写代码/架构/案例/数据。不要把 deep_dive 填成普通 bullet 页。

### 禁止的内容模式

- `"AI在各行业广泛应用"`：太空泛，没有具体行业、应用、数字。
- `"某大型互联网公司"`：匿名实体，必须给真实名称。
- `"显著提升了效率"`：必须写清楚提升幅度和基准。
- 连续 3 页纯 bullet_list 无案例或数据。
- 数据页没有 `source`。

---

## 变更纪律

- 常规页面生成必须复用 `skills/visual_designer/generators` 包。
- 修改 generator 函数签名时，同步 `references/generators.md`、SlideExecutor prompt、模板 JSON 和相关测试。
- 只调整 tasks.json 规划规则时，优先修改本文件和主 Agent prompt，不改 generator。
- 视觉实现细节应沉淀到 `generators/base.py`、具体 generator 或 `references/generators.md`，不要混入本文件作为 LLM 手工排版指令。
