# Generators 参考文档

本文档是 `ppt-deck-planner` Python generators 的函数参考。推荐由调用方先准备包含 `tasks.json` 的工作目录，运行 `generators/validate_deck.py` 预检，再通过 `generators/render_deck.py` 渲染整套 PPTX；调试单页时可用 `generators/render_task.py` 按 `task_id` 渲染。当前生成器使用组件式布局：`content_plan.components` 是页面正文和视觉语义的唯一主入口，由 `generators/component_layout.py` 决定页面结构。

## 导入方式

```python
import sys
from pathlib import Path

generators_pkg_dir = Path("<ppt-deck-planner 绝对路径>").resolve()
sys.path.insert(0, str(generators_pkg_dir))

from generators import (
    new_presentation, save_presentation, save_slide,
    generate_title_slide, generate_section_divider, generate_content_slide,
    generate_quote_slide, generate_card_grid, generate_timeline,
    generate_two_column, generate_image_text, generate_agenda,
    generate_kpi_dashboard, generate_chart_slide, generate_swot_analysis,
    generate_comparison_table,
    generate_image_hero, generate_kanban, generate_brand_focus, generate_region_map,
)
```

## 调用规范（必须严格遵守）

- `new_presentation(palette="xxx")` — 创建空演示文稿（不含幻灯片），每个 PPTX 文件必须单独创建；新规划不传顶层 theme，默认基础 palette 为 `ocean_soft`
- 每个 `generate_xxx` 函数必须传入 `prs` 参数（即使为 None，生成器内部会自动 new_presentation）
- `save_slide(slide, output_path)` — 保存单个 slide 为 PPTX 文件（推荐用法）
- `save_slide` / `save_presentation` 会在导出时为每页写入默认转页动画；默认是中速点击推进的 `fade`，可通过环境变量 `PPT_SLIDE_TRANSITION=none|fade|push|wipe|split|cover|uncover`、`PPT_SLIDE_TRANSITION_SPEED=slow|med|fast`、`PPT_SLIDE_TRANSITION_DIRECTION=l|r|u|d` 调整
- **每个 PPTX 文件 = new_presentation + 一次 generate + save_slide**，禁止复用 prs 生成多个文件
- **所有参数都用 keyword 形式传递**（如 `palette="ocean_soft"`），不要依赖位置参数
- **只传函数接受的参数**，参数表是常用速查；如果本文档和源码不一致，以 `generators/<content_type>_generator.py` 的实际函数签名为准

## 通用参数

所有 generate 函数共享以下参数：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `prs` | `Optional[Presentation]` | `None` | 已有的 Presentation 对象，为 None 时自动创建 |
| `palette` | `str` | `"ocean_soft"` | 基础配色名。新任务不由 Planner 写顶层 `theme`；带背景图页面由生成器从背景图提取弱化色系 token，用于面板、色块、分割线和强调装饰，正文文字仍使用可读文本 token |
| `source` | `str` | `""` | **数据来源/参考资料**。传入非空字符串时，幻灯片底部渲染灰色小字来源行；长 URL 会自动缩写为链接数量提示。格式示例：`"来源: 国家统计局 2025年数据 | https://www.stats.gov.cn"` |
| `background` | `str` | `""` | 可选的显式本地图片路径；为空时使用干净的无图布局。传入时生成器默认按 `cover` 等比铺满并允许边缘适度裁剪，所有背景自动模糊。 |

> **强制要求**：使用 search 工具获取数据后，必须在 `source` 参数中列出信息来源 URL 和机构名称。

合法 `content_type` 只以 `slide_types.md` 和 `templates/component_contracts.json` 为准。`bar_chart`、`line_chart`、`pie_chart`、`doughnut_chart`、`table` 这类别名不能写入新的任务计划，运行时也不再自动修正。

## 模板契约元数据

`templates/component_contracts.json` 集中声明组件语义、渲染归类、容量与已实现变体。规划器、Reviewer、Validator 或通用 Agent 都应优先读取这一份文件，旧 `templates/full-decks` 和 `templates/single-page` 目录不再作为运行契约。

