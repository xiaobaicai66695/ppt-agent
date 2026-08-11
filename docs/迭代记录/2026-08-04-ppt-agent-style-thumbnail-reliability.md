# 2026-08-04 PPT 整套风格与缩略图可靠性修复

## 迭代目标

修复生成 PPT 中章节页与内容页配色体系分裂，以及实际 PPTX/JPG 已生成但前端显示“缩略图转换失败”的问题。本次事项登记为 `PPT-QUALITY-002`，进入 OpenSpec change：`openspec/changes/ppt-agent-deck-style-and-thumbnail-reliability/`。

## 根因

- 报告任务选择 `personal-summary + charcoal_light`，主 Agent 又根据“两会/政府”主题给部分页面写入 `party_government`。
- SlideExecutor 随后通过背景推荐把这些页面改为 `government_red`，造成同一套 PPT 同时存在红色章节页和灰蓝内容页。
- 第 6、9 页 manifest 文件名不含多余空格，实际 PPTX 与 JPG 标题中多了空格；LibreOffice 转换本身成功，但前端请求了不存在的 manifest 路径。

## 已完成

- `TaskOutline` 的背景开关和推荐背景进入 Prompt 模板数据，主 Agent 只能执行任务入口已经确定的背景策略。
- 用户未启用背景时，空 background 保持为空；智能推荐启用背景时，只能使用后端选定的背景。
- tasks.json 顶层 `theme` 成为整套 PPT 唯一 palette；背景只影响图片与对比度，不再调用 `get_palette_for_background` 反向改色。
- 初次生成与对话重生成均要求逐字符使用 task.output_file，禁止根据标题重新拼接文件名。
- 后端 reconcile 在声明文件不存在时查找唯一同页 PPTX，原子重命名为 manifest 文件名，并同步规范化已有 JPG；多个候选时不猜测。
- 前端对历史任务保留兼容：同一页只有一个 ready 文件时使用实际文件名；多个候选时维持失败态。
- 已将报告任务 `95880d3a-3f39-4a07-8c12-64fd99abf864` 的第 6、9 页 PPTX/JPG 规范化，未额外触发高成本 LLM 重生成。

## 验证与部署

- `go test ./pkg/prompts ./pkg/agent/deck ./pkg/task ./pkg/web` 通过。
- `go build ./...`、`npm test -- --run`、`npm run build` 通过。
- `openspec validate ppt-agent-deck-style-and-thumbnail-reliability --strict` 通过。
- 远端聚焦 Go 测试和 Linux 构建通过；服务 PID `4094435`，健康检查 200。
- 第 6、9 页缩略图接口均返回 `200 image/jpeg`；文件为有效 JFIF JPEG，分辨率 2000×1125。

## 说明

整套风格锁定作用于后续新生成和重生成任务。报告任务中已经生成的红色章节页未在本轮重新调用模型生成，因此旧 PPTX 的视觉内容保持原样；其缩略图失败已立即修复。

## 关联内容

- `docs/issues/todo.md`
- `openspec/changes/ppt-agent-deck-style-and-thumbnail-reliability/`
- `ppt-agent/backend/pkg/prompts/planner/*.tmpl`
- `ppt-agent/backend/pkg/agent/deck/types.go`
- `ppt-agent/frontend/src/utils/workbench.ts`
