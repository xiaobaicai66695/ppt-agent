## ADDED Requirements

### Requirement: Accurate routing and delivery phase status
The task status UI SHALL display the model-selected routing source, the actual initial generation pipeline, and metadata-driven delivery completion without showing removed QA or redundant file-verification phases.

#### Scenario: Initial route is displayed
- **WHEN** a task receives an LLM routing decision
- **THEN** the status log SHALL show the selected Agent and `plan -> generate` pipeline
- **AND** it SHALL show QA as disabled

#### Scenario: All pages are delivered
- **WHEN** runtime delivery metadata reaches N/N pages
- **THEN** the UI SHALL transition from generation to completed delivery
- **AND** it SHALL NOT remain in a model-driven `checking generated files` phase