推荐字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `contract.capacity` | object | 容量契约；`target_*_min/max` 表示正常内容密度，`max_*` 表示生成器可渲染的安全上限，另可声明 `min_items`、`max_items`、`density` |
| `contract.required_fields` | string[] | 生成器参数维度的必填字段，例如 `title`、`bullets`、`data` |
| `contract.best_for` | string[] | 最适合的内容场景 |
| `contract.avoid_for` | string[] | 应避免使用的内容场景 |
| `contract.overflow_strategy` | string | 内容超量时的推荐动作，例如 `split_slide`、`reduce_series` |
| `contract.background_policy` | string | 图片策略，例如 `image_recommended`、`image_optional` 或 `clean_default`；不要求每页都有背景图。 |
| `contract.visual_primitives` | string[] | 首选视觉 primitives，例如 `local_icons`、`shapes`、`charts`、`cards` |
| `variants` | object[] | 已有 generator 明确支持的 `layout_variant`。空数组表示 Planner 保持 `layout_variant` 为空，由组件布局引擎自适应。 |

> 注意：部分旧模板 JSON 的 UI 字段名与生成器参数名不同，例如 `chart_data` 对应 `data`、`metrics` 对应 `kpis`。`contract.required_fields` 优先按生成器参数理解，后续 selector/adapter 应负责字段映射。
> 当前阶段 `content_plan.components` 是生成器消费的数据源；不要依赖旧参数承载页面正文。`layout_variant` 只用于已实现分支，当前 `section_divider` 固定 `number_sidebar`，`image_text` 支持 `image_left`、`image_right`、`image_top_band`，其他内容页按组件密度自适配。
> 新 DeckSpec 内容契约以 `title`、`content_bank`、`sections` 和 `tasks[].content_plan.components` 为主；`task_id`、`output_file`、`status` 可由调用方或任务系统根据页码派生和维护，`theme/template/description/qa_report/fix_attempts/capacity_hint/reviewer_status` 不作为 Planner 内容字段。

## 组件规划入口

规划时先读 `templates/component_contracts.json`。常用组件不是越多越好，而是让 Planner 用更贴近业务的语义表达页面内容：

- 结构：`headline`、`subheadline`、`section_marker`、`toc_item`
- 文本：`argument_block`、`paragraph`、`text_block`、`list`、`numbered_list`、`bullet_list`、`evidence_list`
- 卡片：`feature_card`、`fact_card`、`key_point`、`insight`、`recommendation`、`risk_item`、`opportunity_item`、`case_snapshot`、`decision_item`
- 数据：`kpi_metric`、`stat`、`number_callout`、`chart`、`table`、`comparison_matrix`
- 流程：`timeline_node`、`process_step`、`milestone`
- 视觉与来源：`image`、`map`、`diagram`、`quote_block`、`callout`、`source_note`
- 原子渲染语义：`text_block`、`divider`、`icon`、`tag`、`shape`、`arrow`、`architecture_box`

组件 JSON 禁止写坐标、字号、颜色、边距、透明度和卡片尺寸。生成器会把语义组件归入稳定渲染族，并按可用内容带自适应居中、裁剪超长文本或降低字号。需要深入论述时使用 `argument_block` 写完整段落；需要列举检查项、步骤或证据时使用 `list`、`numbered_list` 或 `evidence_list`，不要把所有内容拆成泛化短卡片。

原子组件主要服务于 Python generator：Planner 可以声明“这里需要风险标签”“这些架构模块存在依赖关系”“两组信息需要分割”，但不能指定具体位置。`architecture_box` 会在深度说明、流程、内容页中优先组合为架构图；`arrow` 使用 `relation`、`target` 或 `text` 表达连接含义；`tag/icon/divider/shape` 作为辅助语义参与标题区、卡片区或架构区的自动布局。

## 图片与动态排版

