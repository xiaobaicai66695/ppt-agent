# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
# Backend (Go 1.25.0)
cd backend
go mod tidy
go build ./...          # verify compilation
go build -o ppt-agent . # produce binary

# Frontend (React + TypeScript + Vite)
cd frontend
npm install
npm run dev             # dev server on port 3000, proxies /api → localhost:8080

# Tests
cd backend
go test ./pkg/agent/utils/ -v       # model fallback unit tests
go test ./pkg/tools/qa/ -v          # QA converter integration test
go test ./pkg/tools/search/ -v      # search integration test

# Run with env
export AGENT_MODE=deep              # "deep" (parallel) or empty (serial plan-execute)
export INTERACTIVE=false            # skip human-in-the-loop prompts
export DEEP_AGENT_CONCURRENCY=3     # max parallel slide generation (default 5)
```

## Architecture

### Two execution modes

1. **DeepAgent mode** (`AGENT_MODE=deep`, default): Uses `eino prebuilt/deep`. A master agent (`PPTTaskDeepAgent`) delegates to three sub-agents — `SlideExecutor` (generates slides via python-pptx), `Reviewer` (multimodal visual QA), `Fixer` (repairs QA issues). The master writes `tasks.json` and orchestrates the pipeline. Slides can be generated in parallel (controlled by `DEEP_AGENT_CONCURRENCY`).

2. **Plan-Execute-Replan mode** (legacy, serial): Uses `eino prebuilt/planexecute`. Three agents run sequentially — `Planner` → `Executor` → `Replanner` — in a loop until all slides are done. Each slide is generated one at a time with QA after each.

Both modes share the same tool implementations, skill injection, and model configuration. Prompts are duplicated between modes with ~70% overlap.

### Model fallback chain

`FallbackChatModel` (`pkg/agent/utils/model.go`) wraps multiple ARK models loaded from env:
- `ARK_MODEL` → `ARK_MODEL_BACKUP1` → `ARK_MODEL_BACKUP2`
- On 429 rate limit: pauses the failing model for 30s, continues to next backup
- All models fail → returns error
- The fallback chain is created fresh per agent (Planner, Executor, SlideExecutor, Fixer each get their own instance). The QA Reviewer gets a separate model via `QAModelFn`.

### Task lifecycle (DeepAgent mode)

The master agent writes `tasks.json` to the work directory. Each task progresses through statuses:
```
pending → done → qa_done → fixed
```
Helper methods on `TasksManifest` (`NeedsFix()`, `PendingTasks()`, `DoneTasks()`) drive the orchestration loop. QA results are stored per-task in `qa_report` fields. Fix attempts are capped at 2 per slide.

### Visual QA pipeline

`single_qa_review` tool (`pkg/tools/qa/qa_tool.go`):
1. Runs `pptx_qa_converter.py` which calls LibreOffice (PPTX→PDF) then pdftoppm (PDF→JPEG at 150 DPI)
2. Finds the image matching the requested PPTX filename stem
3. Sends the image + system prompt to a multimodal LLM (via `modelFn`)
4. Parses the response for `high`/`medium`/`low` severity issues
5. Merges results into `.qa_result.json`; tracks attempts in `.qa_attempts.json` (max 2/slide)

QA results use additive merging via `||` — once `HasHighIssue` becomes true it can never become false.

### Human-in-the-loop search approval

Search tools are wrapped via `InvokableSearchApprovalTool` (`pkg/tools/human_in_the_loop.go`). On first invocation, the tool calls `StatefulInterrupt` with a `SearchApprovalInfo`. The `human.Manager` loop catches the interrupt, prompts the user (1=skip, 2=confirm, 3=edit query), then resumes with `ResumeWithParams`. In non-interactive mode, all searches default to skip.

### Critical hardcoded paths

- **Python binary**: `/root/pptx_env/bin/python` — hardcoded in `python_runner.go:83` and `qa_tool.go:170`. The project only runs on Linux with python-pptx installed at this exact path.
- **Converter search**: `qa_tool.go` searches up 8 parent directories for `pptx_qa_converter.py` using relative paths, with fallback to `PROJECT_ROOT` env var.

### Prompt injection pattern

Skills are loaded from `skills/` directory (`SKILL.md` files) via `LoadSkillsFromDir` → `FormatSkillsForPrompt`, then injected as a `{skills}` template variable into Planner/Executor prompts. The skill middleware from `eino/adk/middlewares/skill` is initialized but never actually used — skills are injected manually as text.

### Prompt string patterns

`slide_executor.go` and `planexecute/executor.go` use a package-level `bt` constant (`"`"`) to embed backtick characters inside Go raw string literals. When adding markdown code formatting (backticks) to prompt strings, use `+ bt +` concatenation — never nest raw backticks inside a raw string literal.
