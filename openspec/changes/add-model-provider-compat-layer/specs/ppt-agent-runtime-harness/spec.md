## ADDED Requirements

### Requirement: Runtime wrappers survive provider abstraction
The system SHALL preserve runtime status injection, context compression, stream read fallback, and streaming tool-call delta sanitization after model creation is moved behind provider adapters.

#### Scenario: Wrapped provider model streams
- **WHEN** a provider-created model is used by PPTPlanner, Reviewer, Fixer, or Compressor
- **THEN** the model is still wrapped by existing runtime status, compression, fallback, and stream sanitation behavior
- **AND** observable runtime traces remain available without provider-specific code in the Agent constructors

### Requirement: Conversation timeline preserves observable event order
The system SHALL present user-visible assistant text and observable tool/runtime events in one chronological timeline rather than rendering tool calls and text in separate batches.

#### Scenario: Assistant text appears before and after a tool call
- **WHEN** the agent streams visible assistant text, calls a tool, and then streams more visible assistant text
- **THEN** the workbench conversation renders the first text segment, the tool event card, and the later text segment in that order
- **AND** adjacent tool calls may be grouped only when no assistant/user text occurs between them

#### Scenario: Runtime trace excludes hidden reasoning
- **WHEN** the UI builds the observable execution timeline
- **THEN** it uses SSE answer chunks and sanitized runtime metadata
- **AND** hidden chain-of-thought or provider-private reasoning content is not exposed as assistant chat content
