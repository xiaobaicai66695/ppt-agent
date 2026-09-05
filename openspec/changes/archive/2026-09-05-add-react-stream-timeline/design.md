## Context

The task endpoint already uses SSE with monotonic event IDs, bounded replay, heartbeats, and explicit lifecycle terminal events. The ADK runner also already executes the model/tool loop. The ordering defect is introduced after execution: `TaskState.Broadcast` holds `tool_call` in `pendingTools`, waits for `tool_result`, and rewrites the pair into one terminal `tool_call`; the Dashboard then nests tools under a phase and removes those phases at completion.

The product contract prohibits exposing private chain-of-thought. The requested “thought” surface therefore represents safe, user-visible process narration: explicit assistant text that is already public, or deterministic stage summaries such as “正在检索并核实资料”. Raw `ReasoningContent`, hidden prompts, credentials, full tool payloads, and internal compression JSON remain excluded.

## Goals / Non-Goals

**Goals:**

- Deliver an event as soon as its corresponding visible step occurs.
- Preserve strict arrival order through live delivery, reconnect replay, and multi-turn use.
- Correlate calls and results without delaying backend events or duplicating UI cards.
- Stream safe thought/final-answer deltas into stable timeline entries.
- Keep the task stream URL, event ID cursor, heartbeat, task lifecycle, and delivery UI compatible.
- Provide keyboard-accessible disclosure, long-payload handling, auto-scroll that respects manual scroll-up, and narrow-screen stability.

**Non-Goals:**

- Exposing model chain-of-thought, `ReasoningContent`, system prompts, or unredacted runtime metadata.
- Replacing CloudWeGo Eino ADK with a custom tool loop.
- Persisting transient tool traces in MySQL conversation messages.
- Redesigning unrelated Dashboard navigation, delivery preview, or task CRUD.

## Decisions

### 1. Extend the existing SSE transport instead of adding WebSocket or a second endpoint

The existing `/api/tasks/:id/stream` endpoint already provides ordered one-way delivery, `Last-Event-ID` recovery, heartbeats, and nginx buffering controls. New first-class events use the same envelope and monotonic task-local `id`:

| Type | Required payload | Semantics |
|---|---|---|
| `thought` | `segment_id`, `content`, optional `phase`, `delta` | Safe public process narration; repeated deltas append only within the same segment. |
| `tool_call` | `tool_call_id`, `tool_name`, `tool_args`, optional `segment_id` | A call has started or is about to execute and is emitted immediately. |
| `tool_result` | `tool_call_id`, `tool_name`, `tool_result`, `tool_status`, optional preview | A later observation correlated to exactly one call; clients update the matching visible call card. |
| `final_answer` | `segment_id`, `content`, `delta` | User-facing answer text; deltas append within one answer segment. |
| `error` | `error`, optional correlation/phase fields | Recoverable or terminal failure detail safe for the user. |

Lifecycle events (`answer_end`, `complete`, `continue_complete`, `conversation_complete`) remain unchanged. The frontend continues accepting legacy `answer`, `system_step`, and merged `tool_call` frames during rolling deployment.

WebSocket was rejected because the flow is server-to-client after a REST mutation, and the existing replay cursor already solves the required reliability problem.

### 2. Assign IDs before any aggregation and keep correlation separate from presentation

`Broadcast` assigns an event ID and appends the event to the replay buffer before notifying listeners. It no longer delays `tool_call` or rewrites `tool_result`. `pendingTools` is retained only as a correlation/fallback registry, not as an event buffer. A result with no provider call ID resolves the latest compatible pending call by name; a duplicate completion is suppressed. Unresolved calls receive one fallback result at agent termination.

This keeps source ordering authoritative and lets the UI render an append-only timeline. Updating a streaming segment's text is a view-state operation keyed by `segment_id`, not a mutation of event order.

### 3. Use runtime tool callbacks for real completion timing

The ADK agent event exposes model tool calls but does not reliably expose a separate result event across providers. The existing callback/`RuntimeMeta` path does observe actual tool start/end/error. The task event sink converts public, redacted tool end/error summaries into `tool_result`; raw results remain in runtime detail storage. The existing end-of-run fallback closes only calls still pending.

Alternative “mark every call complete after the whole agent run” was rejected because it makes observations arrive too late and destroys interleaving.

### 4. Treat `thought` as safe narration, never hidden reasoning

Ordinary chat emits deterministic analysis/answer-stage summaries around optional retrieval tools. PPT planning maps already-visible planner narration and phase changes to `thought`. `ReasoningContent` and `reasoning_preview` are not copied into SSE. This preserves the useful ReAct mental model without claiming to reveal private internal reasoning.

### 5. Use a flat frontend event model with paired tool cards

`conversationTimeline` becomes an ordered union of message, thought, paired tool call/result, final answer, execution, and delivery items. Reducer helpers accept the server event ID as the stable key and track processed IDs for reconnect deduplication. Backend `tool_call` and `tool_result` remain separate source events, but the UI owns their presentation as one collapsible card keyed by `tool_call_id`: the call appears immediately in source order, and the later result updates that same card's status and internal result panel instead of creating a second row.

The signature visual is a quiet vertical “execution thread”: a thin connector and compact semantic cards. Thought is a subdued disclosure, tool cards use running/success/error state text, the result panel is scrollable for long safe content without truncating it, and the final answer keeps the established assistant bubble. This extends current runtime tokens rather than introducing a new palette.

### 6. Preserve user scroll intent and accessibility

New content auto-scrolls only while the viewport is already near the bottom. A user who scrolls upward is not pulled back until they return near the bottom. Disclosures use native buttons with `aria-expanded`/`aria-controls`; the timeline uses `aria-live=polite` without making every token assertive. Motion is limited to the existing running indicator and a short entry fade disabled under reduced motion.

## Risks / Trade-offs

- **Provider callbacks omit or reorder a tool completion** → Correlate by provider ID when available, then name/argument preview; emit one fallback result only for still-pending calls at run end and cover repeated-name cases in tests.
- **Duplicate tool activity from ADK events and RuntimeMeta** → Use ADK events as the call source and RuntimeMeta end/error as the result source; suppress generic runtime rows for first-class tool events.
- **Rolling frontend/backend versions disagree on event names** → New frontend accepts both protocols; endpoint and terminal lifecycle stay stable. Deploy backend and frontend together.
- **Thought text could be mistaken for private reasoning** → Label it “思考摘要”, document the boundary, and only emit allow-listed public narration/stage summaries.
- **Long tool payloads hurt rendering or leak data** → Send safe redacted result text, format JSON when valid, keep the full safe content inside a bounded scrollable panel, and never append it as a detached bottom row.
- **Transient trace is lost after a cold restart** → Keep the current bounded in-memory/Redis trace policy; durable MySQL chat history remains user/assistant only. This change improves live/reconnect ordering, not long-term audit retention.

## Migration Plan

1. Add backend event constants/fields, immediate broadcast behavior, callback result forwarding, and tests.
2. Ship frontend dual-protocol consumption and the flat timeline UI with updated contracts/tests.
3. Run focused Go tests, frontend tests/build, Premium audit, anti-pattern search, and online SSE/API state checks.
4. Build the Linux artifact, deploy backend/frontend together, restart, verify health and process identity, then run one low-cost chat request that uses retrieval and confirms event order.
5. Roll back by redeploying the previous commit; no database migration is required.

## Open Questions

None. The existing architecture and product contracts define the security, persistence, and transport boundaries needed for implementation.
