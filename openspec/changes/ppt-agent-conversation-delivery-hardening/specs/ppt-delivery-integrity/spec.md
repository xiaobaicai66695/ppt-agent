## ADDED Requirements

### Requirement: Stable Logical Slide Identity
The system SHALL derive a stable identity for every logical slide and SHALL render at most one preview card for that identity even when task progress and file events use different path forms.

#### Scenario: Relative and absolute paths identify the same slide
- **WHEN** a progress item references `1_cover.pptx` and a file event references an absolute path ending in `1_cover.pptx`
- **THEN** the delivery view renders one page 1 card
- **AND** the card becomes file-ready without creating a second entry

#### Scenario: Duplicate page indices are reported
- **WHEN** two output records claim the same valid page index
- **THEN** the system keeps one deterministic preview entry
- **AND** exposes a delivery warning describing the duplicate identity

### Requirement: Non-Overlapping Completed Layout
The completed task view SHALL keep progress, previews, completion status, diagnostics, and the conversation composer in coherent document flow with reserved space for any sticky control.

#### Scenario: Completed desktop task
- **WHEN** a completed task with at least 16 slides is shown on a desktop viewport
- **THEN** no completion panel, toolbar, preview card, diagnostic panel, or composer overlaps another interactive region

#### Scenario: Completed mobile task
- **WHEN** the same task is shown at a 375 pixel viewport width
- **THEN** controls wrap or stack without horizontal overflow
- **AND** the last slide and composer remain reachable by scrolling

### Requirement: Compact Task Display Title
The frontend SHALL display a deterministic compact title for long or Markdown-formatted task prompts while retaining access to the complete original prompt.

#### Scenario: Long prompt is displayed
- **WHEN** a task query exceeds the configured display length or contains multiple Markdown sections
- **THEN** the top bar and task heading show a single compact summary with ellipsis when needed
- **AND** the complete original query is available through an explicit details control

