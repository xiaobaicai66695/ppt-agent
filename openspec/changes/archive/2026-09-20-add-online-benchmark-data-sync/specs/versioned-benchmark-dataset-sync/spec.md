## ADDED Requirements

### Requirement: Approved evidence becomes explicit benchmark cases
The system SHALL require an administrator to create or approve an explicit benchmark case from captured evidence before it is eligible for a published dataset. The case SHALL contain a suite, canonical input, expected/rubric fields, a stable split group, and a revision identifier.

#### Scenario: Candidate lacks benchmark expectations
- **WHEN** a candidate has captured stage artifacts but no approved case revision
- **THEN** it is not eligible for dataset publication
- **AND** user feedback alone does not satisfy the expected-result requirement

#### Scenario: Candidate case is approved
- **WHEN** an administrator approves a redacted case revision with required benchmark fields
- **THEN** the candidate becomes eligible for a compatible Planner, Reviewer, or Fixer dataset
- **AND** its source stage provenance remains auditable

### Requirement: Dataset versions isolate development and holdout evidence
The system SHALL publish immutable named dataset versions with roles `core`, `active-dev`, or `validation`. A case or split group previously exposed through an `active-dev` or `core` dataset SHALL NOT be published into a validation dataset.

#### Scenario: Validation is frozen
- **WHEN** an administrator freezes a validation dataset version
- **THEN** its membership and pinned case revisions cannot be changed
- **AND** its bundle can be exported for evaluation without exposing it through the active development channel

#### Scenario: Prior validation is promoted
- **WHEN** a completed validation dataset is promoted after its evaluation window
- **THEN** the system publishes a new `active-dev` version from its approved membership
- **AND** a later validation dataset excludes the promoted case split groups

### Requirement: Dataset bundles are verified before local use
The system SHALL export published development or validation datasets as manifests with file checksums and SHALL require local import to verify those checksums before resolving cases for `pptbench`.

#### Scenario: Local pull succeeds
- **WHEN** an authorized developer pulls a published dataset bundle and all checksums match
- **THEN** the bundle is installed under the local private benchmark import root
- **AND** `pptbench` can select the named imported dataset for Planner, Reviewer, or Fixer

#### Scenario: Bundle integrity fails
- **WHEN** a downloaded bundle is missing a declared file or has a checksum mismatch
- **THEN** local import fails without activating the dataset
- **AND** existing local datasets remain usable
