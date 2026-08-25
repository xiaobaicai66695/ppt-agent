# Planner、Task Reviewer 与 PPT Fixer 职责边界

## 决策背景

旧链路把规划、规划自审、恢复指令和生成后修复都塞进同一个 Planner prompt。结果是提示词过长、职责冲突，模型需要在同一上下文中反复读写 `tasks.json`，并可能把生成后的局部修复规则误用于首轮规划。

本次改造采用三个 Agent、两个运行时机：

```text
用户输入
  -> 意图分类 / 画像门控
  -> PPTPlanner
  -> TaskPlanReviewer
  -> Go Validator / Commit
  -> Worker Pool 按 task_id 并发渲染
  -> 交付

用户对已生成 PPT 提出问题
  -> 继续请求意图路由
  -> PPTFixer（仅目标页面）
  -> Go 更新状态并定点重渲染
```

## Agent 边界

### PPTPlanner

- 无论用户是否提供大纲，都生成或补全一份完整 DeckSpec 草稿。
- 用户大纲是内容约束和输入，不再成为跳过规划的快捷路径。
- 一次性调用 `update_tasks_manifest(mode="initialize")` 写入完整页面数组，不逐页 patch，不审查、不 commit。
- 负责主题理解、资料准备、叙事拆页、组件语义和视觉意图。

### TaskPlanReviewer

- 只审查和修正 DeckSpec 草稿，不参与首轮内容发散。
- Go 先执行确定性审查并生成结构化 issues；Reviewer 根据 issues 批量调用一次 `patch_tasks_draft`。
- 每轮之后由 Go 重新校验，最多三轮；通过后由 Go 原子提交正式 `tasks.json`。
- Reviewer 工具在代码层面没有 initialize/commit 权限。

### PPTFixer

- 只在 PPT 已生成且用户提出定点问题后运行。
- 后端先确定目标页面，并将允许修改的 task_id 白名单传给 `patch_selected_tasks`。
- Fixer 不能新增、删除、重排页面，不能修改未授权页面、整套 theme/template 或运行时身份字段。
- Fixer 修订完成后，Go 将目标页标记为 pending、删除旧单页文件并定点重渲染；模型失败时保留用户要求并走代码兜底。

## 非 Prompt 策略

- 配色不维护独立优先级 prompt。当前用户明确要求和当前主题优先，历史偏好仅作为弱参考。
- 背景图片、图片检索、内容密度、组件容量和 `visual_intent` 规则统一维护在 `skills/ppt-deck-planner/SKILL.md`，Planner 与 Reviewer 共用。
- 缺失草稿恢复、审查轮次、硬校验、指纹一致性、原子 commit、渲染重试和 Fixer 失败兜底全部由 Go 代码控制。
- `tasks.draft.json` 是内部暂存草稿；只有通过 Reviewer 与硬校验后才发布正式 `tasks.json`，渲染器只读取正式文件。

## 代码落点

- Agent 构造与提示词装配：`ppt-agent/backend/pkg/agent/deck/agent.go`
- Planner/Reviewer 编排与提交：`ppt-agent/backend/pkg/agent/deck/run.go`
- Reviewer 硬校验：`ppt-agent/backend/pkg/agent/deck/plan_review_tool.go`
- Reviewer/Fixer 限权工具：`ppt-agent/backend/pkg/agent/deck/fixer_manifest_tool.go`
- 生成后继续请求：`ppt-agent/backend/pkg/web/handler.go`
- 共享图片与组件策略：`ppt-agent/skills/ppt-deck-planner/SKILL.md`
