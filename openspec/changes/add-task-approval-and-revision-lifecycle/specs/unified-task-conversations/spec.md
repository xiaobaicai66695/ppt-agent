## ADDED Requirements

### Requirement: Continuation respects task approval and revision lifecycle
The system SHALL keep the same task identity for post-delivery modification while applying its effective approval policy and publishing accepted changes through the task's revision lifecycle.

#### Scenario: Manual approval before a continuation change
- **WHEN** an owner continues a completed task configured for `manual` approval
- **THEN** the task creates an approval request and pauses before invoking the modification workflow

