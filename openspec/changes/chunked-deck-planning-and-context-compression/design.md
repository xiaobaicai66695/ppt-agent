## Context

The current deck workflow creates one `PPTPlanner` agent and asks it to initialize a full executable `tasks.draft.json`. That single call must read skill files, choose narrative structure, plan page content, plan visual/background assets and emit a large schema-heavy manifest. As deck size grows, output quality degrades toward generic page filler, especially in later pages.

The runtime already has a chat model compressor, token tracking and runtime metadata. However the compressor is configured with a very high token threshold and no user-facing progress callback, so users do not see when compression happens and generation can still approach the model window before compression is useful.

## Goals / Non-Goals

**Goals:**
- Keep user-facing generation as one action while splitting internal initial planning into blueprint, section shards, deterministic merge and one Task Reviewer gate.
- Reduce LLM output burden by making section planners generate only page semantics and component content for a small page range.
- Preserve granular page and section metadata so later user fixes can target a page or section.
- Trigger context compression before model calls approach the window limit and show a concise progress event to the frontend.

**Non-Goals:**
- Do not add another corrective Reviewer loop.
- Do not change the PPT renderer contract beyond accepting optional section metadata and richer planning provenance.
- Do not implement multimodal visual inspection as part of this change.
- Do not require the frontend to understand every internal planning shard.

## Decisions

### Blueprint first, shard second

Initial planning will use a lightweight blueprint step that locks:
- deck title, theme and template fallback
- global narrative sections
- page count, page indexes, titles and `content_type`
- optional global content bank/evidence anchors

Section planners receive only the blueprint, shared content bank, their page range and neighbor summaries. They may fill `description`, `content_plan`, image intents and evidence references for their assigned pages, but they must not change page count, page index or section order.

Alternative considered: keep one Planner call and strengthen the prompt. This keeps implementation simple but leaves the model with the same large-output bottleneck.

### Deterministic merge owns engineering fields

The merger will assemble shard outputs into `tasks.draft.json` and own deterministic fields:
- `task_id`
- `output_file`
- `status`
- `capacity_hint.component_count`
- missing `created_at`
- same-`content_type` background keyword normalization

This keeps LLM shard output focused on content and avoids cross-shard overwrites.

### One final Task Reviewer gate

After merge, the existing deterministic review/report plus TaskPlanReviewer path remains the only quality gate. Section planning is not a retry loop; it is the first-draft construction strategy.

### Context compression becomes a visible preflight

Compression should have a lower practical threshold based on the configured planning model window. When compression triggers, the compressor builds its summary from:
- historical messages selected for compression
- the previous compression handoff if present
- preserved recent messages
- the current emphasized user request

The runtime emits a `progress` event with phase `compressing_context` before or during compression and records compression details in runtime metadata. Internal JSON summaries remain hidden from assistant chat output.

## Risks / Trade-offs

- [Risk] Shards may duplicate or contradict each other. → The blueprint locks page purpose and the merger validates page indexes, section ids and duplicate evidence refs.
- [Risk] Parallel section calls increase simultaneous model usage. → Use existing routing/concurrency limits and cap shard count by page count.
- [Risk] Some short decks do not need chunking. → Keep a threshold so small decks can use the existing monolithic path or a single shard.
- [Risk] Compression summarizer failure could block generation. → Fall back to deterministic anchors and emit a warning event while continuing with the uncompressed or partially compressed context.
- [Risk] Frontend status wording may expose internals. → Use concise user-facing phase text such as “正在压缩较早对话，保留你的最新要求”.

## Migration Plan

1. Add blueprint/section metadata structs and merge helpers behind the existing deck workflow.
2. Add blueprint and section planner agents/prompts/tools, then route eligible initial generation through chunked planning.
3. Add compression progress callback plumbing from compressor to task SSE events.
4. Update frontend status mapping for `compressing_context`.
5. Add focused unit tests and run backend package tests.
6. For deployment, use the standard backend build/deploy/smoke workflow because this changes runtime behavior.
