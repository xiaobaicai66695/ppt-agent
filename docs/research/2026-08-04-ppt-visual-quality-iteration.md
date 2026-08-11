# PPT 视觉质量迭代预研：多版式变体与主题背景

## 背景

用户对比了当前系统生成的大兴安岭 PPT 与豆包生成结果。豆包样例观感明显更好，用户判断主要来自两点：

1. 同一种页面类型有多种可套用模板，整套 PPT 的视觉节奏更丰富。
2. 使用了与主题匹配的背景图片，尤其封面和内容页局部图像让主题感更强。

本预研目标是判断当前 PPT Agent 下一轮应如何迭代，以降低“AI 感”和模板填充感。

## 样例观察

豆包样例的第一屏优势很明显：

- 封面是全幅森林山谷/河流照片，主题“大兴安岭”在首屏即成立。
- 文字层级简单：主标题、副标题、三个定位短语，信息不拥挤。
- 背景图有暗化/虚化/渐变蒙层，文字可读，画面仍有真实空间感。
- 左侧缩略图显示同一章节下页面存在多种排法：KPI 卡片、图文页、三图卡片、左右图文、生态照片组等。
- 整体色彩围绕森林绿、浅绿、米白展开，和地域生态主题一致。

相比之下，当前系统生成结果的 AI 感主要来自：

- 每个 `content_type` 基本只有一个 generator 布局，重复度高。
- 背景资源是泛主题，如 `snowy_mountain`、`ink_wash_mountain`、`minimalist_blue`，不能精准覆盖“大兴安岭/森林/河谷/生态旅游”等语义。
- `content_plan` 仍偏通用，无法表达“这一页应该用地图、照片、路线、资源分布、生态剖面”等视觉意图。
- 生成器更擅长画卡片/色块/文字，而不是将真实主题图片作为叙事材料。
- 章节页和内容页的视觉节奏缺少变体，容易形成“统一模板批量套壳”的观感。

## 当前能力盘点

### 已具备能力

- `templates/single-page/*.json` 已有 24 个合法页面类型。
- 每个模板 JSON 已有 `contract.capacity`、`best_for`、`avoid_for`、`background_policy` 等元数据。
- 生成器支持 `background` 参数，`background_manager.py` 可按主题选择本地背景。
- `assets/manifest.json` 有本地图标、editorial 背景、subtle pattern。
- 主 Agent 已能在 manifest 中传递 `background`，SlideExecutor 要求原样传入 generator。
- 前端和后端已有模板/背景接口，可展示可用背景。

### 主要缺口

- 缺少 `layout_variant` 概念：同一个 `content_type` 只有一种主要排法。
- 缺少主题图片资产层：当前背景多为泛主题图，不足以支撑地域、行业、人物、产品类主题。
- 缺少视觉素材规划字段：`content_plan` 不能稳定表达 image role、image query、crop、caption、where to place。
- 缺少 deck-level 视觉节奏规则：没有控制“连续 3 页不能同构”“章节开头必须强主题图”“内容页图文比例”。
- 缺少本地视觉 QA 指标：目前主要看文件交付，缺少对重复度、背景命中、占位残留、图文比例的自动检查。

## 公开产品/研究方案实现路径

本轮补充调研了 Gamma、Beautiful.ai、Canva、Microsoft Copilot in PowerPoint、Plus AI、Presentations.AI，以及论文/开源项目 PPTAgent。公开资料通常不会披露完整内部实现，但它们的产品能力和论文流程能清晰反推出一条行业共性路线：**LLM 负责理解、规划和局部改写；版面质量主要来自模板/智能布局/品牌约束/素材库/渲染 QA，而不是让 LLM 从零摆元素。**

