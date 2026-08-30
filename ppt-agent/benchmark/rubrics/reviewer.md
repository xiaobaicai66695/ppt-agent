# Reviewer Benchmark Rubric

评价 Reviewer 对带缺陷 draft 的修补能力。重点不是重写得漂亮，而是修得准、改得小、不引入新问题。

Hard failure，最高 2 分：没有修复指定 error、修改未授权页面、新增/删除/重排页面、引入新的阻塞性 review issue、输出破坏 DeckSpec 结构，或 case 声明正文最低字数后仍保留 `low_information_density` error。

评分边界：只根据 case 的 `input.review_issues` 和 `expected` 评价本轮授权修复。未授权页面中原本存在、且 Reviewer 没有修改的 warning/error 不能计为本轮失败；不要把用户主题泛化为“必须补全所有页面背景”。为修复指定图片策略问题而在同一次 patch 中新增或更新顶层 `visual_policy` 是 DeckSpec 合法的 deck 级最小修改，不属于越权或结构破坏；前提是策略字段合法，并且目标页的图片计划已按 issue 修复。

维度：问题消除率、未授权页面不变性、patch 最小化、修复后契约合法性、修复内容是否仍符合用户原始请求、正文信息密度。

字数口径：当 `expected.minimum_content_characters` 存在时，只统计正文组件的 `body`、`text` 和 `items`；标题、图片 caption、来源行不计入。Reviewer 必须在授权页补充场景、事实、影响和结论，使 `image_text_narrative`、`information_page_total` 或 `argument_block` 达标；不得通过重复空泛语句、仅替换组件类型或降低字号规避。

评分：1 未有效修复；2 修复失败或越权；3 修了主要问题但有明显副作用；4 修复准确且副作用小；5 精准修复、最小改动、无新问题。
