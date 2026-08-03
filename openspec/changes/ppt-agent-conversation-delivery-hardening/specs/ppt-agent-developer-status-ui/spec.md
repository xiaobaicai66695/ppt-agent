## MODIFIED Requirements

### Requirement: Dashboard Developer Status Panel
The Dashboard SHALL render an intent alignment panel for the selected or active task, including the compact user intent anchor, frozen planning snapshot, current phase and slide, structural alignment status, deviation warnings, elapsed time, tool and error totals, token usage, budget warnings, slide progress, and manifest validation status.

#### Scenario: Runtime status is visible during generation
- **WHEN** the Dashboard receives a runtime metadata event
- **THEN** it updates the intent alignment and runtime details without interrupting normal task progress display
- **AND** missing data is rendered as an explicit pending or unavailable state rather than causing a UI error

#### Scenario: Agent diverges from frozen plan
- **WHEN** runtime metadata reports a page-count, page-identity, title, content-type, or output-file deviation from the frozen plan
- **THEN** the Dashboard identifies the affected planning or execution step
- **AND** shows the expected and observed values when available

#### Scenario: Budget and validation warnings are visible
- **WHEN** runtime metadata contains budget warnings, alignment warnings, or missing generated files
- **THEN** the Dashboard surfaces those warnings in the status panel
- **AND** the normal download, preview, and conversation flow remains available
