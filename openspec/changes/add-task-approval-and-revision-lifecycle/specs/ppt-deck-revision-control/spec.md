## ADDED Requirements

### Requirement: Candidate revision before change publication
The system SHALL create a candidate revision for a completed deck modification without overwriting the active revision. A candidate SHALL retain its parent revision, affected scope, manifest snapshot and change summary.

#### Scenario: User requests a deck modification
- **WHEN** an owner changes a completed deck
- **THEN** the active revision remains downloadable while the modification is stored as a candidate

### Requirement: User confirms or rejects candidate modifications
The system SHALL place a task in `waiting_confirmation` after a candidate modification is ready and SHALL offer the owner an accept or reject decision.

#### Scenario: User accepts a candidate
- **WHEN** the owner accepts a candidate revision
- **THEN** the candidate becomes active and the prior active revision becomes superseded

#### Scenario: User rejects a candidate
- **WHEN** the owner rejects a candidate revision
- **THEN** the candidate becomes rejected and the prior active revision remains active

### Requirement: Immutable rollback history
The system SHALL preserve superseded revisions and SHALL create a new active revision when the owner rolls back to a historical revision.

#### Scenario: User rolls back an active deck
- **WHEN** the owner selects a superseded revision as the rollback source
- **THEN** the system creates a new active revision from that source without mutating the historical source record

