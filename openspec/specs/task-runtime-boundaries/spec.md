# task-runtime-boundaries Specification

## Purpose
TBD - created by archiving change refactor-runtime-module-boundaries. Update Purpose after archive.
## Requirements
### Requirement: Task runtime responsibilities SHALL be independently testable
Task state transitions, SSE event buffering, persistence, conversation state, and delivery reconciliation SHALL be implemented as separable units with explicit ownership of locks and callbacks.

#### Scenario: State transition behavior is unchanged
- **WHEN** a task starts, reports progress, completes, fails, is cancelled, or is resumed
- **THEN** the same state, timestamps, event IDs, and persisted task metadata are produced as before the refactor

### Requirement: SSE replay boundaries SHALL remain stable
The task runtime SHALL preserve monotonic event IDs, listener isolation, replay-after-cursor behavior, and terminal event delivery while moving stream code out of the manager's general lifecycle logic.

#### Scenario: Client reconnects after an event cursor
- **WHEN** a client subscribes with its last-seen event ID
- **THEN** it receives each subsequent event once, in order, and receives the same terminal event semantics as before

### Requirement: Conversation and generation locks SHALL remain explicit
Conversation-stream exclusivity and generation/continue mutual exclusion SHALL be owned by the state unit, while persistence and delivery synchronization SHALL never acquire those locks implicitly.

#### Scenario: Duplicate conversation stream is rejected
- **WHEN** a second conversation stream starts while the first stream is active
- **THEN** the second start is rejected using the existing conflict behavior without corrupting the first stream's state