- 需要图片时，在每页 `visual_intent` 中声明图片语义，并在渲染前解析为有效的本地路径；没有图片意图的页面可使用无图布局。
- 图片搜索、下载、保存、来源和署名由调用方或宿主系统负责；生成器只消费本地文件，不直接访问图片 provider，也不读取离线 `assets/manifest.json`。
- 若组件本身需要图片（如 `image_text`、`image_hero`），`asset_query` / `asset_subject` 必须在渲染前解析为真实 `local_path`；来源与署名应一并保留。
- 文字 glyph、几何形状、卡片、图表和分隔线可作为无图页面的完整视觉表达；它们也可补充有图页面的信息层级。

生成器辅助模块：

- `generators/layout_intelligence.py`：计算内容密度，提供动态字号、对齐、网格选择和内容带居中。
- `generators/base.py`：`add_text` 默认进行文本框自适配，包含行数估算、稀疏内容适度放大、溢出时缩小和自动垂直 anchor；需要显式表达时可用 `add_text_boxed`。

内容密度约定：

- `sparse`：1-3 条短内容，应使用居中焦点布局、较大字号和语义图标。
- `normal`：常规扫描布局，保持左对齐和稳定层级。
- `dense`：压缩间距和字号但不牺牲可读性；超出模板容量时拆页。

布局平衡约定：

- 标题/章节/引用等低密度页应按标题组实际高度居中，避免固定 y 坐标造成偏上或偏下。
- 目录、卡片、流程、时间线、KPI、列表等成组元素应先计算实际占用高度，再放入可用内容带。
- 页面存在 `source` 时，工作台内容区必须在底部来源分隔线上方保留安全间距；图片、卡片、图表和正文面板都不得侵入来源栏。
- 背景图片默认使用 `cover` 适配：按当前幻灯片真实宽高比等比铺满，允许边缘被适度裁剪，但底层图片锚点必须严格限制在幻灯片画布内。内部 helper 保留 `contain`，供明确要求完整显示原图时使用同图模糊扩展层补边；Planner 不控制该参数。
- 所有带背景图的页面会做轻度降饱和、降对比和极低比例柔化；正文卡片则裁切对应背景区域，做约 58px 高半径模糊后仅叠加约 1%–13% 的中性浅层。因此背景照片持续可见、正文仍有稳定对比，且大面积信息页不会形成灰色实底。标题区根据顶部背景区域自动选择高对比反相文字。标题页使用全局虚化背景和居中内容，不使用左右分栏或底部色板。调用方只传递图片路径和语义构图，不控制适配方式、模糊半径、透明度或固定色值。
- 生成器大改后必须跑全单页模板 smoke test：一页一个模板生成 PPTX，LibreOffice 转 PDF，Poppler 渲染 PNG，输出 contact sheet 和 JSON 报告。

## 生成器函数参数

### 结构引导类

#### generate_title_slide — 标题页
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"AI大模型技术概述"` |
| subtitle | str | `"从Transformer到GPT-4"` |
| author | str | `"张三"` |
| date | str | `"2025年1月"` |
| kicker | str | `"产品发布 · 2025"` (可选，标题上方小标签) |
| layout_variant | str | 当前保持为空；标题页由组件与图片语义自适应 |
| background | str | 显式本地图片路径；为空时不使用背景图，非空但无效时直接失败 |

#### generate_section_divider — 章节分隔页
| 参数 | 类型 | 示例 |
|------|------|------|
| number | str | `"01"` |
| title | str | `"技术背景"` |
| subtitle | str | `"从感知机到大模型"` |
| kicker | str | `"第三章"` (可选，编号上方小标签) |
| layout_variant | str | `"number_sidebar"`；当前实现的章节页结构 |
| background | str | 显式本地图片路径；为空时不使用背景图，非空但无效时直接失败 |

#### generate_agenda — 目录页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"目录"` |
| title | str | `"内容概览"` |
| items | `List[str]` | `["01  背景", "02  方法", "03  结论"]` |
| background | str | 显式本地图片路径；为空时不使用背景图，非空但无效时直接失败 |

