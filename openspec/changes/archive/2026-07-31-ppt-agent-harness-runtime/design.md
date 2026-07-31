## Context

The PPT Agent spans Go orchestration, Eino ADK callbacks, task lifecycle SSE, Python PPT generation, QA tools, and a Vue dashboard. Current diagnosis depends on logs, natural-language summaries, and final files. That makes repeated tool loops, context compression loss, missing slide artifacts, and poor QA outcomes difficult to see while a task is running.

The user also identified harness design ideas from an external AI Agent engineering document, especially metadata counting. The implementation should therefore create low-friction runtime observability first, then use those signals for budget warnings, handoff, validation, and UI exposure.

## Goals / Non-Goals

**Goals:**

- Keep one per-task runtime metadata object that can be updated by callbacks, tools, task manager, and compressors.
- Inject a short status bar into model calls without changing existing prompt templates.
- Record budget pressure as warnings first, leaving hard stop behavior for follow-up tuning.
- Emit runtime metadata through SSE and display it in Dashboard for developer inspection.
- Add an offline PPT quality eval harness that can be expanded into Pass^k and Best@k experiments.
- Validate manifest and slide output presence as structured data.

**Non-Goals:**

- Do not replace the full planning or slide generation architecture.
- Do not introduce a new queue, database schema, or external observability service.
- Do not run expensive LLM evals as part of the default local verification.
- Do not make budget warnings silently abort user tasks in this first slice.

## Decisions

### Decision 1: RuntimeMeta lives in backend agent utils

`RuntimeMeta` is placed in `ppt-agent/backend/pkg/agent/utils` so DeepAgent, compressors, callbacks, tools, and task manager can share one implementation without creating a dependency cycle. It exposes explicit mutator methods and immutable snapshots for SSE/API usage.

Alternative considered: keep metadata only in task manager. That would make callbacks and compression unable to update state directly, forcing fragile log parsing or duplicated counters.

### Decision 2: Model wrapper appends status as a user message

`RuntimeStatusChatModel` wraps existing chat models and appends a generated `<agent_status>` block per call. This keeps prompt templates stable and preserves existing fallback model behavior.

Alternative considered: edit every prompt template. That would be more invasive and harder to keep consistent across DeepAgent, SlideExecutor, Reviewer, and Fixer.

### Decision 3: Budget enforcement starts as warnings

The first implementation records warning thresholds for tool calls, same-argument repetition, elapsed runtime, token usage, slide progress, and QA pressure. Warnings are surfaced to agents and UI, while hard stop/change-strategy policies remain explicit future work.

Alternative considered: fail tasks immediately when thresholds trip. That is risky before enough runtime traces exist to tune thresholds safely.

### Decision 4: Structured handoff is embedded in compression output

Compression already produces handoff JSON. The change adds runtime snapshots and compression counters to that payload, preserving execution state across context reductions without replacing the compressor API.

Alternative considered: create a separate handoff file contract immediately. That is useful later but larger than needed for this first harness slice.

### Decision 5: Eval harness is offline and artifact-based

The evaluation script reads fixture cases and generated work directories, checking manifest, slide files, QA summaries, and rubric metadata. This makes it useful in CI/local development without requiring model credentials.

Alternative considered: LLM-as-judge scoring in the first pass. That would be closer to visual quality evaluation, but less deterministic and harder to run by default.

## Risks / Trade-offs

- [Risk] Runtime metadata may double-count tokens if both callbacks and explicit trackers report the same call. → Mitigation: keep tracker-derived totals authoritative when available and use callbacks mainly for visibility; verify behavior in tests/build.
- [Risk] Status injection increases prompt tokens. → Mitigation: keep the status bar compact and inject only when runtime metadata exists.
- [Risk] SSE clients may not expect the new event type. → Mitigation: add a new optional event type and keep old progress/complete events unchanged.
- [Risk] Budget warnings may be mistaken for hard enforcement. → Mitigation: name them warnings in code/UI and document hard stops as future work.
- [Risk] Offline eval cannot prove visual beauty. → Mitigation: make it a regression harness for artifact completeness and rubric scaffolding, not the final visual judge.

## Migration Plan

1. Add runtime metadata structs and model wrapper.
2. Pass runtime metadata through task creation and agent configs.
3. Wire callbacks, QA tool, compression, and manifest validation into metadata updates.
4. Emit metadata over SSE and add frontend types/UI.
5. Add offline eval fixtures/script.
6. Run targeted Go, frontend, and Python verification.

Rollback is straightforward because runtime metadata is additive: remove wrapper/config wiring and ignore the optional SSE event if needed.

## Open Questions

- What default budget thresholds best match real production traces?
- Should budget warnings later become hard stops, strategy changes, or human confirmation gates?
- Should visual scoring eventually use screenshot rendering plus an LLM judge, a deterministic layout analyzer, or both?
