## 1. Delivery Identity And Backend Contract

- [x] 1.1 Add canonical output-file and logical-slide identity helpers with focused Go tests.
- [x] 1.2 Deduplicate task completion, continuation, refresh, and persisted file lists while preserving download compatibility.
- [x] 1.3 Add backend handler tests for relative/absolute duplicate files and stable completion payloads.

## 2. Conversation Lifecycle

- [x] 2.1 Change continuation POST handling to return accepted JSON promptly and publish all incremental events through the task SSE stream.
- [x] 2.2 Buffer and persist one assistant message per generation/continuation turn without punctuation-based splitting.
- [x] 2.3 Return a complete structured conversation for active, persisted, and legacy tasks with ownership checks and compatibility fallbacks.
- [x] 2.4 Add focused Go tests for accepted/queued continuation, message ordering, restart recovery, and Markdown chunk preservation.

## 3. Intent Alignment Observability

- [x] 3.1 Extend RuntimeMeta with user intent anchor, frozen plan summary, current slide, alignment status, and structured deviation warnings.
- [x] 3.2 Freeze the first valid manifest as the plan baseline and compare later progress/output state without invoking an extra model.
- [x] 3.3 Expose alignment metadata through SSE/report snapshots and cover snapshot/comparison behavior with tests.

## 4. Unified Frontend Workbench

- [x] 4.1 Add client utilities for compact task titles, canonical slide de-duplication, legacy conversation recovery, and safe Markdown rendering with tests.
- [x] 4.2 Build a reusable unified conversation composer for create, queue, and continue modes with preset/custom-template selection.
- [x] 4.3 Change ComposePage to save a versioned custom outline draft and return it to the Dashboard composer for explicit submission.
- [x] 4.4 Refactor Dashboard and Sidebar to use one composer, preserve conversation history while loading, and accumulate SSE chunks into one assistant turn.
- [x] 4.5 Rebuild completed-state flow, title details, slide counts, and responsive spacing so panels and controls do not overlap.
- [x] 4.6 Upgrade the diagnostics panel to show intent, planned outline, current execution point, and affected deviation step.

## 5. Portable Assets And Linux CLI

- [x] 5.1 Add manifest validation and deployment-like icon-grid smoke coverage for assets bundled under `ppt-agent`.
- [x] 5.2 Make icon rendering produce a visible semantic fallback when a manifest entry or image cannot be loaded.
- [x] 5.3 Remove Windows shell/Python branches and document/test the Linux-only CLI defaults and overrides.

## 6. Verification And Project Records

- [x] 6.1 Run focused Go tests/build, frontend unit/build checks, Python compile/asset smoke tests, and strict OpenSpec validation.
- [x] 6.2 Use Playwright at desktop/tablet/mobile widths to verify duplicate removal, long prompts, completed layout, restored Markdown chat, unified composer, and template draft handoff.
- [x] 6.3 Update iteration records and mark `PPT-UX-002` done with completion date, implementation links, and verification evidence.