| 方案 | 公开资料体现的实现路径 | 对当前系统的借鉴 |
| --- | --- | --- |
| Gamma | 从 prompt 或文档生成 presentation/document，使用 themes、templates、smart templates/blocks，支持导出 PPT/PDF，并允许后续编辑和主题切换。Gamma 的模板页也按 Company、Consulting、Education、Marketing、Sales、Strategy 等场景组织。资料：[Gamma](https://gamma.app/)、[Gamma templates](https://gamma.app/templates)、[Gamma presentations](https://gamma.app/products/presentations) | 借鉴“内容与视觉主题分离”：主 Agent 先生成 deck brief / style plan，再由模板块和主题系统渲染。当前不宜让 prompt 直接写底层 python-pptx，应把 `layout_variant`、`visual_theme`、`asset_role` 作为中间层。 |
| Beautiful.ai | 核心卖点是 Smart Slides：编辑内容时自动对齐、缩放和重排；DesignerBot/Create with AI 先让用户确认 outline，再生成设计；支持 brand controls、slide-level AI panel、替代版本、图片库和 PowerPoint add-in。资料：[Beautiful.ai](https://www.beautiful.ai/)、[AI presentation maker](https://www.beautiful.ai/presentation-maker)、[Smart Slides/templates](https://www.beautiful.ai/template-category/popular)、[DesignerBot](https://www.beautiful.ai/blog/introducing-designerbot-ai-presentations)、[Create with AI workflow](https://www.beautiful.ai/blog/introducing-the-create-with-ai-workflow) | 借鉴“智能布局而非静态模板”：generator 应按内容容量自动增删卡片、调整字号/列数/图文比例；Fixer 更适合修 generator 参数或变体选择，而不是直接手改最终 PPT。 |
| Canva Magic Design | Magic Design for Presentations 可根据 idea、outline、slides 或 content 生成 customized presentation templates；生成后仍在 Canva 编辑器中继续应用品牌、素材和 AI 工具。资料：[Canva AI presentations](https://www.canva.com/create/ai-presentations/)、[Magic Design](https://www.canva.com/magic-design/)、[Canva Help: using Magic Presentations](https://www.canva.com/help/using-magic-presentations/) | 借鉴“模板库 + 素材库 + 可编辑工作台”：不要只增加几张背景图，应建设可检索的主题照片、图标、色板和页面模板组合，并在前端暴露更清晰的模板/主题选择。 |
| Microsoft Copilot in PowerPoint | 支持从 Word/file/prompt 创建演示文稿；官方建议 Word Styles 帮助 Copilot 理解结构，会尝试使用文档中的相关图片；可从文件加单页、用 Copilot 编辑、保持品牌模板/Brand Kit，并可插入 stock/brand/AI-generated images。资料：[Create a new presentation with Copilot](https://support.microsoft.com/en-us/powerpoint/copilot/create-a-new-presentation-with-copilot-in-powerpoint)、[Prepare your presentation](https://support.microsoft.com/en-us/microsoft-365-copilot/prepare-your-presentation-with-microsoft-365-copilot)、[Keep your presentation on-brand](https://support.microsoft.com/en-us/powerpoint/copilot/keep-your-presentation-on-brand-with-copilot)、[Add an image](https://support.microsoft.com/en-us/powerpoint/copilot/add-an-image-to-your-presentation-with-copilot-in-powerpoint) | 借鉴“结构化输入优先”：如果用户给 outline/文档，先保留标题层级、图片、来源和章节边界，再规划页面。图片应进入 `visual_assets`，不是事后随便补背景。 |
| Plus AI | 作为 PowerPoint/Google Slides add-in 工作；从 prompt 或上传文件生成大纲和可编辑 PPT/Slides；支持 Insert、Rewrite、Remix、Custom instructions、Preset library，能添加 images/icons/charts/tables，提供 hundreds of layouts。资料：[Plus AI](https://plusai.com/)、[AI PowerPoint maker](https://plusai.com/ai-powerpoint-maker/)、[AI for Google Slides](https://plusai.com/google-slides-ai/)、[Remix](https://plusai.com/features/reformat-slides-with-remix/) | 借鉴“局部 Remix 能力”：当前可以先实现 slide-level regenerate/variant switch，而不是每次整套重跑。后续前端可让用户对某页选择“换版式/换图片/压缩文字/改成图表”。 |
| Presentations.AI | 从 topic、URL、document 生成 on-brand deck，包含 slide copy、layouts、PPTX export；有品牌 colors/logos/fonts 自动应用和 500+ slide templates。资料：[Presentations.AI](https://www.presentations.ai/)、[AI presentation maker](https://www.presentations.ai/ai-presentation-maker)、[Slide templates](https://www.presentations.ai/slide-templates)、[Brand customization FAQ](https://www.presentations.ai/presentation-templates/case-study-presentation) | 借鉴“品牌/主题约束贯穿全局”：`deck_style_plan` 应包含色板、字体、图片风格、图标风格和页面节奏，不能只在单页 generator 内临时决定。 |
| PPTAgent 论文/开源项目 | 将生成定义为两阶段 edit-based workflow：先分析参考 PPT，做 slide clustering 和 schema extraction；再生成 outline，给每页选择 reference slide 和文档片段，通过有限编辑 API 迭代修改参考页，并用执行反馈自修正；PPTEval 从 Content、Design、Coherence 三维评估。资料：[arXiv: PPTAgent](https://arxiv.org/abs/2501.03936)、[arXiv HTML](https://arxiv.org/html/2501.03936v1)、[GitHub](https://github.com/icip-cas/PPTAgent) | 借鉴“参考页/模板 schema 化”：我们已有 single-page JSON 和 generator，可以进一步把真实优秀 PPT/模板抽成 variant schema，减少人工写模板的成本；QA 也应覆盖内容、设计、连贯性三维，而不只是文件是否存在。 |

### 行业共性架构

从这些方案看，成熟 PPT 生成系统通常不是一条“prompt -> PPTX”的直线，而是分层流水线：

1. **输入理解**：prompt、文档、URL、已有 PPT、品牌资产进入系统，先提取主题、受众、目标、章节、事实、图片和品牌约束。
2. **Deck planning**：生成 deck brief、章节叙事、页面数量、每页 message、内容密度和素材需求。
3. **Template/layout retrieval**：按页面意图选择模板或变体，而不是只按 `content_type` 固定走一个布局。
4. **Visual asset retrieval/generation**：按主题和页面角色选择照片、图标、图表、品牌图片；必要时再调用图片生成。
5. **Constraint-based rendering**：由受控 renderer/generator 自动处理对齐、容量、字号、留白、裁剪、蒙层、品牌色。
6. **Iterative refinement**：支持整套风格调整，也支持单页 rewrite/remix/regenerate image/switch variant。
7. **Evaluation loop**：至少检查 Content、Design、Coherence；视觉上看重复度、图文比例、重叠、可读性、品牌一致性。
8. **Native editable output**：输出仍然是可编辑的 PPTX/Slides，而不是一张张不可维护的大图。

### 对当前 PPT Agent 的架构映射

当前系统已经有多 Agent 编排、任务 manifest、single-page 模板 JSON、Python generator、本地背景资源和前端工作台，基础路线是对的。需要补的不是“让模型更会设计”，而是把行业方案里的几个中间层补齐：

- 用 `deck_style_plan` 对齐 Gamma/Presentations.AI/Copilot 的全局主题和品牌约束。
- 用 `layout_variant` 对齐 Beautiful.ai/Plus AI 的多布局和 Remix 能力。
- 用 `visual_intent` 与 `visual_assets` 对齐 Canva/Copilot 的素材驱动生成。
- 用模板 schema / 参考页分析对齐 PPTAgent 论文，后续可从优秀 PPT 反向抽取变体，而不是手工堆 24 类模板。
- 用 `visual_quality_report.json` 对齐 PPTEval 思路，把 QA 从“有没有文件”升级到“内容、设计、连贯性是否成立”。

这也解释了豆包样例观感更好的原因：它看起来不是单页 generator 重复填充，而像是有“主题照片资产 + 页面类型变体 + deck 级节奏控制”。下一轮应优先补这些系统层能力。

## 技术方向

### 方向 A：为核心 content_type 增加多版式变体

建议先选 8 个高频类型做变体，而不是一次铺满 24 个类型：

| 类型 | 建议变体 |
| --- | --- |
| `title_slide` | `photo_full_bleed_center`、`photo_full_bleed_left`、`editorial_split` |
| `section_divider` | `photo_band`、`number_sidebar`、`quiet_title` |
| `content_slide` | `icon_rows`、`statement_cards`、`one_big_idea` |
| `card_grid` | `equal_grid`、`featured_card_plus_grid`、`masonry_cards` |
| `image_text` | `left_photo`、`right_photo`、`photo_strip` |
| `kpi_dashboard` | `large_stat_row`、`metric_cards_with_photo` |
| `chart_slide` | `chart_with_insight_panel`、`full_chart` |
| `case_study` | `photo_context_results`、`problem_solution_results` |

实现方式：

- 在 `TaskItem` / `SlideOutline` 中增加 `layout_variant,omitempty`。
- 在 `templates/single-page/*.json` 的 `contract` 或新增 `variants` 字段声明可用变体、适用场景和容量。
- generator 增加 `variant: str = "auto"` 参数，内部通过 helper 分派布局。
- 主 Agent 规划阶段选择 variant，SlideExecutor 原样传参。

收益：

- 同一个页面类型不再只有一个外观。
- 仍保留 `content_type` 稳定契约，不需要制造大量新类型。

风险：

- 变体过多会增加 generator 复杂度。
- 如果没有 deck-level 选择策略，模型可能随机乱选，反而不统一。

### 方向 B：建立主题视觉资产层

豆包样例的最大优势是主题图片。当前系统只靠背景主题 id，不够细。

建议新增 `visual_assets` 层：

```json
{
  "visual_theme": "forest_mountain_ecology",
  "image_policy": "local_first_then_search",
  "hero_image": {
    "asset_id": "forest_river_valley_01",
    "role": "cover",
    "crop": "cover",
    "overlay": "dark_gradient"
  },
  "supporting_images": [
    {"asset_id": "forest_canopy_01", "role": "section"},
    {"asset_id": "wetland_01", "role": "image_text"}
  ]
}
```

落地切片：

1. 先做本地主题资产，不急着接外部图片搜索。
2. 资产 manifest 扩展字段：`domain_tags`、`scene_tags`、`mood`、`recommended_roles`、`dominant_colors`、`license`。
3. 增加 `asset_selector.py`：按用户主题、页面类型、visual role 返回候选图。
4. 生成器接收 `image_asset` 或 `background_asset`，而不是只有粗粒度 `background` theme。

首批主题建议：

- `forest_mountain_ecology`：森林、山谷、河流、湿地、雪景、木屋、生态旅游。
- `business_technology`：城市、服务器、抽象科技、会议。
- `education_academic`：校园、课堂、论文、实验。
- `government_civic`：党政、城市治理、公共服务。

风险：

- 图片版权和来源必须明确。
- 真实图片会引入裁剪、对比度、主体位置问题，需要 QA。

### 方向 C：升级内容规划为“页面意图 + 视觉意图”

当前 `content_plan` 更像内容结构，不足以指导视觉。

建议扩展为：

```json
{
  "summary": "...",
  "message": "本页要证明什么",
  "visual_intent": {
    "role": "hero_photo | supporting_photo | map | chart | icon | cards",
    "asset_query": "大兴安岭 森林 河谷 航拍",
    "preferred_variant": "photo_full_bleed_center",
    "image_position": "background | left | right | strip",
    "caption": "大兴安岭原始森林与河谷地貌"
  },
  "elements": []
}
```

好处：

- 主 Agent 不只是决定 `content_type`，还决定“这页为什么这么排”。
- SlideExecutor 不需要猜图片和布局。
- 后续 QA 能检查“规划要求照片，但实际没图”的偏差。

### 方向 D：增加 deck-level 视觉节奏控制

豆包样例不是每页都很复杂，而是节奏更像人工设计：

- 强视觉封面
- 浅色信息页
- 图文页穿插
- 章节页统一但有区别
- KPI/案例/生态照片页穿插

建议新增 `deck_style_plan`：

```json
{
  "motif": "forest green + soft cards + natural photography",
  "rhythm": [
    "hero",
    "data_cards",
    "image_text",
    "photo_cards",
    "quote_or_transition"
  ],
  "constraints": {
    "max_same_variant_streak": 2,
    "min_photo_pages_ratio": 0.35,
    "section_pages_use_photo": true
  }
}
```

实现上可以先不写复杂算法，只在主 Agent prompt 和 manifest 增加字段，然后由本地 QA 统计重复度。

### 方向 E：补本地视觉 QA

下一轮不要只靠“能生成 18/18”。建议加轻量 QA：

- 生成 contact sheet。
- 检查每页是否有真实图片/背景。
- 统计连续同 `content_type + variant` 的页数。
- 检查占位符和残留方块。
- 检查信息页文字密度和图片比例。
- 输出 `visual_quality_report.json`。

这可以借鉴 `D:\environment\codeGo\llm-examples\pptx` skill 的 `thumbnail.py` / PDF-to-image QA 思路。

## 推荐迭代路线

### 第 0 阶段：补齐行业路线对应的中间契约

目标：先让系统表达“为什么这一页长这样”，避免直接把外部竞品能力翻译成零散 prompt。

任务：

- 新增 `deck_style_plan`：主题、受众、色板、图片风格、图标风格、页面节奏、品牌限制。
- 新增 `layout_variant`：同一 `content_type` 下的具体排法。
- 新增 `visual_intent`：图片/图标/图表/地图/大数字/对比等页面视觉意图。
- 新增 `visual_assets`：本地图片、图标、品牌素材、来源和 license。
- 新增 `visual_quality_report.json` 草案字段：Content、Design、Coherence、photo coverage、variant repetition、text density、overlap warnings。

验证：

- 只生成 manifest，不生成 PPT，也能看出整套 deck 的视觉计划。
- 大兴安岭样例中，封面/章节页/旅游页/生态页应明确使用森林、河谷、湿地、路线或地图类视觉资产。

### 第 1 阶段：选择器和高频变体，不大改全部生成器

目标：让系统能“计划”出豆包式 PPT，即使生成器先只支持少数变体。

任务：

- 给 `TaskItem`、`SlideOutline` 增加 `layout_variant` 和 `visual_intent`。
- 给核心 8 类模板 JSON 增加 `variants` 元数据。
- 写 `visual_style_planner` 规则或 prompt 片段：根据主题选择 `visual_theme`、背景策略、变体节奏。
- 主 Agent 输出 manifest 时写入 variant/visual_intent。

验证：

- manifest JSON schema 测试。
- 大兴安岭样例规划中，封面/章节/生态旅游页应明确 photo/background intent。

### 第 2 阶段：先做 4 个高收益 generator 变体和局部 Remix

优先：

1. `title_slide`: 全幅照片封面。
2. `section_divider`: 主题照片章节页。
3. `image_text`: 真实照片图文页。
4. `card_grid`: 特色卡 + 小卡组合。

原因：

- 这四类最影响第一眼观感。
- 可以最快降低“统一模板套壳感”。
- Plus AI / Beautiful.ai 都强调单页级别替换和再生成；我们也应把“换版式/换图片/压缩文字”做成局部能力。

验证：

- 生成 12-18 页大兴安岭 deck。
- contact sheet 肉眼检查。
- 至少 30%-40% 页面具备主题照片或明确视觉资产。

### 第 3 阶段：主题资产库

任务：

- 建立 `assets/photos/` 和扩展版 manifest。
- 首批做 `forest_mountain_ecology` 主题资产。
- 引入 `asset_selector.py`。
- 背景从 `background` theme 升级为 `background_asset` / `image_asset`。
- 为每个资产记录 `source`、`license`、`dominant_colors`、`subject_position`、`recommended_roles`，便于裁剪和合规。

验证：

- 相同主题不同页面不重复使用同一张图超过 2 次。
- 封面、章节页、图文页图片主题命中。

### 第 4 阶段：视觉 QA 和评分闭环

任务：

- 新增 `scripts/visual_quality_check.py`。
- 输出 `visual_quality_report.json`。
- 加入指标：
  - variant repetition
  - photo coverage
  - placeholder residue
  - text density warning
  - background readability warning
- 增加 PPTEval 风格三维汇总：Content、Design、Coherence。

验证：

- 对旧生成结果和新生成结果做 A/B 报告。

### 第 5 阶段：参考 PPT/模板 schema 提取

目标：降低手工写 variant 的成本，向 PPTAgent 论文的 reference-presentation analysis 靠近。

任务：

- 选取 20-50 套高质量可授权 PPT 模板或内部优秀 deck。
- 渲染缩略图，按功能页/内容页做聚类。
- 抽取每类的 schema：元素角色、位置比例、容量、图片角色、标题层级、适用场景。
- 将 schema 转为 `templates/single-page/*.json` 的 `variants` 候选，人工审核后进入生成器。

验证：

- 从参考模板抽出的 variant 能被 generator 使用。
- 对比人工写 variant，减少重复定义，且不会引入无法渲染的复杂形状。

## 不建议做的事

- 不建议为每种页面类型复制出大量新 `content_type`，会破坏现有稳定契约。
- 不建议直接让模型在线搜图并随意插图，版权、裁剪、稳定性和成本都不可控。
- 不建议一次重写所有 generator，应该从封面/章节/图文/卡片四个高感知页面开始。
- 不建议让背景图片反向改写全局 palette；全局主题色应仍由 deck 统一控制。

## 结论

下一步最该做的是“视觉表达层”的升级，而不是继续微调单个卡片或字号。

结合公开产品和 PPTAgent 论文，推荐路线调整为：

1. 先补 `deck_style_plan`、`layout_variant`、`visual_intent`、`visual_assets` 这四个中间契约。
2. 给高频页面类型加多版式变体元数据和 variant selector。
3. 先实现封面、章节、图文、卡片 4 类 generator 的照片/变体能力，并支持单页 Remix。
4. 建立本地主题照片资产库，先覆盖自然/地域类主题，同时记录来源、license、主体位置和推荐角色。
5. 用 contact sheet 和 `visual_quality_report.json` 做 A/B 验证，指标覆盖内容、设计、连贯性、照片覆盖率和版式重复度。
6. 后续从优秀 PPT 反向提取模板 schema，把“模板数量”建设为可持续能力，而不是一次性手写堆量。

这样可以保留当前系统的受控生成器优势，同时补上豆包样例更强的“真实主题视觉”和“页面节奏变化”。
