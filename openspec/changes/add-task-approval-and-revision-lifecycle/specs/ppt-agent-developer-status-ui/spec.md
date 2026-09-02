## ADDED Requirements

### Requirement: Dashboard renders actionable approval interruption
The Dashboard SHALL render a pending approval with its reason, action summary, affected scope and decisions without treating it as a failed task.

#### Scenario: Task enters waiting approval
- **WHEN** the selected task reports `waiting_approval`
- **THEN** the Dashboard presents approve, adjust-scope and reject controls with the server-provided explanation
