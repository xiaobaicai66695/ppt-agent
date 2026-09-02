# PPT Agent Benchmark 操作手册

`pptbench` 是 PPT Agent 的独立效果评测工具。它不是单元测试，也不替代 `go test`；它的目标是用人工可审的 JSON case 跑真实 Agent，再用项目配置的 Judge LLM 按 1-5 分评价效果。

推荐使用方式是分两步：先跑 `-p model` 生成 `model_output.json`，人可以先审这份核心输出；确认没有明显问题后，再跑 `-p judge` 生成 `score.json`。

评测不达标后需要让 Agent 修复时，先看 [Benchmark 不达标修复指南](fix-guide.md)。

## 评测范围

当前覆盖 4 个 suite：

- `router`：评估创建入口和继续任务的意图识别，确认请求进入正确链路。评测输出使用稳定词汇 `create_deck`、`fix_existing`、`fix`、`regenerate_all`；创建入口 HTTP API 内部的 `create`、`fix` 会在 benchmark 适配层映射为前两者，并保留实际下游 Agent 和原始请求。
- `planner`：只评估 Planner 首稿 `tasks.draft.json`，不让 Reviewer 修补后再评分；benchmark 中不挂载图片下载工具，同时评价图片语义规划、主题聚焦和跨页叙事连贯性。
- `reviewer`：用带缺陷的 draft 和 review issue 评估 Reviewer 是否精准修补。
- `fixer`：用真实用户追改请求评估 Fixer 是否只改授权页面和必要字段；不运行 Reviewer 或全量 DeckSpec review。

旧的 `backend/test/plan_benchmark` 只保留为低成本契约 smoke。Agent 效果评测以本目录和 `cmd/pptbench` 为主。

## 目录结构

```text
benchmark/
  README.md
  config.example.json
  rubrics/
    router.md
    planner.md
    reviewer.md
    fixer.md
  cases/
    router/                 # 原有测试集，保持不动
    planner/
    reviewer/
    fixer/
  validation_cases/         # 新建的独立 holdout，不得据此逐例调 prompt
    router/
    planner/
    reviewer/
    fixer/
  runs/
    .gitkeep
```

`benchmark/runs/` 用于保存本地评测结果，已被 `.gitignore` 忽略，运行产物不会进入提交。每个 run 目录固定命名为 `YYYYMMDD-HHMMSS-<dataset>-<suite>`，例如 `20260829-150405-validation-planner`；全量评测使用 `all`。即使显式传入 `-o`，也必须使用这个格式，二阶段评分复用完全相同的目录。

### 样本基线

`test` 与 `validation` 的 `router`、`planner`、`reviewer`、`fixer` 各自至少包含 **10 条** case。validation 是独立 holdout：不得复用 test 的 case ID 或完全相同的用户请求。`go test ./cmd/pptbench` 会在无需模型 Key、无需网络的情况下校验这条基线。

## 环境准备

从项目根目录进入后端目录运行命令：

```powershell
cd ppt-agent/backend
```

`pptbench` 会自动加载：

- `../.env`
- `.env`

真实 Agent 依赖项目现有模型配置。Benchmark 中的 Agent 和 Judge 固定使用 DeepSeek provider，同一把 key 只从本机环境变量读取：

- `PPT_BENCH_JUDGE_API_KEY`

这把 key 只在 benchmark 的模型初始化中使用：`-p model` 初始化 Planner/Reviewer/Fixer/Router，`-p judge` 初始化 Judge；不会进入生产服务、上线配置或其它测试。

Planner benchmark 会显式禁用图片下载工具，不依赖 `UNSPLASH_ACCESS_KEY`、网络状态或本地图片落盘。评分只看是否规划了合理的 `visual_policy`、`asset_query`、`asset_subject`、`composition` 和 `search_status="planned"`，不要求出现 `local_path`。生产主流程不受此限制：配置 `UNSPLASH_ACCESS_KEY` 后仍会在 Planner 中挂载 `search_images(download=true)`。

如果模型额度、Key、Provider 配置异常，输出会显式写入 `agent_error` 或 `judge_error`，不会被隐藏成成功。

## 常用命令

只跑 Router 的 Agent 输出，不调用 Judge：

```powershell
go run ./cmd/pptbench -s router -p model
```

只跑某一个 case：

```powershell
go run ./cmd/pptbench -s fixer --case fixer_title_only_001 -p model
```

