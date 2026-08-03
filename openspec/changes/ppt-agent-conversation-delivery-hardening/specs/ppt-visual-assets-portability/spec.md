## ADDED Requirements

### Requirement: Self-Contained Visual Asset Resolution
The Visual Designer SHALL resolve all bundled icons, backgrounds, and patterns relative to the deployed `ppt-agent/skills/visual_designer` directory without depending on sibling workspace directories.

#### Scenario: Only ppt-agent is deployed
- **WHEN** the application runs from a deployment containing only the `ppt-agent` directory and its runtime dependencies
- **THEN** every asset declared by `assets/manifest.json` resolves to an existing file

### Requirement: Visible Icon Fallback
An icon-based generator SHALL render a visible semantic fallback when a requested asset cannot be resolved.

#### Scenario: Requested icon is missing
- **WHEN** an icon id is absent from the manifest or its file is unavailable
- **THEN** the generated slide contains a themed fallback mark and readable label
- **AND** does not contain an empty placeholder square

### Requirement: Asset Manifest Validation
The project SHALL provide a low-cost validation that verifies manifest paths, asset readability, and icon-grid rendering from a deployment-like working directory.

#### Scenario: Asset validation runs
- **WHEN** the asset validation command is executed
- **THEN** it fails with the missing asset ids when any declared file is unavailable
- **AND** succeeds without network access when the bundle is complete

