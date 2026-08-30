# Benchmark 不达标修复指南

这份文档给后续修复 Agent 使用：当 `pptbench` 跑出低分、失败或模型错误时，先按这里定位，再决定改 case、prompt、工具还是 workflow。不要直接扩大重构范围。

## 先确认评测阶段

`pptbench` 是两步流程：

```text
-p model  ->  生成 model_output.json，供人工/Agent 审阅
-p judge  ->  读取同一 run 目录的 model_output.json，生成 score.json
```

如果用户说“对第一次跑出来的结果评分”，重点确认两次命令的 `-o` 是否完全相同。例如：

```powershell
go run ./cmd/pptbench --dataset test -s reviewer -p model -o ../benchmark/runs/20260829-150405-test-reviewer
go run ./cmd/pptbench --dataset test -s reviewer -p judge -o ../benchmark/runs/20260829-150405-test-reviewer
```

不要把 run 目录直接写成命令末尾的位置参数；必须写 `-o <run-dir>`。目录统一为 `YYYYMMDD-HHMMSS-<dataset>-<suite>`；同一次 model/judge 两阶段使用完全相同的时间戳、dataset 和 suite。

第二步不会重新跑 Agent；它只读取：

```text
benchmark/runs/<run>/reviewer/<case-id>/model_output.json
```

## 先看哪些文件

每个失败 case 目录下按这个顺序看：

1. `score.json`

   只看分数、`critical_failures`、`weaknesses`、`recommended_fix`。这是判断“为什么不达标”的第一入口。

2. `model_output.json`

   看被测 Agent 的核心产物。Planner 看 `output`，Reviewer 看 `before`、`after` 和 `deterministic_review`，Fixer 只看 `before` 和 `after`，Router 看 `output.intent`。

3. `case.json`

   确认 case 自己是否写偏，尤其是 `expected` 是否过严、fixture 是否混入无关错误。

4. `trace.json`

   只有当 `model_output.json.error` 不为空、工具没调用、模型异常、输出缺失时才看。正常质量问题不要先从 trace 开始。

5. `judge_input.json`

   只有怀疑 Judge 收到的信息不完整或 rubric 不对时才看。

## 判断问题归属

优先把失败分成四类：

- Case 问题：输入 fixture 本身不符合当前 DeckSpec 契约，或一个 case 混入多个无关缺陷。
- Agent 问题：模型输出没有满足任务目标，例如 Planner 首稿漏字段、Reviewer 没修目标 error、Fixer 越权修改。
- Judge 问题：评分理由和 `model_output.json` 明显不一致，或 rubric 表述导致误判。
- 环境问题：模型 key、provider、额度、网络、超时导致 `error`，不是能力低分。

不要把环境错误当作 prompt 质量问题。不要用 Reviewer/Fixer 的结果倒推 Planner 首稿质量。Fixer 流程不运行 Reviewer，也不使用全量 PlanReview 结果评分。

## 各 Suite 修复入口

### Router

看：

```text
model_output.json -> output.intent / output.reason / target_pages / fix_details
```

常见失败：

- 新建 PPT 请求没有走 `create_deck`。
- 创建入口里的已有页面修改请求没有走 `fix_existing`。
- 继续任务中局部修改没有走 `fix`。
- 全部重做没有走 `regenerate_all`。
- 页码或修改对象丢失。

优先修改：

- 创建入口路由：`backend/pkg/web/request_router.go`
- 继续任务路由 prompt：`backend/pkg/web/handler.go` 中 `classifyIntentByLLM`
- Benchmark 专用入口只在调用方式有问题时改：`backend/pkg/web/benchmark_router.go`

### Planner

看：

```text
model_output.json -> output
model_output.json -> deterministic_review
```

Planner benchmark 只评首稿，不看 Reviewer 后结果。常见失败：

- 没生成合法 `tasks.draft.json`。
- 缺顶层 `visual_policy`。
- `content_type` 非法。
- `content_plan.components` 缺失或组件类型非法。
- KPI/chart/table 等结构化数据不完整。
- 把坐标、字号、颜色、margin 等生成器职责写进 DeckSpec。
- 用户给的事实、数字、页数、受众没有进入页面规划。

Planner benchmark 显式禁用图片下载工具，只评估图片语义规划。`search_status="planned"` 且有可执行英文 `asset_query`、`asset_subject`、`composition` 时是可接受输出；不要因为缺少 `local_path` 去修 Planner。生产主流程是否能下载图片取决于 `UNSPLASH_ACCESS_KEY` 和网络状态。

优先修改：

- Planner prompt：`backend/pkg/prompts/planner/master_instruction.tmpl`
- Planner manifest 工具约束：`backend/pkg/agent/deck/manifest_tool.go`
- 硬校验和首稿质量门：`backend/pkg/agent/deck/plan_review_tool.go`
- 组件契约：`skills/ppt-deck-planner/templates/component_contracts.json`

不要修改 Reviewer 来掩盖 Planner 首稿问题。

### Reviewer

看：

```text
model_output.json -> before
model_output.json -> after
model_output.json -> deterministic_review
case.json -> input.review_issues
case.json -> expected.allowed_change_pages / must_not_change_pages
```

常见失败：

- 指定 error 没修复。
- 修改了未授权页面。
- 新增、删除或重排页面。
- patch 太大，把整页或整套重写了。
- 修复一个问题后引入新的阻塞问题。

优先修改：