目录项前缀表示章节出现顺序，不是章节分割页的绝对 `page_index`。即使章节位于第 5、8、12 页，目录仍显示 `01`、`02`、`03`。

组件计划中 `agenda` 最多 6 个组件，推荐 1 个 `insight/key_point` 说明阅读路径 + 3-5 个 `toc_item`。不要把每一页都列为独立目录项，章节过多时合并相邻主题。

### 内容陈述类

#### generate_content_slide — 普通内容页
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"深度学习发展历程"` |
| section_header | str | `"核心机制"` (可选；必须是真实小标题，禁止 `{小节标题}` 等占位符) |
| bullets | `List[str]` | `["感知机(1957)：首个线性分类器，仅能处理线性可分数据，并为后续神经网络研究提供了可训练范式", ...]` (4-6条，目标每条45-85字，可渲染上限100字) |
| kicker | str | `"要点 · 核心技术"` (可选，标题上方小标签) |
| lede | str | `"一句话概括本页核心信息，在 section_header 和 bullets 之间作为引导段落"` (可选) |
| layout_variant | str | 新规划保持为空；由 `content_plan.components` 和组件密度触发布局 |
| background | str | 显式本地图片路径；为空时不使用背景图，非空但无效时直接失败 |

#### generate_quote_slide — 金句/引言页
| 参数 | 类型 | 示例 |
|------|------|------|
| quote | str | `"弱小和无知不是生存的障碍，傲慢才是"` |
| attribution | str | `"— 刘慈欣《三体》"` |
| kicker | str | `"金句"` (可选，引言上方小标签) |
| background | str | 显式本地图片路径；为空时不使用背景图，非空但无效时直接失败 |

#### generate_image_text — 图文混排页
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"GPT-4多模态能力"` |
| layout | str | `"right-image"` 或 `"left-image"` |
| layout_variant | str | `"image_left"` / `"image_right"` / `"image_top_band"` / `"image_bottom_band"`；为空时 `render_task.py` / `render_deck.py` 或调用适配层按页序轮换 |
| image_path | str | 显式本地图片路径；图文页必填，必须由调用方提前解析为真实文件 |
| header | str | `"核心技术突破"` |
| paragraph | str | `"300-450字的自然语言段落..."` **（强制，禁止拆分为 bullets）** |
| bullets | `List[str]` | ~~（已废弃，勿用 paragraph 拆分后的 bullets）~~ |
| kicker | str | `"功能 · 核心"` (可选，标题上方小标签) |
| sub_header | str | `"能力亮点"` (可选，header 与内容之间的次级标题) |
| source | str | `"来源: 腾讯云 2025 | https://..."` (可选，数据来源标注) |
| background | str | 显式本地图片路径；为空时不使用背景图，非空但无效时直接失败 |

> **强制规则**：`paragraph` 是唯一正文来源。禁止将 paragraph 内容拆分为 bullets 后只传 bullets。paragraph 必须是240-450字的完整自然语言段落，禁止罗列要点。需要图片的页面必须传入真实本地文件；禁止自行绘制图片占位符、传入虚构路径或依赖旧 `asset:` id。
> `image_left` 为左图右文，`image_right` 为左文右图，`image_top_band` 为上方横幅图加下方正文，`image_bottom_band` 为上方正文加下方横幅图。四者都保留来源栏安全区和图片 caption 面板；正文过短时生成器会压缩文本面板高度并垂直居中，避免空白大框。

### 对比与并列类

