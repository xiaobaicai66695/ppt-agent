# Generators 参考文档

本文档是 SlideExecutor 生成 PPT 时的生成器函数参考。当前生成器已切换为组件式布局：各 `generate_xxx` 函数保留兼容入口，但内部统一把参数和 `content_plan.components` 转成语义组件，再由 `generators/component_layout.py` 决定页面结构。

## 导入方式

```python
import sys
from pathlib import Path

script_dir = Path("<工作目录绝对路径>")  # 必须是绝对路径，不要用 __file__ 或 os.getcwd()
generators_pkg_dir = (script_dir / ".." / ".." / "skills" / "visual_designer").resolve()
sys.path.insert(0, str(generators_pkg_dir))

from generators import (
    new_presentation, save_presentation, save_slide,
    generate_title_slide, generate_section_divider, generate_content_slide,
    generate_stat_slide, generate_quote_slide, generate_card_grid,
    generate_timeline, generate_process_flow, generate_two_column,
    generate_three_column, generate_summary_slide, generate_image_text,
    generate_example_detail, generate_deep_dive, generate_agenda,
    generate_case_study, generate_kpi_dashboard, generate_chart_slide,
    generate_icon_grid, generate_swot_analysis, generate_comparison_table,
    generate_image_hero, generate_kanban, generate_brand_focus, generate_region_map,
)
```

## 调用规范（必须严格遵守）

- `new_presentation(palette="xxx")` — 创建空演示文稿（不含幻灯片），每个 PPTX 文件必须单独创建
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
| `palette` | `str` | `"ocean_soft"` | 配色方案名，见 palettes.md |
| `source` | `str` | `""` | **数据来源/参考资料**。传入非空字符串时，幻灯片底部渲染灰色小字来源行；长 URL 会自动缩写为链接数量提示。格式示例：`"来源: 国家统计局 2025年数据 | https://www.stats.gov.cn"` |
| `background` | `str` | `""` | 背景图片配置。可传主题名（如 `"party_government"`），也可传主题内具体图片引用（如 `"party_government/images/3.jpg"`）。留空则用纯色背景。可取值见 `background_templates/SKILL.md` |

> **强制要求**：使用 search 工具获取数据后，必须在 `source` 参数中列出信息来源 URL 和机构名称。

合法 `content_type` 只以 `slide_types.md` 和 `templates/component_contracts.json` 为准；`templates/single-page/*.json` 继续作为前端和 loader 兼容元数据。`bar_chart`、`line_chart`、`pie_chart`、`doughnut_chart`、`table` 这类历史别名只能在旧任务生成时兼容映射，不能写入新的任务计划。

## 模板契约元数据

`templates/component_contracts.json` 集中声明组件语义、渲染归类、容量与已实现变体。规划器优先读取这一份文件，减少多次 read_file；`templates/single-page/*.json` 可以声明可选的 `contract` 对象用于 UI 展示和兼容，不改变现有 generator 调用方式。

推荐字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `contract.capacity` | object | 容量契约；`target_*_min/max` 表示正常内容密度，`max_*` 表示生成器可渲染的安全上限，另可声明 `min_items`、`max_items`、`density` |
| `contract.required_fields` | string[] | 生成器参数维度的必填字段，例如 `title`、`bullets`、`data` |
| `contract.best_for` | string[] | 最适合的内容场景 |
| `contract.avoid_for` | string[] | 应避免使用的内容场景 |
| `contract.overflow_strategy` | string | 内容超量时的推荐动作，例如 `split_slide`、`reduce_series` |
| `contract.background_policy` | string | 背景策略：`image_recommended`、`image_optional`、`clean_default` |
| `contract.visual_primitives` | string[] | 首选视觉 primitives，例如 `local_icons`、`shapes`、`charts`、`cards` |
| `variants` | object[] | 已有 generator 明确支持的 `layout_variant`。空数组表示 Planner 保持 `layout_variant` 为空，由组件布局引擎自适应。 |

