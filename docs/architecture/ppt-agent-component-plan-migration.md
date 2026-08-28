# PPT Agent 组件级规划迁移

## 背景

当前快速迭代目标是把生成前置规划做重，而不是继续让渲染阶段猜测自然语言描述。生成器仍按页执行，但 `DeckSpec/tasks.json` 开始承载可执行的页内组件计划。

## 新主线

```text
用户输入
→ 创建入口意图分类
→ Deck Planner：整套 PPT 叙事规划
→ Content Planner：页级内容规划
→ Component Planner：页内组件编排
→ Plan Reviewer：审查结构、密度、场景匹配、模板容量
→ Plan Refiner：按审查意见重写规划，最多 3 轮
→ Validator：硬校验 DeckSpec
→ 逐页渲染
→ 视觉 QA / 局部修复
→ 交付
```

## 边界

组件级规划不是让 LLM 画 UI。LLM 只决定：

- 这一页有哪些语义组件。
- 每个组件承载什么信息。
- 哪个组件是主视觉或主论点。
- 组件之间是什么关系。
- 内容密度是否超过模板容量。
- 是否需要拆页、合并页或换布局。

坐标、字号、颜色、背景处理、卡片尺寸、图表区域和文字适配仍由 Python generator 与模板 contract 决定。

## 第一阶段：组件级规划，逐页渲染

本阶段已引入 `content_plan.components`：

```json
{
  "slide_intent": "说明产品能力矩阵",
  "components": [
    {"id": "headline_1", "type": "headline", "text": "三层能力矩阵支撑端到端交付", "role": "main_point"},
    {"id": "feature_card_1", "type": "feature_card", "title": "数据接入", "body": "统一连接业务库、文件和接口数据", "emphasis": "primary"}
  ],
  "capacity_hint": {"estimated_density": "normal", "overflow_risk": "low", "component_count": 2},
  "reviewer_status": {"planner_round": 1, "locked": true, "issues": []}
}
```

`render_task.py` 消费 `components` 作为主数据源。旧 `elements/summary` 不再作为可维护契约；`description` 只保留页面意图摘要，不能替代组件计划。后端 manifest 写入和渲染前都会做硬校验。

## 后续阶段

第二阶段再把 Python generator 内部抽成组件渲染器，例如 `feature_card`、`kpi_metric`、`quote_block`、`chart_component`。第三阶段再做组件级 QA 和局部修复，例如只缩短第 3 页第二张卡片，而不是重写整页。

当前阶段先验证收益：提升内容密度审查能力，并降低 generator 从自然语言中猜参数的压力。