每个 suite 只跑 1 个 case，用于低成本检查：

```powershell
go run ./cmd/pptbench -s all -l 1 -p model
```

跑完整评测，包括 Judge 打分：

```powershell
go run ./cmd/pptbench -s all -p all -o ../benchmark/runs/20260829-150405-test-all
```

复用已有 Agent 输出重新打分：

```powershell
go run ./cmd/pptbench -s planner -p judge -o ../benchmark/runs/20260829-150405-test-planner
```

注意：二阶段评分必须用 `-o` 指向第一阶段的 run 目录。不要把目录直接追加在命令末尾；位置参数不会表示输出目录。

## 参数说明

- `--dataset test|validation`：选择评测数据集，默认 `test`。`test` 使用既有 `benchmark/cases/`，`validation` 使用独立新建的 `benchmark/validation_cases/`。
- `-s router|planner|reviewer|fixer|all`：选择评测套件，默认 `all`。
- `-p model|judge|all`：`model` 只生成模型输出，`judge` 只给已有输出打分，`all` 一次跑完两步。默认 `model`。
- `--cases <file-or-dir>`：指定 case 文件或目录。默认读取 `test` 的 `../benchmark/cases/<suite>` 或 `validation` 的 `../benchmark/validation_cases/<suite>`。
- `-o <dir>`：指定输出目录。默认写入 `../benchmark/runs/YYYYMMDD-HHMMSS-<dataset>-<suite>`；显式指定时也必须遵循该格式。
- `--case <case-id>`：只运行指定 case。
- `-l <n>`：每个 suite 最多运行 n 个 case。
- `--timeout <duration>`：单 case 超时时间，默认 `45m`。

## 推荐工作流

1. 编写或修改 case。

   原有测试 case 继续放在 `benchmark/cases/<suite>/` 下。每个 case 应包含：`id`、`name`、`input`、`expected`、`judge_focus`。

2. 先跑 `-p model`。

   目标是检查 Agent 是否能被调用、case 是否写偏、输出是否落盘、错误是否暴露。

   ```powershell
   go run ./cmd/pptbench --dataset test -s reviewer -p model -o ../benchmark/runs/20260829-150405-test-reviewer
   ```

3. 人工查看输出。

   重点看每个 case 目录里的 `case.json`、`model_output.json` 和 `summary.md`。`trace.json` 只在需要排查工具事件或模型错误时再看。

4. 确认 case 合理后跑完整评测。

   ```powershell
   go run ./cmd/pptbench --dataset test -s reviewer -p judge -o ../benchmark/runs/20260829-150405-test-reviewer
   ```

5. 对比评分并修 prompt、工具或 workflow。

   修改后重新跑同一批 case，用 `summary.json` 和 `score.json` 对比变化。

## 输出文件说明

一次运行会生成类似结构：

```text
benchmark/runs/20260829-150405-test-all/
  config.json
  summary.json
  summary.md
  planner/
    planner_business_kpi_001/
      case.json
      model_output.json
      trace.json
      judge_input.json
      score.json
```

文件含义：

- `config.json`：本次运行参数。
- `summary.json`：机器可读汇总。
- `summary.md`：人工可读汇总。
- `case.json`：本次实际使用的 case 快照。
- `model_output.json`：给人审阅的核心模型/Agent 产物，例如 Planner 首稿 task.json、Reviewer 修补前后结果、Fixer 定点修改前后结果。
- `trace.json`：完整调试轨迹，包含事件和耗时；正常审阅不必看。
- `judge_input.json`：发送给 Judge LLM 的完整输入。
- `score.json`：Judge 的 1-5 分结构化评分。

## 评分规则

Judge 必须输出 JSON，核心字段为：

```json
{
  "case_id": "planner_business_kpi_001",
  "suite": "planner",
  "score": 4,
  "pass": true,
  "dimension_scores": {
    "intent_coverage": 5,
    "structure_quality": 4,
    "contract_validity": 4,
    "specificity": 4,
    "scope_control": 5
  },
  "strengths": ["页面结构完整"],
  "weaknesses": ["区域对比页数据来源略弱"],
  "critical_failures": [],
  "recommended_fix": "加强 chart_slide 数据字段要求。"
}
```

统一标准：