#### generate_two_column — 双栏对比
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"CNN vs Transformer 对比"` |
| left_header | str | `"CNN"` |
| right_header | str | `"Transformer"` |
| kicker | str | `"方案对比"` (可选，标题上方小标签) |
| left_bullets | `List[str]` | `["擅长空间特征提取", ...]` (3-6条) |
| right_bullets | `List[str]` | `["擅长全局依赖建模", ...]` (3-6条) |
| left_intro | str | `"CNN是计算机视觉的基础架构..."` (可选，开篇引言段落) |
| right_intro | str | `"Transformer在NLP领域取得突破..."` (可选，开篇引言段落) |
| left_sections | `Dict[str, List[str]]` | `{"key_points": [...], "analysis": [...], "data": [...]}` (可选，多区块结构) |
| right_sections | `Dict[str, List[str]]` | 同上 (可选，多区块结构) |
| left_items | `List[dict]` | `[{"title": "...", "desc": "...", "metric": "↑ 30%"}, ...]` (可选，逐项卡片模式) |
| right_items | `List[dict]` | 同上 (可选，逐项卡片模式) |
| layout_variant | str | 新规划保持为空；对比结构优先用 `comparison_table` 或 `comparison_matrix` 组件表达 |
| background | str | 显式本地图片路径；为空时不使用背景图，非空但无效时直接失败 |

> **内容模式优先级**：优先使用 `left_sections` / `right_sections`（多区块模式），包含"核心要点"、"深度分析"、"数据支撑"等子区块；其次使用 `left_intro` + `left_bullets`（引言+要点模式）；最后才用纯 `left_bullets` / `right_bullets`。每条内容必须包含具体数字或事实，禁用模糊描述。

#### generate_card_grid — 卡片阵列
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"六大核心能力"` |
| layout | str | `"2x2"` 或 `"2x3"` 或 `"3x2"` |
| layout_variant | str | 新规划保持为空；卡片数量和主次通过组件类型、`emphasis` 表达 |
| cards | `List[dict]` | `[{"header": "智能问答", "body": "基于大模型的自然语言交互系统，支持多轮对话，并结合企业知识库提供可追溯答案..."}, ...]` ×4-6 (body 目标80-140字，可渲染上限160字) |
| kicker | str | `"能力 · 核心模块"` (可选，标题上方小标签) |
| subtitle | str | `"全方位赋能企业数字化转型"` (可选，标题下方副标题) |
| background | str | 显式本地图片路径；为空时不使用背景图，非空但无效时直接失败 |

### 流程与关系类

#### generate_timeline — 时间轴
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"AI发展里程碑"` |
| direction | str | `"horizontal"` 或 `"vertical"` |
| nodes | `List[dict]` | `[{"year": "2017", "event": "Transformer论文发表", "icon": "01"}, ...]` ×4-6 |
| kicker | str | `"技术演进"` (可选，标题上方小标签) |
| subtitle | str | `"从深度学习到大模型的时代跨越"` (可选，标题下方副标题) |
| background | str | 显式本地图片路径；为空时不使用背景图，非空但无效时直接失败 |

### 数据与指标类

#### generate_kpi_dashboard — 指标看板（固定 2x2 布局，最多 4 个 KPI）
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"数据 · 季度增长"` |
| title | str | `"核心业务指标"` |
| kpis | `List[dict]` | `[{"value": "1248K", "label": "月活用户", "delta": "↑38% YoY", "baseline": "去年902K"}, ...]` ×4（固定 2x2 网格，最多 4 个） |
| subtitle | str | `"业务线关键绩效数据"` (可选，标题下方副标题) |
| background | str | 显式本地图片路径；为空时不使用背景图，非空但无效时直接失败 |

> **强制规则**：每个 KPI 字典必须包含全部 4 个字段：`value`（具体数值+单位）、`label`（效果说明）、`delta`（变化趋势，如 ↑38%）、`baseline`（对比基准，如 vs 传统方案）。禁止使用占位符如 `"{数值}"`。数据必须真实（通过 search 获取）。

### 数据与可视化类

#### generate_chart_slide — 图表专页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"数据分析"` |
| title | str | `"季度营收对比"` |
| subtitle | str | `"2025年各季度数据"` (可选) |
| chart_type | str | `"bar"`, `"pie"`, `"line"`, `"doughnut"`, `"stacked_bar"` |
| data | `Dict` | `{"labels": ["Q1","Q2","Q3"], "datasets": [{"name": "2025", "values": [100,200,300]}]}` |
| show_legend | bool | `True` (是否显示图例) |
| background | str | 显式本地图片路径；为空时不使用背景图，非空但无效时直接失败 |