> 注意：部分旧模板 JSON 的 UI 字段名与生成器参数名不同，例如 `chart_data` 对应 `data`、`metrics` 对应 `kpis`。`contract.required_fields` 优先按生成器参数理解，后续 selector/adapter 应负责字段映射。
> 当前阶段 `content_plan.components` 是生成器优先消费的数据源；components 非空时，旧参数只作为兼容输入保留。`layout_variant` 只用于已实现分支，当前 `section_divider` 固定 `number_sidebar`，其他内容页按组件密度自适配。

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

## 背景与素材策略

- 启用背景时，整套 PPT 每一页都应使用同一个背景主题目录下的图片，避免跨主题混用造成风格漂移。
- 任务规划阶段会优先写入主题内具体图片引用，以随机选择图片并避免相邻页重复同一张图；若只收到主题名，生成器仍会解析该主题下的图片。
- 信息密集页、图表页、KPI 页和表格页也保留同主题背景，由生成器通过局部信息面板、浅色磨砂蒙版和对比度控制保证可读性。
- 所有图片背景都会在生成器内叠加浅色磨砂玻璃蒙版；不要在 Agent 代码里额外压暗图片、改写 palette 或手工叠遮罩。
- 当前阶段优先使用本地图标、文字 glyph、几何形状、卡片、图表、分隔线和浅色块作为视觉元素。
- 不默认使用外部图片搜索或生成式图片资产；确需真实照片、产品图、人物图时，应作为单独需求确认来源、授权和成本。

## 本地素材库和动态排版

本地素材库位于 `skills/visual_designer/assets/`，由 v2 `assets/manifest.json` 统一登记；完整维护说明见 `skills/visual_designer/README.md`。

| 类型 | 目录 | 用途 |
|------|------|------|
| `icon` | `assets/icons/core/` | 统一 Icons8 线性图标，用于卡片、图标网格、结构页和少字内容页 |
| `background` | `assets/backgrounds/editorial/` | 标题页、章节页、金句页、总结页的照片型 editorial 背景 |
| `pattern` | `assets/patterns/subtle/` | 信息页轻量纹理，已归一化为 1920x1080 PNG |

`assets/manifest.json` 的外部素材必须包含 `source_id`、`source_url`、`download_url`、`license`、`attribution` 和 `dimensions`。图标语义来自 `keywords`/`tags`，最长关键词和 priority 决定匹配结果；未知语义默认返回空字符串，不渲染无关图标。

主题照片独立登记在 `background_templates/manifest.json`。生成器按该文件读取主题场景、推荐 palette 和图片列表；主题 id 与 `<theme>/images/<file>` 具体路径继续兼容。

生成器辅助模块：

- `generators/asset_manager.py`：通过 manifest 解析素材，按 id/tag/keyword 查找图标和背景，并提供 `validate_manifest()`。
- `generators/background_manager.py`：通过主题 manifest 选择背景和推荐 palette，并提供 `validate_background_manifest()`；仅在 manifest 缺失时回退到旧目录扫描。
- `generators/layout_intelligence.py`：计算内容密度，提供动态字号、对齐、网格选择和内容带居中。
- `generators/base.py`：`add_text` 默认进行文本框自适配，包含行数估算、稀疏内容适度放大、溢出时缩小和自动垂直 anchor；需要显式表达时可用 `add_text_boxed`。

内容密度约定：

- `sparse`：1-3 条短内容，应使用居中焦点布局、较大字号和语义图标。
- `normal`：常规扫描布局，保持左对齐和稳定层级。
- `dense`：压缩间距和字号但不牺牲可读性；超出模板容量时拆页。

布局平衡约定：

- 标题/章节/引用等低密度页应按标题组实际高度居中，避免固定 y 坐标造成偏上或偏下。
- 目录、卡片、流程、时间线、KPI、列表等成组元素应先计算实际占用高度，再放入可用内容带。
- 没有真实图片输入时，图文页应渲染本地素材摘要面板，不得暴露 `[图片占位]` 一类占位文案。
- 生成器大改后必须跑全单页模板 smoke test：一页一个模板生成 PPTX，LibreOffice 转 PDF，Poppler 渲染 PNG，输出 contact sheet 和 JSON 报告。

