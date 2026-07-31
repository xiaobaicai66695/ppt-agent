## Context

The confirmed scope is intentionally local: every task writes its own event log and report in `work_dir`; the Dashboard only displays the current task's recent events. A global event index and historical task timeline are explicitly out of scope for this slice.

## Goals / Non-Goals

**Goals:**

- Add an append-only JSONL event trail.
- Keep event payloads small and scrubbed to previews/metadata.
- Derive a final report from the runtime snapshot and event counters.
- Add a compact Dashboard Timeline under the developer status panel.

**Non-Goals:**

- No OpenTelemetry exporter.
- No cross-task global index.
- No historical cold-load timeline.
- No full prompt or full tool output persistence.

## Decisions

### Decision 1: Extend RuntimeMeta

Runtime events are recorded through `RuntimeMeta` so callbacks, task manager, compressors, and tools share one sink. `RuntimeMeta` keeps a small recent-event ring for SSE and appends all events to JSONL.

### Decision 2: Local Files First

The work directory already contains task artifacts and manifests. Storing `runtime_events.jsonl` and `runtime_report.json` there makes debugging self-contained and avoids database migrations.

### Decision 3: Timeline Is Current-Task Only

The frontend consumes recent events from live `runtime_meta` payloads. Historical replay can be added later from the JSONL file if needed.

## Risks / Trade-offs

- [Risk] Event logs can grow large. -> Mitigation: store only compact metadata and keep SSE recent events capped.
- [Risk] Tool args may contain sensitive content. -> Mitigation: truncate previews and prefer event metadata over full payloads.
- [Risk] Report write can fail on filesystem errors. -> Mitigation: log warnings and do not fail the user task solely because report writing failed.
