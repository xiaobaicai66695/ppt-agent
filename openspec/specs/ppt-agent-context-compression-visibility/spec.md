# ppt-agent-context-compression-visibility Specification

## Purpose
TBD - created by archiving change chunked-deck-planning-and-context-compression. Update Purpose after archive.
## Requirements
### Requirement: Context compression triggers before planning calls exceed the safe window
The system SHALL run context compression before model calls when message count or estimated token usage exceeds the configured safe threshold for planning models.

#### Scenario: Context exceeds safe threshold
- **WHEN** a planning or fixer model call is about to use context above the safe threshold
- **THEN** the system compresses older history before calling the primary model
- **AND** the compressed context preserves the current user request with explicit priority

#### Scenario: Previous compression exists
- **WHEN** a later model call needs compression after an earlier compression handoff
- **THEN** the system includes the prior handoff in the new compression input
- **AND** the resulting handoff keeps still-active user requirements and decisions

### Requirement: Users can see context compression progress
The system SHALL emit user-visible progress events when context compression starts, succeeds or falls back.

#### Scenario: Compression starts
- **WHEN** runtime context compression begins
- **THEN** the SSE stream emits a `progress` event with phase `compressing_context`
- **AND** the phase detail explains that earlier conversation is being compacted while the latest request is preserved

#### Scenario: Compression completes
- **WHEN** runtime context compression succeeds
- **THEN** runtime metadata records before/after message and token estimates
- **AND** the frontend can display that the conversation was compressed without showing internal JSON

#### Scenario: Compression falls back
- **WHEN** summarizer compression fails
- **THEN** the system emits or records a warning status
- **AND** generation continues using deterministic anchors or the safest available context