素材维护与验证：

```powershell
python skills/visual_designer/scripts/sync_external_assets.py
python -m unittest discover -s skills/visual_designer/tests -p "test_asset_library.py"
```

Icons8 图标保留 `Icons by Icons8` attribution；照片和 pattern 的来源及许可随 manifest 提交。Bing 图片搜索只可作为发现入口，不能替代原始页面和许可记录。

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
| layout_variant | str | 新规划保持为空；旧任务可传入但当前不触发独立渲染分支 |
| background | str | `"artistic"` (可选，背景图片主题，为空则用纯色) |

#### generate_section_divider — 章节分隔页
| 参数 | 类型 | 示例 |
|------|------|------|
| number | str | `"01"` |
| title | str | `"技术背景"` |
| subtitle | str | `"从感知机到大模型"` |
| kicker | str | `"第三章"` (可选，编号上方小标签) |
| layout_variant | str | `"number_sidebar"`；当前实现的章节页结构 |
| background | str | `"artistic"` (可选，背景图片主题) |

#### generate_agenda — 目录页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"目录"` |
| title | str | `"内容概览"` |
| items | `List[str]` | `["01  背景", "02  方法", "03  结论"]` |
| background | str | `"artistic"` (可选，背景图片主题) |

#### generate_summary_slide — 总结页
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"总结"` |
| key_points | `List[str]` | `["01  核心结论1", "02  核心结论2"]` (最多4条) |
| thank_you | str | `"感谢聆听"` |
| contact | str | `"{联系方式}"` |
| kicker | str | `"总结"` (可选，标题上方小标签) |
| background | str | `"artistic"` (可选，背景图片主题) |

### 内容陈述类

#### generate_content_slide — 普通内容页（兜底类型）
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"深度学习发展历程"` |
| section_header | str | `"核心机制"` (可选；必须是真实小标题，禁止 `{小节标题}` 等占位符) |
| bullets | `List[str]` | `["感知机(1957)：首个线性分类器，仅能处理线性可分数据，并为后续神经网络研究提供了可训练范式", ...]` (4-6条，目标每条45-85字，可渲染上限100字) |
| kicker | str | `"要点 · 核心技术"` (可选，标题上方小标签) |
| lede | str | `"一句话概括本页核心信息，在 section_header 和 bullets 之间作为引导段落"` (可选) |
| layout_variant | str | 新规划保持为空；由 `content_plan.components` 和组件密度触发布局 |
| background | str | `"minimalist_blue/images/1.jpg"` (可选，内容要点页可使用背景，生成器会加局部玻璃面板保证正文可读) |

#### generate_quote_slide — 金句/引言页
| 参数 | 类型 | 示例 |
|------|------|------|
| quote | str | `"弱小和无知不是生存的障碍，傲慢才是"` |
| attribution | str | `"— 刘慈欣《三体》"` |
| kicker | str | `"金句"` (可选，引言上方小标签) |
| background | str | `"artistic"` (可选，背景图片主题) |

