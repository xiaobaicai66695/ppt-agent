## MODIFIED Requirements

### Requirement: Manifest Outcome Validation
The system SHALL maintain generated-page delivery state as code-owned metadata and use that metadata as the authoritative terminal completion signal. Backend reconciliation MAY observe output artifacts to update metadata, but the LLM SHALL NOT inspect the filesystem or natural-language output to decide whether generation is complete.

#### Scenario: Metadata completes before the Agent returns
- **WHEN** the delivery metadata reaches a nonzero delivered count equal to the planned slide count
- **THEN** the task manager SHALL stop the remaining main Agent control loop
- **AND** the task SHALL complete from the metadata snapshot without a second LLM or terminal disk-validation step

#### Scenario: Metadata remains incomplete
- **WHEN** the main Agent ends while delivery metadata reports pending or failed pages
- **THEN** the task SHALL NOT be marked completed
- **AND** the terminal error SHALL report delivered and planned counts from metadata

#### Scenario: Metadata is empty
- **WHEN** no planned slides exist in delivery metadata
- **THEN** the task SHALL NOT treat zero delivered out of zero planned as successful delivery
