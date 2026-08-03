## ADDED Requirements

### Requirement: Incremental SSE replay
The task event stream SHALL assign stable increasing event IDs and SHALL replay only events newer than the client's last received event ID.

#### Scenario: SSE connection reconnects
- **WHEN** EventSource reconnects with a `Last-Event-ID` header
- **THEN** the backend replays only newer buffered events
- **AND** the frontend does not duplicate already rendered logs or progress

### Requirement: Complete task event contract
The frontend and backend SHALL share event types for system steps, progress, files, thumbnails, runtime metadata, errors, and terminal states.

#### Scenario: Planning system step is emitted
- **WHEN** the backend emits a `system_step` event before page generation
- **THEN** the Dashboard renders a concise user-facing activity message

#### Scenario: Thumbnail conversion completes
- **WHEN** a generated slide receives a ready or failed thumbnail result
- **THEN** the Dashboard updates that slide preview without waiting for the whole deck to complete

### Requirement: Background thumbnail generation
The server SHALL start thumbnail preparation after each PPTX file becomes available without blocking task progress event polling.

#### Scenario: Single slide PPTX appears
- **WHEN** TaskManager discovers a new PPTX file
- **THEN** it emits `file_ready` immediately
- **AND** schedules thumbnail preparation in the background
- **AND** emits `thumbnail_ready` or `thumbnail_error` when preparation finishes

#### Scenario: Browser and background generation overlap
- **WHEN** a browser requests a thumbnail while background preparation is already running
- **THEN** conversion is serialized and the PPTX is not converted twice after the first result reaches disk

### Requirement: Lightweight preview loading
The frontend SHALL load thumbnail media progressively and SHALL NOT pre-download complete PPTX files without an explicit download action.

#### Scenario: User watches an active task
- **WHEN** slide files arrive
- **THEN** fixed-aspect placeholders appear immediately
- **AND** thumbnail images load lazily as they become visible
- **AND** full PPTX transfer starts only after a user download action

### Requirement: Reliable terminal state recovery
Polling SHALL synchronize the selected task summary and stop only after the UI has received the terminal status and latest files/counts.

#### Scenario: Final SSE event is lost
- **WHEN** polling observes completed, failed, or cancelled status after an SSE interruption
- **THEN** the Dashboard updates the task list and selected task to that terminal state
- **AND** no running indicator remains stuck

### Requirement: Responsive progress timing
The progress component SHALL remain visible during planning and SHALL update elapsed time reactively while a task is running.

#### Scenario: Total page count is not known yet
- **WHEN** a task is running with total count zero
- **THEN** the progress area shows the current phase and an indeterminate indicator

#### Scenario: No page event arrives for several seconds
- **WHEN** a running task remains in the same phase
- **THEN** elapsed time continues updating at least once per second
