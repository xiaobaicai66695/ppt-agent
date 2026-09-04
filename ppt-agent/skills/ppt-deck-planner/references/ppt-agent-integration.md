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
- 一次性规划整套 PPT 的页序、`content_type`、`layout_variant`、`content_plan.components`、视觉意图和来源信息。
- 如果确定性预检未通过，Planner 不反复全量重写；后续由 TaskPlanReviewer 根据失败页/章节切片修正。

## TaskPlanReviewer

- 根据 Go workflow 生成的确定性审查报告和 `included_tasks` 切片进行修正。
- 只 patch 授权页面和必要 deck 级字段。
- 不负责搜索或下载素材；素材缺失时只保留或修正可执行的图片语义字段。
- 审查结果保存在后端约定的 review report 中，不写入单页 item。

## 素材物化

通用 skill 不规定图片保存目录；在 `ppt-agent` 后端中，图片下载、去重、路径写回和背景收敛由后端适配层负责。

- 图片搜索是默认执行步骤：除用户明确要求无图、页面明确标记 `clean_text_only` 或素材服务返回已记录的实际失败外，每页都必须通过后端既有素材链路搜索、下载并写回 `local_path`、`provider` 和 `search_status`。
- `search_images` Tool 仅供项目内闲聊 Agent 搜索并展示 Unsplash 图片候选，不下载 PPT 素材；Planner、Reviewer 与 Fixer 均不注册或调用它，只提交视觉意图。通用 Agent 需要处理独立 DeckSpec 时使用本 skill CLI。需要在服务器上手工运行 CLI 时，使用 `unsplash auth --from-env` 从服务 `.env` 的 `UNSPLASH_ACCESS_KEY` 完成无交互认证；后端进程也必须保留该环境变量，供闲聊 Tool 与素材物化层访问 Unsplash API。
- 同一任务内相同 `content_type` 的页面必须收敛到同一张背景图：共享 `asset_query`、`local_path`、来源和署名；只有不同 `content_type` 才可使用不同背景。默认背景图缺失应阻断无理由交付，只有已记录豁免或素材失败才可继续无图渲染；图片必填组件始终须有有效本地图片路径。

## Benchmark 含义

- Planner benchmark 评价“完整草稿质量”，包括初次审查错误数、意图覆盖、叙事结构和契约字段质量。
- Reviewer benchmark 评价“按报告修补能力”，包括问题消除率、未授权页面不变性和收敛轮次。
- Renderer benchmark 使用已过审 `tasks.json`，只评价 skill scripts 的可渲染性和视觉稳定性。
