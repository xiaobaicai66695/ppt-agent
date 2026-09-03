# PPT Planner Benchmark

This benchmark uses the same Go test framework as the backend. It can run the
full planning workflow:

```text
query
→ NewPPTPlannerAgent
→ RunPPTPlannerWithCallback
→ tasks.draft.json
→ ReviewTasksDraftManifest / commit
→ tasks.json + tasks.review.json
→ optional Judge API
→ optional gold/render validation
```

## Low-Cost Default

The default test does not call external model APIs or render PPTX files. It
loads the five gold `tasks.json` files and verifies that they pass the backend
DeckSpec reviewer.

```powershell
cd D:\environment\codeGo\llm-examples\projects\ppt-agent\backend
go test ./test/plan_benchmark -v -run TestGoldDeckSpecsPassReviewer -count 1
```

## Generate Planner Artifacts

This calls the real Planner Agent and writes generated `tasks.json`,
`tasks.review.json`, event logs, and summary files under
`test/plan_benchmark/results/`.

```powershell
cd D:\environment\codeGo\llm-examples\projects\ppt-agent\backend
$env:PPT_BENCH_RUN_PLANNER="true"
$env:PPT_BENCH_LIMIT="1"
go test ./test/plan_benchmark -v -run TestPlannerWorkflowGeneratesReviewedDeckSpec -count 1 -timeout 20m
```

`PPT_BENCH_RUN_LIVE=true` is the unified real-model switch. It also enables
the real Planner and Judge tests below; use `PPT_BENCH_LIMIT` for a bounded
smoke run before evaluating the complete dataset.

## Judge API

By default, the Judge test scores the gold manifests. To score generated
Planner artifacts, set `PPT_BENCH_JUDGE_RESULTS_DIR` to a result run directory.

```powershell
cd D:\environment\codeGo\llm-examples\projects\ppt-agent\backend
$env:PPT_BENCH_RUN_JUDGE="true"
$env:PLAN_JUDGE_API_KEY="..."
$env:PLAN_JUDGE_MODEL="..."
go test ./test/plan_benchmark -v -run TestPlanJudgeAPIScoresDeckSpecs -count 1 -timeout 10m
```

## Gold Render Check

This copies each gold `tasks.json` into a temporary work directory and invokes
the production `RenderDeckByTaskIDWorkflow`, which calls
`skills/ppt-deck-planner/generators/render_task.py`.

```powershell
cd D:\environment\codeGo\llm-examples\projects\ppt-agent\backend
$env:PPT_BENCH_RUN_GOLD_RENDER="true"
go test ./test/plan_benchmark -v -run TestGoldDecksRenderWithSkillScripts -count 1 -timeout 10m
```

## Environment Variables

| Variable | Purpose |
| --- | --- |
| `PPT_BENCH_CASES` | Override benchmark case file. |
| `PPT_BENCH_LIMIT` | Limit number of cases. |
| `PPT_BENCH_RUN_LIVE` | Unified real-model switch for Planner and Judge benchmarks. |
| `PPT_BENCH_RUN_PLANNER` | Set to `true` to call the real Planner Agent. |
| `PPT_BENCH_RUN_JUDGE` | Set to `true` to call the Judge API. |
| `PPT_BENCH_RUN_GOLD_RENDER` | Set to `true` to render gold manifests. |
| `PPT_BENCH_JUDGE_RESULTS_DIR` | Score generated Planner artifacts under this result directory. |
| `PPT_BENCH_PLANNER_TIMEOUT` | Planner per-case timeout, for example `12m`. |
| `PLAN_JUDGE_API_KEY` | Judge API key. Falls back to `OPENAI_API_KEY`, then `ARK_API_KEY`. |
| `PLAN_JUDGE_MODEL` | Judge model. Falls back to `ARK_QA_MODEL`, `ARK_TEXT_MODEL`, then `ARK_MODEL`. |
| `PLAN_JUDGE_BASE_URL` | OpenAI-compatible base URL. Falls back to `OPENAI_BASE_URL`, `ARK_BASE_URL`, then `https://api.openai.com/v1`. |
| `PLAN_JUDGE_MIN_SCORE` | Minimum score. Defaults to `7.0`. |
