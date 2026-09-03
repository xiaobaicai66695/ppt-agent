# Chat Agent Quality Evaluation

## Purpose

Keep workbench chat replies grounded, attributable, and evaluable through both deterministic contracts and explicit real-model runs.

## Requirements

### Requirement: Chat benchmark coverage and holdout isolation

The system SHALL provide at least ten valid, uniquely identified `chat` cases in each `test` and `validation` benchmark dataset. Test and validation chat requests SHALL not reuse an identical normalized request.

#### Scenario: Developer loads chat benchmark data

- **WHEN** the benchmark loader reads the `chat` suite for either dataset
- **THEN** it returns at least ten cases with non-empty unique IDs and reports an error for duplicate IDs or overlapping requests

### Requirement: Chat benchmark executes production reply assembly

The system SHALL evaluate chat cases through the production chat reply assembly with fixture-controlled conversation context, route fallback and retrieval/image augmentation inputs. When explicitly enabled, the benchmark SHALL call the configured production text-model chain for every selected case, persist each response, and fail visibly if model generation fails.

#### Scenario: Retrieval-backed material request

- **WHEN** a chat case provides usable web evidence for a request to supplement a prior topic
- **THEN** its model output includes a topic-relevant synthesized answer and source references rather than only a generic PPT invitation

#### Scenario: Real model is unavailable

- **WHEN** the explicit real-model benchmark is enabled and the configured model provider rejects or fails a generation
- **THEN** the benchmark records the model error and fails instead of treating the production fallback text as a successful model answer

### Requirement: Deterministic chat quality gates

The benchmark SHALL fail a chat case that emits unsafe or malformed source links, presents unavailable image search as successful, or ignores supplied retrieval evidence in favor of a generic fallback.

#### Scenario: Image capability is unavailable

- **WHEN** a chat case requests images and fixture augmentation reports image search is unavailable
- **THEN** the output explicitly states that no image candidates were retrieved and contains no fabricated image Markdown
