## ADDED Requirements

### Requirement: Content items accept common structured variants
The system MUST accept `content_plan.elements[].items` containing strings, scalar values, structured objects, or a mixture of these forms without failing the entire outline request.

#### Scenario: Model returns object items
- **WHEN** an outline response contains an item such as `{ "title": "增长", "description": "同比提升 20%" }`
- **THEN** the backend normalizes it into one meaningful string item and continues processing the outline

#### Scenario: Model returns mixed scalar items
- **WHEN** an items array contains strings, numbers, booleans, or null values
- **THEN** the backend preserves meaningful values as strings, ignores null values, and returns a valid content plan

### Requirement: Normalized content plan remains stable for consumers
The system SHALL expose normalized `items` as a string array to task manifests, API responses, prompts, and generators.

#### Scenario: Normalized plan is serialized
- **WHEN** a compatible structured content plan has been decoded and serialized again
- **THEN** every emitted `items` entry is a string and no downstream consumer needs dynamic type handling

### Requirement: Invalid outer shapes remain actionable
The system MUST reject irrecoverable content plan shapes with a focused error instead of hanging or silently discarding the whole request.

#### Scenario: Items uses an unsupported outer type
- **WHEN** `items` is neither an array nor a supported scalar/object representation
- **THEN** the API returns a bounded parse error identifying the content plan field
