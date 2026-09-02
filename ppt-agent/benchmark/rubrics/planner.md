# Planner Benchmark Rubric

只评价 Planner 首稿 `tasks.draft.json`，不要把 Reviewer 后的修补结果计入 Planner 成绩。Planner benchmark 不测试图片下载和本地落盘能力；图片相关只评价 `visual_policy`、`visual_intent` / `image` 组件中的语义规划质量。

Hard failure，最高 2 分：输出不是合法 DeckSpec、使用非法 `content_type`、缺少核心页面、核心数据丢失、把生成器坐标/字号/颜色等渲染职责硬塞进 Planner、明显无法渲染，或未满足 case `expected.minimum_content_characters` 中适用的正文最低字数。若 case 声明 `expected.content_quality.required_narrative_chain`，而输出缺少中心结论、无法形成该论证链或只是用同义重复替代推进，也属于主要目标失败，最高 2 分。

维度：用户意图覆盖、页面叙事结构、DeckSpec 契约合法性、字段完整度、数据与事实保真、视觉/图片规划合理性、容量控制、正文信息密度、**主题聚焦与跨页叙事连贯性**。

内容质量口径：`model_output.content_quality` 是确定性证据，列出从组件中抽取的 deck 主张、逐页主张、缺失主张页、完全重复主张、连续同类叙事布局，以及目录 `toc_item.title/body` 的完整性；它不替代语义评分。Judge 必须结合 case 的 `expected.content_quality` 判断：

- `deck_thesis` 要回答这套 PPT 希望听众接受或据此决策的核心判断，不能只是主题标题。
- `required_narrative_chain` 中每一环必须由相应页面的 claim、事实或行动承接，形成“问题/判断 → 证据/方案 → 决策/行动”的推进；相邻页不能仅换标题重复同一结论。
- 信息页必须有可复述的主张，组件可作为证据、解释或行动，不能全部等权卡片化。`missing_claim_pages`、`duplicate_claim_groups` 或超出 case `max_consecutive_same_layout` 的连续布局必须明确扣分；若这些报告为空，仍需人工检查 claim 是否具体、是否服务中心结论。
- 若 case 声明 `agenda_subtitles`，每个目录项必须在 `tasks.json` 用 `toc_item.title` 保存章节名、用独立的 `toc_item.body` 保存非重复副标题。`agenda_subtitle_issues` 出现 `missing_toc_subtitle`、`missing_toc_title` 或 `repeated_toc_title` 时不得给满分；缺失或重复副标题导致目录无法解释阅读路径时应视为契约与内容质量缺陷。
- 评分结果的 `dimension_scores` 必须包含 `主题聚焦与跨页叙事连贯性`（1-5）并给出可核对的 strengths/weaknesses。

字数口径：只统计正文组件的 `body`、`text` 和 `items`；标题、图片 caption、来源行不计入。若 case 声明阈值，`image_text` 主正文必须达到 `image_text_narrative`，信息页正文组件合计必须达到 `information_page_total`，完整论述型 `argument_block` 必须达到 `argument_block`。不能通过重复句子、模板话术或缩小字号规避。

图片评分口径：当 `model_output` 中图片计划为 `search_status="planned"` 且包含可执行英文 `asset_query`、明确 `asset_subject` 和合理 `composition` 时，不因缺少 `local_path/source_url/attribution` 扣分；只有伪造下载结果、完全缺少视觉策略、图片计划与主题无关或把下载失败伪装成成功时才扣分。

评分：1 无有效规划；2 主要目标失败；3 可用但有明显质量或契约问题；4 满足主要要求且小问题有限；5 高质量、结构完整、内容具体、可直接进入 Reviewer。
