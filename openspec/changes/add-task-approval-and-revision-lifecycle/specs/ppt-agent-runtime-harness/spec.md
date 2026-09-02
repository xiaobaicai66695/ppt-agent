## ADDED Requirements

### Requirement: Approval and recovery metadata is observable
The runtime metadata snapshot SHALL expose the current approval and recovery phase without depending on model natural-language output.

#### Scenario: Pending approval appears in runtime metadata
- **WHEN** a task is paused for approval
- **THEN** runtime metadata reports `waiting_approval` and references the pending approval identifier

