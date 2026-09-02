## ADDED Requirements

### Requirement: User-selectable task approval policy
The system SHALL persist an effective approval mode for every generation task. It SHALL support `auto`, `sensitive`, and `manual`; existing and newly created tasks without an explicit setting SHALL use `auto`.

#### Scenario: Automatic task execution
- **WHEN** a task uses `auto` approval mode
- **THEN** an otherwise approvable operation continues without entering `waiting_approval`

#### Scenario: Sensitive operation requires approval
- **WHEN** a task uses `sensitive` approval mode and the server classifies an operation as high impact
- **THEN** the task enters `waiting_approval` before the operation executes

### Requirement: Explainable durable approval interruption
The system SHALL persist an approval request before pausing a task. The request SHALL include a trusted reason code, user-readable reason, proposed action, affected scope, rejection consequence and available decisions.

#### Scenario: User reconnects while approval is pending
- **WHEN** a client reconnects to a task in `waiting_approval`
- **THEN** the client can retrieve the same unresolved approval request and its explanation

#### Scenario: User rejects an approval request
- **WHEN** the owner rejects a pending approval request
- **THEN** the proposed operation SHALL not execute and the currently active deck revision SHALL remain unchanged

### Requirement: Bounded recovery state
The system SHALL record recoverable runtime failures with a failure class and recovery action. It SHALL use `recovering` while executing automatic retry, fallback or targeted repair, and SHALL record `failed` only after recovery is exhausted or unavailable.

#### Scenario: Recoverable model failure
- **WHEN** a model call fails before producing usable output and a configured fallback remains
- **THEN** the task records a recovery event, enters `recovering`, and returns to `running` when fallback execution succeeds

