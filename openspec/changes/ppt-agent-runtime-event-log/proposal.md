## Why

`RuntimeMeta` now exposes a useful current-state snapshot, but task failures still cannot be replayed as a timeline. Developers need a low-cost local execution trace that explains which agent, tool, phase, slide, file, or validation step caused slowdowns or bad output.

## What Changes

- Append runtime events to `work_dir/runtime_events.jsonl` for each task.
- Generate `work_dir/runtime_report.json` at task end.
- Include recent runtime events in `runtime_meta` SSE snapshots for the active task.
- Render a current-task Dashboard Timeline using the local event stream.
- Keep the first implementation local-only; no global index or external tracing service.

## Capabilities

### New Capabilities

- `ppt-agent-runtime-event-log`: Local event log, report, and current-task timeline for PPT Agent execution.

### Modified Capabilities

<!-- No existing capability requirements are modified; this change adds a dedicated trace capability. -->

## Impact

- `ppt-agent/backend/pkg/agent/utils`: runtime event model, JSONL appender, report writer, snapshot events.
- `ppt-agent/backend/pkg/callback`: record model/tool lifecycle events.
- `ppt-agent/backend/pkg/task`: record file, phase, manifest, completion events and write final report.
- `ppt-agent/frontend/src`: add runtime event types and current-task timeline UI.
- `docs/issues/todo.md` and `docs/xq-todo.md`: track confirmed decisions and implementation state.