- Reviewer prompt：`backend/pkg/prompts/reviewer/master_instruction.tmpl`
- Reviewer 切片输入：`backend/pkg/agent/deck/plan_review_revision.go`
- 草稿 patch 工具：`backend/pkg/agent/deck/fixer_manifest_tool.go`
- 审查规则：`backend/pkg/agent/deck/plan_review_tool.go`

如果 `before` 本身有很多非目标错误，优先修 case fixture，而不是调 Reviewer。

### Fixer

看：

```text
model_output.json -> before
model_output.json -> after
case.json -> input.allowed_page_indexes
case.json -> expected.must_change / must_not_change_pages
```

常见失败：

- 没执行用户明确修改。
- 改了未授权页面。
- 改了未授权字段，例如用户只改标题但同时改了 `content_plan`。
- 为了局部修改重写整套 DeckSpec。
- 破坏了被修改页的基本结构，例如清空标题、组件列表或写入不可解析字段。

优先修改：

- Fixer prompt：`backend/pkg/prompts/fixer/master_instruction.tmpl`
- 选页 patch 工具：`backend/pkg/agent/deck/fixer_manifest_tool.go`
- 继续任务路由和页码识别：`backend/pkg/web/handler.go`

Fixer case 应尽量显式写 `allowed_page_indexes`，避免把页码推断问题混入 Fixer 能力评测。不要把 Reviewer 或 `ReviewTasksManifest` 接入 Fixer benchmark；Fixer 的职责是局部追改，不承担全量质量门修复。

## 什么时候改 Case

可以改 case 的情况：

- `draft_tasks` 或 `base_tasks` 使用了当前契约不支持的组件类型。
- Reviewer case 同时触发多个无关 error，无法隔离被测缺陷。
- `expected` 和用户请求不一致。
- case 缺少必要上下文，例如 Fixer 没有给 `allowed_page_indexes`。
- case 的目标不是当前 suite 的职责，例如 Planner case 要求 Reviewer 才能完成的修复能力。

不要为了让分数好看而降低 `expected`。如果期望合理，应该修 Agent、prompt 或工具。

## 什么时候改 Rubric 或 Judge

可以改 rubric 的情况：

- Judge 对同一类问题评分前后不一致。
- Judge 把环境错误当成能力低分。
- Judge 忽略 hard failure，例如越权修改页面仍给 4 分。
- Judge 没有区分语义质量和确定性契约错误。

优先修改：

- `benchmark/rubrics/<suite>.md`
- Judge 输入构造：`backend/cmd/pptbench/judge.go`

不要在 Judge 里塞业务补丁逻辑。Judge 只评分，不修复。

## 复跑方式

修 case 或 Agent 后，先跑 model step：

```powershell
go run ./cmd/pptbench --dataset test -s reviewer --case reviewer_repair_chart_001 -p model -o ../benchmark/runs/20260829-150405-test-reviewer
```

人工检查 `model_output.json` 后再评分：

```powershell
go run ./cmd/pptbench --dataset test -s reviewer --case reviewer_repair_chart_001 -p judge -o ../benchmark/runs/20260829-150405-test-reviewer
```

只改 rubric 或 Judge 逻辑时，不要重新跑 Agent，直接复评同一 run：

```powershell
go run ./cmd/pptbench --dataset test -s reviewer -p judge -o ../benchmark/runs/20260829-150405-test-reviewer
```

## 修复后的最低验证

## 测试集通过后必须使用验证集复验

`test` 只用于定位问题和迭代修复，不能作为上线依据。任何 prompt、Agent 工具、质量门、rubric 或 benchmark case 修复完成后，必须遵循下面的发布门禁：

```text
test 集定位与修复
  -> test 集全量通过
  -> 使用 validation 集跑同一 suite 的全量 model + judge
  -> 所有 validation case 达标、无 critical failure
  -> 才能构建、提交和上线
```

- 不允许拿 test 集分数替代 validation 分数，也不要只挑 validation 中的一两个 case。
- validation run 必须使用新的独立目录，保留 `model_output.json`、`score.json` 与 `summary.md` 作为发布证据；`benchmark/runs/` 仍是本地产物，不提交到 Git。
- 任一 validation case 低于当前 suite 的达标线、出现 `critical_failures`、模型/环境错误或无法完成评分时，均视为未达标：回到 test 集定位并修复，随后重新跑一轮完整 validation；不得带着失败 validation 上线。
- 只修改 rubric 或 Judge 时，可以复用同一批 model 输出执行 judge；修改 prompt、Agent、工具或工作流时，validation 必须重新执行 model 和 judge，不能复用旧输出。

示例（Planner 全量验证）：

```powershell
go run ./cmd/pptbench --dataset validation -s planner -p all -o ../benchmark/runs/20260831-153000-validation-planner
```

完成验证后，在迭代记录中写明 test 与 validation 的 run 路径、各 case 分数、是否存在 critical failure，以及该证据对应的发布版本。

修改 Go benchmark 或 Agent 入口后，至少运行：

```powershell
go test ./cmd/pptbench ./pkg/agent/deck ./pkg/web
go run ./cmd/pptbench -s router -p model -l 1
```

如果改了某个 suite 的 prompt 或工具，再跑对应单 case：

```powershell
go run ./cmd/pptbench --dataset test -s planner --case planner_business_kpi_001 -p model -o ../benchmark/runs/20260829-150405-test-planner
```

`benchmark/runs/` 是本地评测产物，不要提交。不要把 API key 写入代码、case、README 或 `score.json`。
