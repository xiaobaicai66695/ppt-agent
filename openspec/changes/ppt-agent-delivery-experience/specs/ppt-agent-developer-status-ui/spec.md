## MODIFIED Requirements

### Requirement: Dashboard Developer Status Panel
The Dashboard SHALL provide runtime metadata for the selected or active task, including phase, elapsed time, tool totals, error totals, QA issue count when available, token usage, budget warnings, slide progress, and manifest validation status, while keeping user-facing generation progress and slide previews visually primary.

#### Scenario: Runtime status is available during generation
- **WHEN** the Dashboard receives a runtime metadata event
- **THEN** it updates a collapsible diagnostics region without interrupting normal task progress display
- **AND** missing data is rendered as empty or zero state rather than causing a UI error

#### Scenario: Budget and validation warnings are visible
- **WHEN** runtime metadata contains budget warnings or missing generated files
- **THEN** the Dashboard surfaces a compact warning summary outside or on the diagnostics disclosure
- **AND** detailed runtime data remains available by expanding the diagnostics region
- **AND** the normal download and preview flow remains available
