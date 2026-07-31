## ADDED Requirements

### Requirement: Local Runtime Event Log
The system SHALL append structured runtime events to `runtime_events.jsonl` inside the task work directory.

#### Scenario: Task records local events
- **WHEN** a PPT task runs
- **THEN** the task work directory contains `runtime_events.jsonl`
- **AND** each line is a JSON object with task id, event kind, timestamp, phase, agent or component name, status, and metadata when available

### Requirement: Runtime Report
The system SHALL write a `runtime_report.json` summary when the task finishes, fails, or is cancelled.

#### Scenario: Completion writes report
- **WHEN** task execution reaches a terminal state
- **THEN** `runtime_report.json` exists in the work directory
- **AND** it includes the final runtime snapshot and aggregate event counts

### Requirement: Current Task Timeline
The system SHALL expose recent runtime events in `runtime_meta` SSE snapshots and render them in the Dashboard for the currently selected task.

#### Scenario: Timeline updates during generation
- **WHEN** the Dashboard receives runtime metadata with recent events
- **THEN** it renders a timeline for the active task
- **AND** the timeline can show event kind, phase, name, status, elapsed time, and detail without requiring historical task replay
