## ADDED Requirements

### Requirement: Runtime wrappers survive provider abstraction
The system SHALL preserve runtime status injection, context compression, stream read fallback, and streaming tool-call delta sanitization after model creation is moved behind provider adapters.

#### Scenario: Wrapped provider model streams
- **WHEN** a provider-created model is used by PPTPlanner, Reviewer, Fixer, or Compressor
- **THEN** the model is still wrapped by existing runtime status, compression, fallback, and stream sanitation behavior
- **AND** observable runtime traces remain available without provider-specific code in the Agent constructors
