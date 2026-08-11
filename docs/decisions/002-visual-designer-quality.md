# 002 Visual Designer 与 PPT 视觉质量决策

## 当前决策

Visual Designer 的长期方向是“结构化规划 + 生成器兜底视觉质量”。LLM 负责选择 `content_type`、控制内容容量、填写 `description/content_plan/layout_variant/background/source`；具体字号、文本框、蒙版、图片解析、卡片和图表绘制由 Python generators 与 `base.py` 负责。

## 执行准则

- `content_type` 只能使用 `references/slide_types.md` 中的稳定英文 id，图表形态、版式意图和视觉材料写入 `content_plan` 或 `layout_variant`。
- 信息页优先做结构化参数和容量控制，不把 300-400 字长段落硬塞给 `content_slide`、`card_grid`、`two_column`、`kpi_dashboard`、`chart_slide`。
- 背景库当前按 10 类主题维护，启用背景时整套 PPT 使用同一主题目录；具体图片可按页变化，但不能跨主题乱跳。
- 背景、图片和素材必须来自 manifest 可追溯资源；不再硬编码主题目录，不用黑白或跑偏背景充当默认素材。
- 图文页缺少用户图片时使用可替换的分类默认图片，不暴露“图片占位”或内部标签。
- 生成器改动必须同步模板 JSON、`references/generators.md`、必要测试和 smoke 渲染；只改规划规则时优先改 Skill/prompt，不改 generator。

## 取舍

把视觉实现下沉到 generator 会降低 LLM “临场发挥”的自由度，但能保证同一内容类型的容量、对比度、背景处理和文件可替换性稳定，也避免 prompt 中出现与实际渲染无效的字号/坐标指令。

## 来源任务

- `PPT-SKILL-001`
- `PPT-SKILL-002`
- `PPT-SKILL-003`
- `PPT-SKILL-004`
- `PPT-SKILL-005`
- `PPT-QUALITY-001` 至 `PPT-QUALITY-011`
- `PPT-PROMPT-001`

