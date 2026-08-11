## MODIFIED Requirements

### Requirement: Runtime Metadata SSE Event
The frontend API contract SHALL include a `runtime_meta` SSE event type carrying a bounded recent-event tail and runtime metadata snapshot from the backend task manager.

#### Scenario: Backend emits runtime metadata
- **WHEN** a task is polled, streams progress, or completes
- **THEN** the SSE stream may include a `runtime_meta` event with phase, elapsed runtime, counts, warnings, validation details, and a bounded event tail
- **AND** clients merge the tail into already loaded history using stable event identity instead of replacing earlier events
- **AND** clients that do not render the event can safely ignore it

### Requirement: Dashboard Developer Status Panel
The Dashboard SHALL render runtime metadata for the selected or active task, including phase, elapsed time, tool totals, error totals, token usage, slide progress, complete tool activity history, and context-compression differences.

#### Scenario: Runtime status is visible during generation
- **WHEN** the Dashboard receives a runtime metadata event
- **THEN** it updates a developer status panel without interrupting normal task progress display
- **AND** missing data is rendered as empty or zero state rather than causing a UI error

#### Scenario: Tool activity remains visible after refresh or task switch
- **WHEN** persisted conversation events and SSE recent events overlap
- **THEN** the Dashboard merges them without duplicates
- **AND** early file, search, manifest, task, shell, Python, conversion, and other tool calls remain available in chronological order

#### Scenario: Budget and validation warnings are visible
- **WHEN** runtime metadata contains budget warnings or delivery validation failures
- **THEN** the Dashboard surfaces those warnings in the status panel
- **AND** the normal download/preview flow remains available

## ADDED Requirements

### Requirement: Compression timeline event
The Dashboard SHALL classify context compression as a distinct runtime event and render a concise before/after comparison.

#### Scenario: Compression event arrives
- **WHEN** a runtime event has kind `compression`
- **THEN** the timeline identifies it as context compression rather than a generic tool or other event
- **AND** its detail view compares message counts, token counts, removed messages, saved tokens or percentage, and preserved user requirements

#### Scenario: Older compression event lacks new fields
- **WHEN** a historical event contains only before and after token counts
- **THEN** the detail view renders the available comparison without failing