#### generate_swot_analysis — SWOT分析页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"战略分析"` |
| title | str | `"AI产品战略SWOT分析"` |
| subtitle | str | `"基于市场与竞争格局"` (可选) |
| swot | `Dict` | 包含 strengths/weaknesses/opportunities/threats，每个有 label/items |
| background | str | 显式本地图片路径；为空时不使用背景图，非空但无效时直接失败 |

#### generate_comparison_table — 对比表格页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"选型对比"` |
| title | str | `"AI平台选型对比"` |
| subtitle | str | `"三大云厂商AI能力对比"` (可选) |
| headers | `List[str]` | `["对比维度", "方案A", "方案B", "方案C"]` |
| rows | `List[List[str]]` | `[["功能丰富度", "★★★☆☆", "★★★★☆"], ...]` |
| recommendation | str | `"综合考虑，建议选择 Azure ML"` (可选) |
| background | str | 显式本地图片路径；为空时不使用背景图，非空但无效时直接失败 |

#### generate_image_hero — 视觉冲击页
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"震撼标题"` |
| subtitle | str | `"副标题说明"` (可选) |
| overlay_color | str | `"primary"` (可选，颜色主题) |
| background | str | 显式本地图片路径；为空时不使用背景图，非空但无效时直接失败 |

#### generate_kanban — 看板进度页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"项目进度"` |
| title | str | `"迭代看板"` |
| subtitle | str | `"版本规划"` (可选) |
| columns | `List[dict]` | `[{"title": "待办", "cards": [{"title": "任务1", "priority": "high"}, ...]}, ...]` |
| progress | int | `65` (整体进度百分比 0-100) |
| background | str | 显式本地图片路径；为空时不使用背景图，非空但无效时直接失败 |

#### generate_brand_focus — 品牌价值聚焦页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"品牌战略"` |
| title | str | `"核心价值观"` |
| subtitle | str | `"品牌体系"` (可选) |
| center_text | str | `"核心\n理念"` (中心圆文字) |
| surrounding_points | `List[dict]` | `[{"title": "创新", "desc": "持续创新驱动发展"}, ...]` (围绕中心的点) |
| principles | `List[dict]` | `[{"title": "原则1", "desc": "描述"}, ...]` (右侧面板内容) |
| background | str | 显式本地图片路径；为空时不使用背景图，非空但无效时直接失败 |

#### generate_region_map — 区域版图页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"市场布局"` |
| title | str | `"全球业务版图"` |
| subtitle | str | `"区域覆盖"` (可选) |
| regions | `List[dict]` | `[{"label": "华东", "fill": "primary", "active": true}, ...]` (地图区域) |
| regions_detail | `List[dict]` | `[{"title": "华东", "metrics": [{"label": "营收", "value": "12亿"}]}, ...]` (右侧详情) |
| background | str | 显式本地图片路径；为空时不使用背景图，非空但无效时直接失败 |

## 常见错误

| 错误 | 原因 | 修复 |
|------|------|------|
| `save_slide` AttributeError | slide 对象无效（前一步 generate 失败未检查） | 检查 generate 返回值，每个文件独立 new_presentation |
| `No module named 'generators'` | sys.path 未指向 `ppt-deck-planner` 目录 | 确认 `sys.path` 包含 `<...>/ppt-deck-planner`，或通过 `render_task.py --skills-dir <包含 ppt-deck-planner 的目录>` 调用 |

## 变更纪律

- 常规页面生成必须复用 `skills/ppt-deck-planner/generators` 包，不在 Agent 生成代码中手写底层 `python-pptx` 绘制逻辑。
- 当需求明确涉及视觉质量、背景、容量或布局能力时，可以修改对应 generator 和 `base.py` helper，但必须同步本文件、模板契约和聚焦 smoke 验证。
- 修改导出函数签名时，同步 `generators/__init__.py`、`generators/render_task.py`、`generators/render_deck.py`、调用适配层和相关模板 JSON；仅修改内部实现时不需要改导出表。


