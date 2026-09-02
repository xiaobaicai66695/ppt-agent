## ADDED Requirements

### Requirement: Dashboard delivery feedback entry
The Dashboard SHALL show a compact delivery-feedback control in its lower-left workspace area for an eligible delivered task, with a 1–5 rating and optional suggestion input.

#### Scenario: Finished PPT is opened
- **WHEN** the owner selects a completed task with delivery files
- **THEN** the Dashboard displays the feedback control without obscuring download or preview actions

#### Scenario: Feedback was previously submitted
- **WHEN** the task response contains saved feedback
- **THEN** the Dashboard restores the rating and suggestion and allows the owner to update them