- `1`：完全失败或违反硬性约束。
- `2`：主要目标失败，只完成少量要求。
- `3`：基本可用，但存在明显质量或约束问题。
- `4`：良好，满足主要要求，仅有小问题。
- `5`：优秀，完整满足要求且质量高。

`score >= 4` 且没有 `critical_failures` 视为通过。

Hard failure 最高只能 2 分，包括：

- 输出不是合法 JSON。
- 使用非法 `content_type`。
- Router 路由到错误 Agent 或错误主流程。
- Planner 漏掉核心用户请求。
- Reviewer 没修复指定 error。
- Fixer 修改了禁止修改的页面或字段。
- Fixer 破坏被修改页的基本可渲染结构。

## Case 编写要点

`planner` case 应评估首稿质量，不写“Reviewer 可以补齐”的期待。重点看页面结构、事实覆盖、字段完整度、图片语义规划、DeckSpec 合法性以及主题聚焦。Planner benchmark 不测图片下载，所以不要把缺少 `local_path` 作为失败条件。

需要验证内容主题时，在 `expected.content_quality` 中声明语义验收标准：`deck_thesis` 是听众应接受的中心判断，`required_narrative_chain` 是必须按页承接的论证链，`max_consecutive_same_layout` 限制连续同类叙事页面。需要目录页时增加 `agenda_subtitles`，要求 `toc_item.title` 保存章节名、`toc_item.body` 保存独立副标题。`pptbench` 会在 `model_output.json` 写入 `content_quality`：它列出 deck/逐页主张、缺失主张页、完全重复主张、连续同类布局以及 `agenda_subtitle_issues`，供 Judge 结合 case 做语义评分；该报告不是替代人工判断的关键词匹配器。

`reviewer` case 应只放一个或少数明确缺陷，避免混入无关错误。否则无法判断 Reviewer 到底修复了目标问题，还是被其它问题干扰。

`fixer` case 必须写清授权页面和禁止改动范围。推荐在 `input.allowed_page_indexes` 明确授权页，避免 benchmark 依赖页码推断。

`router` case 应区分创建入口和继续任务。`has_existing_task=false` 表示创建入口；`has_existing_task=true` 表示已有任务上的继续请求。每轮全量评测应覆盖四个核心意图：`create_deck`、`fix_existing`、`fix`、`regenerate_all`。

## 开发集与验证集

既有 `test` 集保留在原路径，并继续用于回归门禁。新建的 `validation` 集使用不同主题、数据、页面角色和局部修改字段，作为独立 holdout；不得根据其中某一条失败直接改 prompt、fixture 或 rubric。

每次 Agent 修复先跑既有 test 集；通过后再完整跑一次 validation：

```powershell
go run ./cmd/pptbench --dataset validation -s all -p all -o ../benchmark/runs/20260829-150405-validation-all
```

验证失败时，先按输出归因到 Agent、case、Judge 或环境。若确认是 Agent 缺陷，在既有 test 集中新增一个与验证 case 不同的回归 case 后修复，再重新运行整个 validation 集；不要修改已经执行过的 validation case 来抬高分数。需要扩大覆盖面时只追加新的 validation case，并在评测记录中注明版本和新增原因。

## 常见问题

### `agent_error` 出现 402

说明模型 provider 返回额度、权限或计费错误。Benchmark 会记录错误并继续落盘，这不是 case JSON 或 CLI 编译问题。

### `judge_error` 提示 Judge key 未配置

检查 `.env` 中是否有 `PPT_BENCH_JUDGE_API_KEY`。这把 key 只用于 benchmark 的 Agent 和 Judge 初始化，不进入生产服务或其它测试。

### `-p model` 的 summary 没有平均分

这是预期行为。`-p model` 不调用 Judge，所以 `summary.md` 会显示 `not judged`。

### Reviewer case 出现很多非目标错误

优先修 case 输入。Reviewer benchmark 应隔离目标缺陷，否则评分会混乱。

### Planner benchmark 很慢

Planner 会调用真实 Agent 和工具，属于高成本 suite。调试时先用：

```powershell
go run ./cmd/pptbench -s planner --case planner_business_kpi_001 -p model
```

## 本地验证命令

修改 `pptbench` 或 benchmark 入口后，至少运行：

```powershell
go test ./cmd/pptbench ./pkg/agent/deck ./pkg/web
go run ./cmd/pptbench -s router -p model -l 1
```

涉及 Planner/Reviewer/Fixer 调用逻辑时，再分别跑单 case smoke。
