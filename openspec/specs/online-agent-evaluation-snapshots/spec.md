# online-agent-evaluation-snapshots Specification

## Purpose

Capture durable, private production evidence at PPT agent workflow boundaries for later reviewed evaluation use.

## Requirements

### Requirement: Stage-level production evidence is captured

The system SHALL create immutable evaluation stage snapshots for Planner, Reviewer, PlannerRefiner, and Fixer execution boundaries. Each snapshot SHALL identify its source task, stage, attempt, canonical input artifact, canonical output artifact, and execution provenance without storing credentials.

#### Scenario: Planner draft is produced

- **WHEN** the Planner writes a structured draft for a production task
- **THEN** the system records a Planner stage snapshot containing the normalized task context and `tasks.draft.json` evidence
- **AND** the snapshot remains addressable after the task work directory is later removed

#### Scenario: Reviewer and Refiner revise a draft

- **WHEN** the Reviewer returns advice and the PlannerRefiner applies an authorized repair
- **THEN** the system records separate Reviewer and PlannerRefiner snapshots
- **AND** the Refiner snapshot identifies the Reviewer snapshot and preserves the authorized page scope and resulting review evidence

#### Scenario: Fixer changes a delivered task

- **WHEN** a Fixer invocation receives a follow-up request against an existing plan
- **THEN** the system records the scoped plan input, follow-up context, requested scope, patch result, and post-change plan evidence

### Requirement: Evaluation artifacts are durable and integrity checked

The system SHALL store evaluation artifact metadata in the database and artifact bytes in dedicated evaluation storage. Every artifact SHALL have a SHA-256 checksum and redaction status, and stage snapshots SHALL not reference task work-directory paths as durable artifact locations.

#### Scenario: Artifact storage succeeds

- **WHEN** a stage snapshot is persisted
- **THEN** each referenced artifact is available by its immutable storage key and checksum
- **AND** duplicate artifact content is reused without creating divergent content records

#### Scenario: Artifact storage fails

- **WHEN** evaluation artifact persistence fails during a task workflow
- **THEN** the task delivery workflow continues according to its existing success or failure behavior
- **AND** the failure is recorded for operational diagnosis without reporting an evaluation snapshot as complete

### Requirement: Evaluation evidence is privacy gated

The system SHALL mark newly captured production evidence as pending review and SHALL prohibit export until an administrator has explicitly approved its redaction state. A withdrawn snapshot or candidate SHALL be excluded from future dataset publication and export.

#### Scenario: Unreviewed production evidence exists

- **WHEN** a stage snapshot has not been approved for redaction
- **THEN** an administrator can inspect its metadata in the candidate queue
- **AND** no dataset export includes its contents

#### Scenario: Evidence is withdrawn

- **WHEN** an administrator withdraws a snapshot or candidate
- **THEN** it cannot be added to a new dataset version
- **AND** future downloads reject bundles that include it
