## ADDED Requirements

### Requirement: Manifest theme is deck-wide
The system SHALL use the task manifest `theme` as the palette for every generated slide and SHALL NOT replace it based on an individual slide background.

#### Scenario: Background page keeps selected palette
- **WHEN** a task uses theme `charcoal_light` and a slide has background `party_government`
- **THEN** the SlideExecutor uses `charcoal_light` for that slide and does not switch to `government_red`

### Requirement: Background use follows explicit task policy
The system SHALL pass the backend-selected background policy into DeepAgent planning and SHALL preserve an empty background when background use is disabled.

#### Scenario: Preset selected without background
- **WHEN** a user selects a preset whose task outline has background use disabled
- **THEN** the planner leaves empty slide backgrounds empty and does not infer a background from the topic

#### Scenario: Recommendation selects background
- **WHEN** intelligent recommendation enables a named background
- **THEN** the planner applies only that named background to eligible pages without changing the selected theme

### Requirement: Generator output path follows the manifest
The SlideExecutor SHALL save each slide using the corresponding manifest `output_file` exactly as declared.

#### Scenario: Title contains whitespace-sensitive text
- **WHEN** a task title could be reconstructed with different whitespace
- **THEN** the generated PPTX path still exactly matches `output_file`
