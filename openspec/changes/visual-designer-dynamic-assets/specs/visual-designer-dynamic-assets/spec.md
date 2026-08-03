## ADDED Requirements

### Requirement: Local Visual Asset Library
The Visual Designer skill SHALL provide a local asset library for icons, backgrounds, and patterns.

#### Scenario: Generator selects local icon
- **WHEN** a supported generator receives an icon id or content tag
- **THEN** it can resolve a local asset from `assets/manifest.json`
- **AND** it can place the asset without external image search

### Requirement: Density-Aware Layout
The Visual Designer generators SHALL use content density to adjust sizing and composition for sparse and normal content.

#### Scenario: Sparse content uses stronger focal composition
- **WHEN** a slide has one to three short content items
- **THEN** the generator increases focal text or icon size within safe bounds
- **AND** avoids leaving the slide visually hollow

### Requirement: Layout Intent And Alignment
The Visual Designer generators SHALL distinguish centered focus layouts from scan-oriented information layouts.

#### Scenario: Quote or sparse content is centered
- **WHEN** a supported template is best read as a focal statement
- **THEN** text can be centered and visually anchored
- **AND** information-dense layouts still default to left scan alignment

### Requirement: Offline Default Assets
The local asset system SHALL not require external image search or online generation for default PPT generation.

#### Scenario: Offline generation
- **WHEN** the generator runs without network access
- **THEN** default assets and backgrounds still resolve from the local skill directory
