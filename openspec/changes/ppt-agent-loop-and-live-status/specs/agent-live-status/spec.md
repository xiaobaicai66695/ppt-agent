## ADDED Requirements

### Requirement: Task conversation recovery is idempotent
The system SHALL preserve a per-task event cursor and SHALL apply each SSE event at most once when users switch tasks, navigate away and return, or reconnect after an interruption.

#### Scenario: Return to a running task
- **WHEN** a user switches away from a running task and later returns
- **THEN** the client resumes after the last event already applied and existing AI messages are not duplicated

#### Scenario: Reconnect after transport interruption
- **WHEN** an SSE connection is interrupted and reconnects
- **THEN** only events newer than the last applied event ID are merged into the task view

### Requirement: Conversation snapshots expose an event boundary
The conversation API SHALL return the latest event ID associated with the in-memory task snapshot so a client can establish a replay boundary without inspecting internal event storage.

#### Scenario: Fetch running conversation
- **WHEN** a client fetches the conversation for an in-memory running task
- **THEN** the response contains `latest_event_id` as a non-negative integer

### Requirement: Users can see live agent activity
The task conversation UI SHALL display one concise live status describing whether the agent is preparing, thinking, reading, searching, generating slides, retrying, reconnecting, completed, or failed.

#### Scenario: Tool invocation is in progress
- **WHEN** the task emits a supported tool or phase event
- **THEN** the status line immediately shows a user-readable action without exposing raw tool arguments

#### Scenario: No event arrives for a short period
- **WHEN** a task remains running between visible answer chunks
- **THEN** the status line continues to indicate active processing and does not make the task appear frozen

### Requirement: User-facing chat excludes execution narration
The system SHALL keep raw tool operations and repetitive implementation narration out of persisted AI chat while preserving those details in the execution trace.

#### Scenario: Agent reads or writes internal files
- **WHEN** the agent invokes an internal file or manifest tool
- **THEN** the chat does not add a separate assistant paragraph describing each mechanical operation
