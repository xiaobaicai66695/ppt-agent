## ADDED Requirements

### Requirement: LLM-first structured task routing
The system SHALL use a language model as the primary semantic decision maker for initial PPT task intent, domain, complexity, page estimate, Agent type, pipeline, concurrency, confidence, and reasoning, without executing keyword rules first.

#### Scenario: Model returns a valid generation route
- **WHEN** a user creates a PPT task and the routing model returns valid structured output
- **THEN** the task SHALL use that classification and route without a second classification call
- **AND** observable routing status SHALL identify the route as model-decided

#### Scenario: Model routing fails
- **WHEN** the routing model is unavailable or returns invalid structured output
- **THEN** the system SHALL use a deterministic supported generation fallback
- **AND** it SHALL NOT infer semantic intent through keyword matching

### Requirement: Routing selects only executable initial Agents
The system SHALL validate model-selected Agent types and pipeline stages against implementations available to the initial generation runtime.

#### Scenario: Model selects an unsupported Agent
- **WHEN** the model returns an Agent type that has no initial-generation implementation
- **THEN** the router SHALL normalize the decision to the supported DeepAgent route
- **AND** the route SHALL remain `plan -> generate`

### Requirement: Initial generation excludes QA
The initial PPT generation route SHALL skip Reviewer and automatic Fixer stages regardless of legacy QA environment settings.

#### Scenario: Legacy QA flag is enabled
- **WHEN** the service starts with a legacy QA enable flag and a new PPT task is created
- **THEN** Reviewer and automatic Fixer SHALL NOT be registered in the initial Agent
- **AND** observable routing SHALL report QA as disabled

#### Scenario: User requests a post-delivery correction
- **WHEN** a user explicitly asks to modify a delivered page
- **THEN** the existing targeted correction flow MAY invoke Fixer independently of initial generation QA

