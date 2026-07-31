# ppt-agent-developer-status-ui Specification

## Purpose
TBD - created by archiving change ppt-agent-harness-runtime. Update Purpose after archive.
## Requirements
### Requirement: Runtime Metadata SSE Event
The frontend API contract SHALL include a `runtime_meta` SSE event type carrying runtime metadata snapshots from the backend task manager.

#### Scenario: Backend emits runtime metadata
- **WHEN** a task is polled or completes
- **THEN** the SSE stream may include a `runtime_meta` event with phase, elapsed runtime, counts, warnings, and validation details
- **AND** clients that do not render the event can safely ignore it

### Requirement: Dashboard Developer Status Panel
The Dashboard SHALL render runtime metadata for the selected or active task, including phase, elapsed time, tool totals, error totals, QA issue count, token usage, budget warnings, slide progress, and manifest validation status.

#### Scenario: Runtime status is visible during generation
- **WHEN** the Dashboard receives a runtime metadata event
- **THEN** it updates a developer status panel without interrupting normal task progress display
- **AND** missing data is rendered as empty or zero state rather than causing a UI error

#### Scenario: Budget and validation warnings are visible
- **WHEN** runtime metadata contains budget warnings or missing generated files
- **THEN** the Dashboard surfaces those warnings in the status panel
- **AND** the normal download/preview flow remains available

