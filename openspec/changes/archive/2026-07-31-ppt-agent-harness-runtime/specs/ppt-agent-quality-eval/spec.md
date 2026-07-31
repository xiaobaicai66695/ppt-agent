## ADDED Requirements

### Requirement: PPT Quality Evaluation Fixture Set
The system SHALL provide an offline evaluation fixture file containing representative PPT generation cases with user goals, expected slide counts, layout/content constraints, and visual quality rubrics.

#### Scenario: Evaluation cases are available
- **WHEN** a developer opens the eval fixture file
- **THEN** it contains multiple named cases with input prompts and scoring rubrics
- **AND** each case can be consumed by an offline evaluator without calling an LLM

### Requirement: Offline PPT Artifact Evaluator
The system SHALL provide a script that evaluates a generated work directory against manifest presence, slide artifact presence, QA summaries, and rubric metadata.

#### Scenario: Generated work directory is evaluated
- **WHEN** the evaluator is run with a work directory and case file
- **THEN** it emits structured JSON containing per-case checks, score components, and pass/fail status
- **AND** it exits successfully for well-formed local inputs without requiring network access

#### Scenario: Missing artifacts are reported
- **WHEN** expected manifest or slide output files are missing
- **THEN** the evaluator records explicit missing-file findings
- **AND** the output can be used as regression evidence in future tests
