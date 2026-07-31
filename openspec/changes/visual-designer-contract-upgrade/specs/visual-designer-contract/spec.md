## ADDED Requirements

### Requirement: Template Contract Metadata
The Visual Designer templates SHALL expose machine-readable contract metadata for capacity, required fields, best-fit usage, avoid-fit usage, and overflow strategy.

#### Scenario: Agent reads template capacity
- **WHEN** a template JSON is loaded
- **THEN** the template may declare `contract.capacity`, `contract.required_fields`, `contract.best_for`, `contract.avoid_for`, and `contract.overflow_strategy`
- **AND** agents can use the metadata before generating slide content

### Requirement: Background Policy
The Visual Designer contract SHALL distinguish visual-background pages from information-dense pages.

#### Scenario: Information page avoids busy background
- **WHEN** a template is information-dense
- **THEN** its contract marks background usage as clean or optional
- **AND** title, section, and hero templates can mark image backgrounds as recommended

### Requirement: Local Visual Primitives
The Visual Designer contract SHALL prefer local icon, shape, and chart primitives before external image search.

#### Scenario: Agent chooses local primitive
- **WHEN** a slide needs a visual element
- **THEN** the skill guidance recommends local icons, shapes, charts, or layout primitives first
- **AND** external image search is not required for the first implementation
