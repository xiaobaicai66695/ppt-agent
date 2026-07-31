## Why

The PPT Agent already generates decks, but failures are hard to diagnose because agent state, token usage, repeated tool calls, QA results, file progress, and compression loss are scattered across logs or natural-language summaries. Improving the harness gives both the agent and the frontend a reliable view of runtime metadata, budget pressure, validation evidence, and PPT quality signals.

## What Changes

- Add a runtime metadata model that tracks phase, elapsed time, tool/LLM usage, token estimates, compression statistics, QA counts, slide/file progress, and budget warnings.
- Inject a compact agent status bar into model calls so agents can reason from fresh execution state instead of relying only on prose history.
- Upgrade compression handoff with structured runtime state to reduce loss of key task, QA, and budget context.
- Expose runtime metadata over SSE and render it in the Dashboard as a developer status panel.
- Add an offline PPT quality eval harness with sample cases and rubric scoring hooks.
- Validate output manifests and generated slide artifacts so result state is based on observable files rather than optimistic completion text.

## Capabilities

### New Capabilities

- `ppt-agent-runtime-harness`: Runtime metadata, budget observability, structured handoff, and validation contracts for PPT Agent execution.
- `ppt-agent-quality-eval`: Offline PPT quality evaluation cases and scoring harness for generated deck artifacts.
- `ppt-agent-developer-status-ui`: Frontend developer status surface for runtime metadata and budget/QA signals.

### Modified Capabilities

<!-- No existing OpenSpec specs are present in this repository yet. -->

## Impact

- `ppt-agent/backend/pkg/agent/deep`: pass runtime metadata through DeepAgent, SlideExecutor, Reviewer, Fixer, and continue-mode flows.
- `ppt-agent/backend/pkg/agent/utils`: add runtime metadata model, model wrapper, and structured compression handoff fields.
- `ppt-agent/backend/pkg/callback`: record LLM/tool usage into runtime metadata.
- `ppt-agent/backend/pkg/task`: create runtime metadata per task, emit SSE updates, and validate output manifests.
- `ppt-agent/backend/pkg/tools/qa`: feed QA issue counts into runtime metadata.
- `ppt-agent/frontend/src`: add runtime metadata types and Dashboard developer status UI.
- `docs/eval` and `ppt-agent/scripts`: add PPT quality evaluation fixtures and CLI harness.
- `docs/issues/todo.md`: track the seven harness items and link them to this OpenSpec change.