#### generate_image_text — 图文混排页
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"GPT-4多模态能力"` |
| layout | str | `"right-image"` 或 `"left-image"` |
| layout_variant | str | 新规划保持为空；旧任务可传入但当前以组件和图片语义自适应 |
| image_path | str | `"asset:photo_technology_device"`、注册 photo id 或本地文件路径；为空时自动选择语义默认图 |
| header | str | `"核心技术突破"` |
| paragraph | str | `"300-450字的自然语言段落..."` **（强制，禁止拆分为 bullets）** |
| bullets | `List[str]` | ~~（已废弃，勿用 paragraph 拆分后的 bullets）~~ |
| kicker | str | `"功能 · 核心"` (可选，标题上方小标签) |
| sub_header | str | `"能力亮点"` (可选，header 与内容之间的次级标题) |
| source | str | `"来源: 腾讯云 2025 | https://..."` (可选，数据来源标注) |
| background | str | `"artistic"` (可选，背景图片主题) |

> **强制规则**：`paragraph` 是唯一正文来源。禁止将 paragraph 内容拆分为 bullets 后只传 bullets。paragraph 必须是300-450字的完整自然语言段落，禁止罗列要点。`image_path` 缺失或无效时，生成器会从 `photo` 素材中选择可替换的真实图片；禁止自行绘制图片占位符或传入虚构路径。

### 对比与并列类

#### generate_two_column — 双栏对比
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"CNN vs Transformer 对比"` |
| left_header | str | `"CNN"` |
| right_header | str | `"Transformer"` |
| kicker | str | `"方案对比"` (可选，标题上方小标签) |
| left_bullets | `List[str]` | `["擅长空间特征提取", ...]` (3-6条，向后兼容) |
| right_bullets | `List[str]` | `["擅长全局依赖建模", ...]` (3-6条，向后兼容) |
| left_intro | str | `"CNN是计算机视觉的基础架构..."` (可选，开篇引言段落) |
| right_intro | str | `"Transformer在NLP领域取得突破..."` (可选，开篇引言段落) |
| left_sections | `Dict[str, List[str]]` | `{"key_points": [...], "analysis": [...], "data": [...]}` (可选，多区块结构) |
| right_sections | `Dict[str, List[str]]` | 同上 (可选，多区块结构) |
| left_items | `List[dict]` | `[{"title": "...", "desc": "...", "metric": "↑ 30%"}, ...]` (可选，逐项卡片模式) |
| right_items | `List[dict]` | 同上 (可选，逐项卡片模式) |
| layout_variant | str | 新规划保持为空；对比结构优先用 `comparison_table` 或 `comparison_matrix` 组件表达 |
| background | str | `"artistic"` (可选，背景图片主题) |

> **内容模式优先级**：优先使用 `left_sections` / `right_sections`（多区块模式），包含"核心要点"、"深度分析"、"数据支撑"等子区块；其次使用 `left_intro` + `left_bullets`（引言+要点模式）；最后才用纯 `left_bullets` / `right_bullets`。每条内容必须包含具体数字或事实，禁用模糊描述。

#### generate_three_column — 三栏并列
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"三种方案对比"` |
| columns | `List[dict]` | `[{"header": "方案A", "bullets": ["优点1", "优点2"]}, ...]` ×3 |
| kicker | str | `"能力矩阵"` (可选，标题上方小标签) |
| background | str | `"artistic"` (可选，背景图片主题) |

#### generate_card_grid — 卡片阵列
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"六大核心能力"` |
| layout | str | `"2x2"` 或 `"2x3"` 或 `"3x2"` |
| layout_variant | str | 新规划保持为空；卡片数量和主次通过组件类型、`emphasis` 表达 |
| cards | `List[dict]` | `[{"header": "智能问答", "body": "基于大模型的自然语言交互系统，支持多轮对话，并结合企业知识库提供可追溯答案..."}, ...]` ×4-6 (body 目标80-140字，可渲染上限160字) |
| kicker | str | `"能力 · 核心模块"` (可选，标题上方小标签) |
| subtitle | str | `"全方位赋能企业数字化转型"` (可选，标题下方副标题) |
| background | str | `"artistic"` (可选，背景图片主题) |

### 流程与关系类

#### generate_timeline — 时间轴
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"AI发展里程碑"` |
| direction | str | `"horizontal"` 或 `"vertical"` |
| nodes | `List[dict]` | `[{"year": "2017", "event": "Transformer论文发表", "icon": "01"}, ...]` ×4-6 |
| kicker | str | `"技术演进"` (可选，标题上方小标签) |
| subtitle | str | `"从深度学习到大模型的时代跨越"` (可选，标题下方副标题) |
| background | str | `"artistic"` (可选，背景图片主题) |

#### generate_process_flow — 步骤流程图
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"模型训练流程"` |
| direction | str | `"horizontal"` / `"horizontal_zigzag"` / `"vertical"` |
| steps | `List[dict]` | `[{"num": "01", "title": "数据收集", "desc": "采集多源数据"}, ...]` ×3-6 |
| kicker | str | `"工程实践"` (可选，标题上方小标签) |
| subtitle | str | `"端到端自动化训练流水线"` (可选，标题下方副标题) |
| background | str | `"artistic"` (可选，背景图片主题) |

