# Reviewer Benchmark Rubric

评价 Reviewer 对带缺陷 draft 的修补能力。重点不是重写得漂亮，而是修得准、改得小、不引入新问题。

Hard failure，最高 2 分：没有修复指定 error、修改未授权页面、新增/删除/重排页面、引入新的阻塞性 review issue、输出破坏 DeckSpec 结构。

评分边界：只根据 case 的 `input.review_issues` 和 `expected` 评价本轮授权修复。未授权页面中原本存在、且 Reviewer 没有修改的 warning/error 不能计为本轮失败；不要把用户主题泛化为“必须补全所有页面背景”。为修复指定图片策略问题而在同一次 patch 中新增或更新顶层 `visual_policy` 是 DeckSpec 合法的 deck 级最小修改，不属于越权或结构破坏；前提是策略字段合法，并且目标页的图片计划已按 issue 修复。

维度：问题消除率、未授权页面不变性、patch 最小化、修复后契约合法性、修复内容是否仍符合用户原始请求。

评分：1 未有效修复；2 修复失败或越权；3 修了主要问题但有明显副作用；4 修复准确且副作用小；5 精准修复、最小改动、无新问题。
