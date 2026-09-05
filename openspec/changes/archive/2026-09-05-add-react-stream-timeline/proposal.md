## Why

The workbench already transports task activity over SSE, but it buffers each tool call until completion, merges the result into the call, and groups transient trace UI below assistant text. Users therefore cannot follow the actual interleaving of visible reasoning summaries, actions, observations, and the final answer while a multi-step task is running.

## What Changes

- Add an ordered, replayable SSE protocol for safe observable `thought`, `tool_call`, `tool_result`, `final_answer`, and `error` events.
- Emit a tool call immediately, then emit its result as a separate later event correlated by `tool_call_id`; the UI presents the pair as one collapsible card instead of a separate bottom result pile.
- Distinguish incremental visible process summaries from the incremental final answer without exposing private chain-of-thought.
- Replace phase-nested tool rendering with an arrival-ordered ReAct timeline whose individual entries can stream, expand, and survive task completion.
- Preserve legacy `answer`, `progress`, and lifecycle events during migration so existing task delivery and older clients continue to work.
- Add replay/deduplication, multi-turn, long-content, keyboard, responsive, and failure-path coverage.

## Capabilities

### New Capabilities

- `react-stream-timeline`: Defines the public event contract, ordering, replay, safe thought summaries, and workbench rendering behavior for an interleaved ReAct-style timeline.

### Modified Capabilities

- None.

## Impact

- Backend: `pkg/agent/ppt/run.go`, `pkg/runtime/task/manager.go`, `pkg/runtime/web/message_chat.go`, `pkg/runtime/web/streamer.go`, and focused tests.
- Frontend: SSE types/consumer, `conversationTimeline` state helpers, Dashboard rendering, a reusable timeline entry component, design/UX contracts, and focused tests.
- API compatibility: the existing task stream URL, SSE envelope, event IDs, terminal lifecycle events, and task REST responses remain stable; new event types and fields are additive.
- Operations: the runtime behavior change requires a Linux build, deployment, restart verification, and a low-cost online streaming smoke test.