### 数据与指标类

#### generate_stat_slide — 关键数字页
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"系统性能指标"` |
| stats | `List[dict]` | `[{"number": "99.99", "unit": "%", "label": "系统可用性"}, ...]` ×2-4 |
| kicker | str | `"年度成果"` (可选，标题上方小标签) |
| subtitle | str | `"2025财年关键数据一览"` (可选，标题下方副标题) |
| background | str | `"artistic"` (可选，背景图片主题) |

#### generate_kpi_dashboard — 指标看板（固定 2x2 布局，最多 4 个 KPI）
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"数据 · 季度增长"` |
| title | str | `"核心业务指标"` |
| kpis | `List[dict]` | `[{"value": "1248K", "label": "月活用户", "delta": "↑38% YoY", "baseline": "去年902K"}, ...]` ×4（固定 2x2 网格，最多 4 个） |
| subtitle | str | `"业务线关键绩效数据"` (可选，标题下方副标题) |
| background | str | `"artistic"` (可选，背景图片主题) |

> **强制规则**：每个 KPI 字典必须包含全部 4 个字段：`value`（具体数值+单位）、`label`（效果说明）、`delta`（变化趋势，如 ↑38%）、`baseline`（对比基准，如 vs 传统方案）。禁止使用占位符如 `"{数值}"`。数据必须真实（通过 search 获取）。

### 内容叙事类（案例/详解）

#### generate_example_detail — 实例详解页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"实例 · 金融风控"` |
| title | str | `"蚂蚁AlphaRisk：实时风控系统"` |
| lede | str | `"日均处理数亿笔交易，风险识别准确率99.99%"` |
| context_block | str | `"金融欺诈每年造成数百亿损失..."` (1-2句背景) |
| solution_block | str | `"基于深度图学习的实时检测..."` (2-3句方案) |
| metrics | `List[dict]` | `[{"value": "99.99%", "label": "准确率", "trend": "↑"}, ...]` ×3 |
| takeaway | str | `"图学习是风控的核心技术方向"` |
| background | str | `"artistic"` (可选，背景图片主题) |

#### generate_deep_dive — 深入详解页（双栏）
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"详解 · Transformer架构"` |
| title | str | `"自注意力机制原理"` |
| lede | str | `"一句话概括核心价值"` |
| left_header | str | `"核心要点"` |
| key_points | `List[str]` | `["多头注意力：16个子空间并行建模", ...]` (3-5条) |
| analysis | `List[str]` | `["维度分析1：结论", "维度分析2：结论"]` (2条) |
| right_header | str | `"案例/数据"` |
| case_example | `List[str]` | `["GPT-4：万亿参数，MMLU 86.4%", ...]` (3-4条) |
| data_evidence | `List[str]` | `["推理延迟：320ms→18ms", "训练成本：$63M", ...]` (3条) |
| supplement | `List[str]` | 可选补充信息 (0-2条) |
| background | str | `"artistic"` (可选，背景图片主题) |

#### generate_case_study — 案例研究页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"案例 · 智能客服"` |
| title | str | `"某银行AI客服系统"` |
| context | str | `"银行业客服成本高、响应慢..."` (背景) |
| problem | str | `"日均10万+咨询，人工应答率仅60%"` (痛点) |
| solution | str | `"基于RAG+大模型的智能问答..."` (方案) |
| results | `List[dict]` | `[{"metric": "应答率", "value": "95%", "comparison": "提升35%"}, ...]` ×4 |
| background | str | `"artistic"` (可选，背景图片主题) |

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
| background | str | `"artistic"` (可选，背景图片主题) |

#### generate_icon_grid — 图标网格页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"核心能力"` |
| title | str | `"六大技术支柱"` |
| subtitle | str | `"构建完整AI技术体系"` (可选) |
| layout | str | `"3x2"` 或 `"3x3"` 或 `"2x3"` |
| icons | `List[dict]` | `[{"icon": "研", "label": "基础研究", "color": "primary"}, ...]` |
| background | str | `"artistic"` (可选，背景图片主题) |

