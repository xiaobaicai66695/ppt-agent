## 1. Runtime Metadata and Agent Status

- [x] 1.1 Add `RuntimeMeta` structs, snapshot APIs, budget warning calculation, slide progress, QA, manifest, and compression counters.
- [x] 1.2 Wrap DeepAgent, SlideExecutor, Reviewer, and Fixer model calls with runtime status injection.
- [x] 1.3 Record LLM token usage, tool starts, tool errors, and repeated tool-argument streaks through callbacks.

## 2. Handoff, Validation, and Tool Signals

- [x] 2.1 Attach runtime metadata to compressor handoff payloads and record compression statistics.
- [x] 2.2 Record QA issue counts into runtime metadata from the QA tool.
- [x] 2.3 Validate generated task manifests and slide artifact presence during polling and completion.

## 3. Task Lifecycle and SSE Contract

- [x] 3.1 Create and pass runtime metadata through task manager, agent config, continue-mode execution, and task state.
- [x] 3.2 Emit `runtime_meta` SSE events without breaking existing progress, message, error, and complete events.

## 4. Frontend Developer Status

- [x] 4.1 Add TypeScript types for runtime metadata SSE events.
- [x] 4.2 Render a Dashboard developer status panel with phase, elapsed time, tool/error totals, token counts, QA count, slide progress, warnings, and manifest validation.

## 5. PPT Quality Eval Harness

- [x] 5.1 Add representative PPT quality eval cases with input prompts, expected artifacts, constraints, and rubrics.
- [x] 5.2 Add an offline evaluator script that checks generated work directories and emits structured JSON results.

## 6. Verification and Tracking

- [x] 6.1 Run Go formatting and targeted backend verification.
- [x] 6.2 Run frontend build and Python eval script syntax verification.
- [x] 6.3 Update `docs/issues/todo.md` with OpenSpec links, final status, and verification results.
