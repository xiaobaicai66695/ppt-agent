# ppt-agent 后端集成说明

本文档只适用于本仓库的 `ppt-agent` Go 后端。通用 Agent 使用 `ppt-deck-planner` 时不需要读取本文档。

## 集成边界

`ppt-agent` 的正式链路是：

```text
用户需求
→ PPTPlanner 生成完整 DeckSpec 草稿
→ 确定性审查报告
→ TaskPlanReviewer 定向修补
→ Go Validator / Commit
→ 素材物化
→ render_worker_pool 调用 render_task.py
→ 交付对账
```

Planner 应尽力填写完整 DeckSpec 内容字段，但不写 `task_id`、`output_file`、`status`、`qa_report`、`fix_attempts`、`capacity_hint` 等运行态或系统派生字段。

## Planner

- 通过 manifest tool 初始化或 patch DeckSpec 草稿，不直接编辑 JSON 文件。
- 一次性规划整套 PPT 的页序、`content_type`、`layout_variant`、`content_plan.components`、顶层 `visual_policy` 和视觉意图。Planner 不下载图片，也不填写尚未解析的本地路径、来源链接或署名。
- 如果确定性预检未通过，Planner 不反复全量重写；后续由 TaskPlanReviewer 根据失败页/章节切片修正。

## TaskPlanReviewer

- 根据 Go workflow 生成的确定性审查报告和 `included_tasks` 切片进行修正。
- 只 patch 授权页面和必要 deck 级字段。
- 不负责搜索或下载素材；素材缺失时只保留或修正可执行的图片语义字段。
- 审查结果保存在后端约定的 review report 中，不写入单页 item。

## 素材物化

通用 skill 不规定图片保存目录；在 `ppt-agent` 后端中，图片下载、去重、路径写回和背景收敛由后端适配层负责。

- 通用 skill 默认要求 `visual_policy.mode="required"`；只有用户明确请求纯文字时才使用 `mode="none"`。后端在 Reviewer 后由既有素材链路搜索、下载、去重并写回 `visual_intent` 与 `image` 组件的 `local_path`、来源和署名。
- `scripts/hydrate_unsplash_assets.ts` 仅供 `ppt-agent` 项目外的通用 Agent 使用，项目内 Agent 不读取其 `auth.txt`，也不调用该脚本。
- 背景按 `content_type` 收敛，减少同类页面视觉节奏漂移；required 策略下缺少背景或前景素材会阻断渲染。

## Benchmark 含义

- Planner benchmark 评价“完整草稿质量”，包括初次审查错误数、意图覆盖、叙事结构和契约字段质量。
- Reviewer benchmark 评价“按报告修补能力”，包括问题消除率、未授权页面不变性和收敛轮次。
- Renderer benchmark 使用已过审 `tasks.json`，只评价 skill scripts 的可渲染性和视觉稳定性。