#### generate_swot_analysis — SWOT分析页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"战略分析"` |
| title | str | `"AI产品战略SWOT分析"` |
| subtitle | str | `"基于市场与竞争格局"` (可选) |
| swot | `Dict` | 包含 strengths/weaknesses/opportunities/threats，每个有 label/items |
| background | str | `"artistic"` (可选，背景图片主题) |

#### generate_comparison_table — 对比表格页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"选型对比"` |
| title | str | `"AI平台选型对比"` |
| subtitle | str | `"三大云厂商AI能力对比"` (可选) |
| headers | `List[str]` | `["对比维度", "方案A", "方案B", "方案C"]` |
| rows | `List[List[str]]` | `[["功能丰富度", "★★★☆☆", "★★★★☆"], ...]` |
| recommendation | str | `"综合考虑，建议选择 Azure ML"` (可选) |
| background | str | `"artistic"` (可选，背景图片主题) |

#### generate_image_hero — 视觉冲击页
| 参数 | 类型 | 示例 |
|------|------|------|
| title | str | `"震撼标题"` |
| subtitle | str | `"副标题说明"` (可选) |
| description | str | `"描述文字"` (可选) |
| overlay_color | str | `"primary"` (可选，颜色主题) |
| background | str | `"artistic"` (可选，背景图片主题) |

#### generate_kanban — 看板进度页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"项目进度"` |
| title | str | `"迭代看板"` |
| subtitle | str | `"版本规划"` (可选) |
| columns | `List[dict]` | `[{"title": "待办", "cards": [{"title": "任务1", "priority": "high"}, ...]}, ...]` |
| progress | int | `65` (整体进度百分比 0-100) |
| background | str | `"artistic"` (可选，背景图片主题) |

#### generate_brand_focus — 品牌价值聚焦页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"品牌战略"` |
| title | str | `"核心价值观"` |
| subtitle | str | `"品牌体系"` (可选) |
| center_text | str | `"核心\n理念"` (中心圆文字) |
| surrounding_points | `List[dict]` | `[{"title": "创新", "desc": "持续创新驱动发展"}, ...]` (围绕中心的点) |
| principles | `List[dict]` | `[{"title": "原则1", "desc": "描述"}, ...]` (右侧面板内容) |
| background | str | `"artistic"` (可选，背景图片主题) |

#### generate_region_map — 区域版图页
| 参数 | 类型 | 示例 |
|------|------|------|
| kicker | str | `"市场布局"` |
| title | str | `"全球业务版图"` |
| subtitle | str | `"区域覆盖"` (可选) |
| regions | `List[dict]` | `[{"label": "华东", "fill": "primary", "active": true}, ...]` (地图区域) |
| regions_detail | `List[dict]` | `[{"title": "华东", "metrics": [{"label": "营收", "value": "12亿"}]}, ...]` (右侧详情) |
| background | str | `"artistic"` (可选，背景图片主题) |

## 常见错误

| 错误 | 原因 | 修复 |
|------|------|------|
| `save_slide` AttributeError | slide 对象无效（前一步 generate 失败未检查） | 检查 generate 返回值，每个文件独立 new_presentation |
| `No module named 'generators'` | sys.path 指向了 skills/ 而非 skills/visual_designer | 确认路径：`script_dir / ".." / ".." / "skills" / "visual_designer"` |

## 变更纪律

- 常规页面生成必须复用 `skills/visual_designer/generators` 包，不在 Agent 生成代码中手写底层 `python-pptx` 绘制逻辑。
- 当需求明确涉及视觉质量、背景、容量或布局能力时，可以修改对应 generator 和 `base.py` helper，但必须同步本文件、模板契约和聚焦 smoke 验证。
- 修改导出函数签名时，同步 `generators/__init__.py`、SlideExecutor prompt 和相关模板 JSON；仅修改内部实现时不需要改导出表。
