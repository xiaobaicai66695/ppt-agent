## Why

Long decks currently ask the Planner to produce a full executable `tasks.json` in one model turn. As page count, background image planning and component contracts grow, later pages often collapse into generic filler because the model must spend too much attention on schema and engineering fields instead of page-specific evidence.

The runtime context compression path is also too quiet: when conversation context approaches the model window, compression should happen before the next agent call and surface a clear progress event to the workbench instead of failing late or invisibly degrading quality.

## What Changes

- Split initial deck planning into a blueprint phase, parallel section planning shards, deterministic merge, and one final Task Reviewer quality gate.
- Add stable section/page metadata so later user feedback can target one page or one section without regenerating the full deck.
- Move deterministic task fields such as `task_id`, `output_file`, `status`, capacity counts and background keyword normalization out of section-planner output where possible.
- Add a visible context compression stage that sends prior history, the previous compression summary and the emphasized user request to the model when the conversation nears the configured context limit.
- Emit progress/conversation events for context compression so the frontend clearly shows that history was compressed before work continues.
- Keep the user-facing workflow as a single generation action; the internal split is not a corrective retry loop.

## Capabilities

### New Capabilities
- `ppt-agent-chunked-planning`: Covers blueprint planning, section-shard planning, deterministic manifest merge and page/section metadata for granular repair.
- `ppt-agent-context-compression-visibility`: Covers preflight context compression behavior and user-visible compression progress events.

### Modified Capabilities
None.

## Impact

- Backend Agent orchestration under `ppt-agent/backend/pkg/agent/deck`.
- Planner, Reviewer and possibly Fixer prompts under `ppt-agent/backend/pkg/prompts`.
- Manifest schema/types and mutation tools for blueprint/section metadata.
- SSE progress or conversation event payloads consumed by the frontend workbench.
- Frontend status rendering where compression progress should be visible.
- Focused backend tests for manifest merge, section planner orchestration and compression event emission.
