## 1. Backend event protocol

- [x] 1.1 Define first-class thought, tool call, tool result, and final answer event semantics with stable segment and call correlation fields
- [x] 1.2 Change task broadcasting to emit calls and results immediately in source order while deduplicating completions and retaining replay boundaries
- [x] 1.3 Forward real tool completion/error summaries from runtime callbacks and close only unresolved calls with the end-of-run fallback
- [x] 1.4 Emit safe thought summaries and incremental final answers for chat and PPT-agent paths without exposing private reasoning metadata

## 2. Frontend timeline

- [x] 2.1 Replace the nested phase/tool state with a flat arrival-ordered timeline reducer that supports incremental segments, replay deduplication, and legacy events
- [x] 2.2 Add a reusable accessible timeline item component for thought, tool call, tool result, final answer, execution, and error states
- [x] 2.3 Integrate the new event types into Dashboard SSE consumption, preserve manual scroll position, and retain traces after terminal events
- [x] 2.4 Reconcile DESIGN.md and UX-CONTRACT.md with the persistent ReAct timeline and safe-thought boundary

## 3. Verification and release

- [x] 3.1 Add backend tests for immediate ordering, repeated tool names, fallback completion, replay, and streamed chat protocol
- [x] 3.2 Add frontend tests for interleaving, incremental updates, expansion, legacy compatibility, terminal retention, and reconnect deduplication
- [x] 3.3 Run focused Go tests/build, frontend unit/build, DESIGN lint, Premium strict audit, anti-pattern search, and online SSE/API state checks
- [x] 3.4 Commit scoped changes, build and deploy Linux artifacts, verify the new process and health endpoint, run and clean a low-cost online streaming smoke task, and record evidence in done.md
