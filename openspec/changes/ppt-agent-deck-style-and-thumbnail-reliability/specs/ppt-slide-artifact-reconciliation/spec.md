## ADDED Requirements

### Requirement: Unique page artifact is normalized
The system SHALL normalize a uniquely identifiable page PPTX to the manifest `output_file` before marking the page complete.

#### Scenario: Generated filename differs by whitespace
- **WHEN** the exact manifest file is absent and exactly one PPTX exists with the same page prefix
- **THEN** the system atomically renames that PPTX to the manifest filename and marks the page complete

#### Scenario: Multiple candidates exist
- **WHEN** more than one PPTX exists with the same page prefix and the exact manifest file is absent
- **THEN** the system does not rename any candidate and keeps the page unresolved

### Requirement: Existing thumbnail follows normalized artifact
The system SHALL keep an existing rendered JPG aligned with a normalized PPTX filename.

#### Scenario: JPG exists for drifted PPTX
- **WHEN** a unique drifted PPTX and its same-stem JPG are normalized
- **THEN** the JPG is renamed to the manifest stem so the thumbnail endpoint can serve it

### Requirement: Legacy ready files remain previewable
The frontend SHALL use an actual ready artifact filename when it can unambiguously associate that file with a manifest page.

#### Scenario: Legacy task has a drifted ready filename
- **WHEN** task metadata declares one filename but ready files contain one same-page filename
- **THEN** the slide card requests preview and download using the actual ready filename
