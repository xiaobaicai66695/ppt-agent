# ppt-delivery-feedback Specification

## Purpose
TBD - created by archiving change expand-benchmark-and-collect-deck-feedback. Update Purpose after archive.
## Requirements
### Requirement: Owner can record delivery feedback
The system SHALL allow the owner of a completed task to save one task-level rating from 1 through 5 and an optional suggestion of at most 1000 characters.

#### Scenario: User submits a rating with a suggestion
- **WHEN** the owner submits a valid rating and optional suggestion for a completed task
- **THEN** the system persists the feedback and returns the saved value without starting a continuation or changing deck artifacts

#### Scenario: User updates a prior rating
- **WHEN** the owner submits feedback again for the same task
- **THEN** the system updates that owner's existing feedback record instead of creating a duplicate

### Requirement: Feedback is isolated and recoverable
The system SHALL return the owner's saved feedback with task data and SHALL not expose a user's feedback to a non-owner.

#### Scenario: Owner reloads a completed task
- **WHEN** the owner retrieves a task after submitting feedback
- **THEN** task data includes the saved rating, suggestion, and update timestamp

#### Scenario: Another user requests feedback for a foreign task
- **WHEN** a non-owner requests or submits task feedback
- **THEN** the system rejects the request without disclosing feedback content

### Requirement: Feedback input is validated
The system SHALL reject ratings outside 1 through 5, suggestions longer than 1000 characters, and feedback for tasks that have no completed delivery.

#### Scenario: Invalid feedback request
- **WHEN** a client submits invalid feedback data
- **THEN** the system returns a validation error and leaves any prior feedback unchanged

