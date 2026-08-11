## MODIFIED Requirements

### Requirement: Agent Status Injection
The system SHALL inject a compact code-maintained progress block into agent model calls containing only the total slide count and generated slide count required by the current reasoning loop.

#### Scenario: Agent calls include runtime progress
- **WHEN** a DeepAgent, SlideExecutor, Reviewer, or Fixer model call is executed with runtime metadata configured
- **THEN** the outgoing message list includes generated and total slide counts
- **AND** tool histories, missing-file lists, token statistics, warnings, and other diagnostic metadata remain outside the model status block
- **AND** model calls still work when runtime metadata is absent

### Requirement: Structured Compression Handoff
The system SHALL create a structured compression handoff anchored to the original user request and subsequent explicit user constraints, while recording a bounded before/after summary as runtime observability metadata.

#### Scenario: Compression preserves user intent
- **WHEN** context compression runs during a task
- **THEN** the handoff includes `user_intent_summary`, `preserved_requirements`, `progress_summary`, and `conversation_summary`
- **AND** the original user goal and later explicit constraints take precedence over incidental tool chatter

#### Scenario: Compression records observable difference
- **WHEN** context compression completes
- **THEN** runtime metadata records before and after message counts, before and after token counts, removed message count, saved tokens, saved percentage, and a bounded summary of preserved user requirements

#### Scenario: Summary model fails
- **WHEN** the compression summary model does not return a valid structured result
- **THEN** the system creates a bounded deterministic user-intent handoff and continues the task

### Requirement: Manifest Outcome Validation
The system SHALL validate generated task manifests against code-maintained slide completion metadata, and manifest validation events exposed to the reasoning loop SHALL contain only generated and total slide counts.

#### Scenario: Completion checks generated slides
- **WHEN** a task reaches completion polling or final result emission
- **THEN** code-maintained metadata compares generated slide count with total slide count
- **AND** the task result does not rely on natural-language agent output or an additional LLM disk check to infer completeness

#### Scenario: Manifest validation event is emitted
- **WHEN** manifest validation produces a runtime event
- **THEN** the event metadata contains only `done` and `total`

## ADDED Requirements

### Requirement: Persistent runtime tool event replay
The system SHALL persist structured start/end/error events for every tool call observed by the Agent callback path, independent of the bounded in-memory recent-event window.

#### Scenario: Agent invokes different tool types
- **WHEN** the Agent invokes file, search, manifest, task, shell, Python, conversion, or other registered tools
- **THEN** each observed tool call is persisted with its tool name, status, timestamp, event id, and bounded input/result metadata

#### Scenario: Recent event window evicts early calls
- **WHEN** a task produces more runtime events than the in-memory recent-event limit
- **THEN** the persisted conversation history can still replay earlier tool events
