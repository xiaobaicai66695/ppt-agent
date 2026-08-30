# Planner Benchmark Rubric

只评价 Planner 首稿 `tasks.draft.json`，不要把 Reviewer 后的修补结果计入 Planner 成绩。Planner benchmark 不测试图片下载和本地落盘能力；图片相关只评价 `visual_policy`、`visual_intent` / `image` 组件中的语义规划质量。

Hard failure，最高 2 分：输出不是合法 DeckSpec、使用非法 `content_type`、缺少核心页面、核心数据丢失、把生成器坐标/字号/颜色等渲染职责硬塞进 Planner、明显无法渲染，或未满足 case `expected.minimum_content_characters` 中适用的正文最低字数。

维度：用户意图覆盖、页面叙事结构、DeckSpec 契约合法性、字段完整度、数据与事实保真、视觉/图片规划合理性、容量控制、正文信息密度。

字数口径：只统计正文组件的 `body`、`text` 和 `items`；标题、图片 caption、来源行不计入。若 case 声明阈值，`image_text` 主正文必须达到 `image_text_narrative`，信息页正文组件合计必须达到 `information_page_total`，完整论述型 `argument_block` 必须达到 `argument_block`。不能通过重复句子、模板话术或缩小字号规避。

图片评分口径：当 `model_output` 中图片计划为 `search_status="planned"` 且包含可执行英文 `asset_query`、明确 `asset_subject` 和合理 `composition` 时，不因缺少 `local_path/source_url/attribution` 扣分；只有伪造下载结果、完全缺少视觉策略、图片计划与主题无关或把下载失败伪装成成功时才扣分。

评分：1 无有效规划；2 主要目标失败；3 可用但有明显质量或契约问题；4 满足主要要求且小问题有限；5 高质量、结构完整、内容具体、可直接进入 Reviewer。
