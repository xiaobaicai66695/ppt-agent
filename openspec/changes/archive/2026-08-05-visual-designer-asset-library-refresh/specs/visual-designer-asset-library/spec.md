## ADDED Requirements

### Requirement: Offline asset manifest
The Visual Designer SHALL load icons, editorial backgrounds, content photos, and patterns from a versioned local manifest, and every externally sourced asset SHALL declare its source, license, attribution, download URL, and normalized dimensions.

#### Scenario: External asset is available offline
- **WHEN** a generator resolves an asset id registered in the manifest
- **THEN** the system SHALL return an existing local file without making a network request

#### Scenario: External metadata is incomplete
- **WHEN** asset validation encounters an externally sourced entry without source, license, attribution, download URL, or dimensions
- **THEN** validation SHALL report a deterministic error for that asset

### Requirement: Semantic icon selection
The Visual Designer SHALL select icons from manifest keywords and aliases using deterministic ranking, and SHALL not substitute an unrelated engineering icon when no semantic match exists.

#### Scenario: Known business or domain concept
- **WHEN** slide text contains a registered Chinese or English keyword such as policy, meeting, ecology, map, contact, or thanks
- **THEN** the system SHALL resolve the corresponding semantic icon id

#### Scenario: Unknown concept
- **WHEN** slide text contains no registered keyword and the caller provides no explicit fallback
- **THEN** the system SHALL omit the icon instead of rendering `layout`, `primitive`, `review`, or an abbreviation placeholder

### Requirement: Semantic content photo selection
The Visual Designer SHALL register editable content photos as `photo` assets with category tags and multilingual keywords, and SHALL select a deterministic local default photo for an image-text slide when no explicit image is available.

#### Scenario: Explicit image is available
- **WHEN** `generate_image_text` receives an existing `image_path` or registered photo asset id
- **THEN** the system SHALL place that file as a real PowerPoint picture object in the requested image region

#### Scenario: Image-text slide has no explicit image
- **WHEN** `generate_image_text` receives no valid `image_path`
- **THEN** the system SHALL select a semantically related local photo from the slide text, falling back to a fixed general workspace photo
- **AND** the system SHALL NOT draw an icon panel, pattern placeholder, fake image frame, or internal implementation labels in place of the photo

#### Scenario: User replaces the default image
- **WHEN** a user edits the generated image-text slide in PowerPoint
- **THEN** the default visual SHALL be a replaceable picture object rather than a group of decorative shapes

### Requirement: Background theme catalog
The Visual Designer SHALL load background themes, scenarios, recommended palettes, image paths, and provenance from `background_templates/manifest.json`, while preserving existing theme ids and direct image references.

#### Scenario: Theme selection
- **WHEN** a generator requests an existing background theme id
- **THEN** the system SHALL select a local image declared for that theme and expose the theme's recommended palette

#### Scenario: Direct image reference
- **WHEN** a task provides `<theme>/images/<file>` as its background
- **THEN** the system SHALL resolve exactly that local file without replacing it with another image

### Requirement: Background rotation capacity
Each declared background theme SHALL contain at least four valid 16:9 images so adjacent visual pages can rotate backgrounds without immediate repetition.

#### Scenario: Catalog validation
- **WHEN** the background manifest validator scans all declared themes
- **THEN** it SHALL report any theme with fewer than four existing 1920x1080-compatible images

### Requirement: Reproducible asset synchronization
The project SHALL provide a script that can download the curated external sources, normalize files into the repository directory contract, regenerate manifests, and fail before replacing a valid file when a download or image check is invalid.

#### Scenario: Successful synchronization
- **WHEN** the sync script completes with network access
- **THEN** all curated icons, backgrounds, and patterns SHALL exist at stable paths and both manifests SHALL match the downloaded files

#### Scenario: Invalid remote response
- **WHEN** a remote source returns a non-image or unsupported image
- **THEN** the sync script SHALL fail with the affected asset id and SHALL not leave a partial target file

### Requirement: Asset maintenance documentation
The Visual Designer README/SKILL and generator reference SHALL document the v2 directory structure, attribution obligations, semantic fallback behavior, background theme manifest, synchronization command, and validation command.

#### Scenario: Maintainer adds an asset
- **WHEN** a maintainer follows the documented workflow
- **THEN** they SHALL know where to register metadata, how to normalize the file, and how to validate generator compatibility
