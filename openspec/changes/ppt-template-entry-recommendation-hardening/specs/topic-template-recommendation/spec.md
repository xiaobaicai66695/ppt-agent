## ADDED Requirements

### Requirement: Smart recommendation selects executable visual strategy
The system SHALL derive a preset template, valid theme, background-use decision, optional valid background, and explanation from the user's topic without an additional LLM call.

#### Scenario: Topic matches a specialized preset
- **WHEN** a prompt contains terms strongly associated with an available specialized preset
- **THEN** the recommendation selects that preset, its valid default theme, and a background policy derived from the topic and resource catalog

#### Scenario: Topic has no strong match
- **WHEN** no specialized preset receives a meaningful score
- **THEN** the recommendation falls back to the available generic preset and a conservative background policy

### Requirement: Recommendations reference current resources
The system MUST only return template, theme, and background identifiers that exist in the server loaders at recommendation time.

#### Scenario: Preset theme or background is unavailable
- **WHEN** a candidate theme or background identifier is absent from the current resource catalog
- **THEN** the system uses a validated fallback theme or disables the background instead of returning the missing identifier

### Requirement: Recommended strategy is applied server-side
The task service SHALL construct the template scaffold from the recommended preset and apply its theme and background policy before the main Agent run.

#### Scenario: Recommended task is created
- **WHEN** a client creates a task with `template_selection.mode` set to `recommended`
- **THEN** the server resolves and applies the recommendation without requiring the client to copy preset slides or wait for a separate template-fill model call

#### Scenario: User explicitly selects a preset
- **WHEN** a client creates a task with `template_selection.mode` set to `preset` and a valid template name
- **THEN** the server uses that preset and does not override it with an automatic recommendation
