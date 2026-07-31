# ppt-agent-runtime-harness Specification

## Purpose
TBD - created by archiving change ppt-agent-harness-runtime. Update Purpose after archive.
## Requirements
### Requirement: Runtime Metadata Snapshot
The system SHALL maintain a per-task runtime metadata snapshot that records task identity, working directory, phase, elapsed time, tool call counts, tool error counts, LLM token counts, compression statistics, slide progress, QA issue counts, manifest validation, and budget warnings.

#### Scenario: Metadata is created for a new PPT task
- **WHEN** a PPT generation task is created
- **THEN** the task has a runtime metadata snapshot associated with its task id and work directory
- **AND** the snapshot reports an initial phase and elapsed runtime

#### Scenario: Tool and model activity updates metadata
- **WHEN** an agent invokes tools or model calls during task execution
- **THEN** the metadata records aggregate tool counts, error counts, and token usage
- **AND** the metadata can be read without depending on plain log parsing

### Requirement: Agent Status Injection
The system SHALL inject a compact runtime status block into agent model calls so the current phase, budget signals, retries, QA counts, and slide progress are available to the reasoning loop.

#### Scenario: Agent calls include runtime status
- **WHEN** a DeepAgent, SlideExecutor, Reviewer, or Fixer model call is executed with runtime metadata configured
- **THEN** the outgoing message list includes a generated runtime status block
- **AND** model calls still work when runtime metadata is absent

### Requirement: Runtime Budgets and Warnings
The system SHALL compute observable budget warnings for tool calls, same-argument repetition, elapsed runtime, token usage, and per-slide QA/fix pressure.

#### Scenario: Budget pressure is visible
- **WHEN** runtime counters exceed configured warning thresholds
- **THEN** the runtime metadata exposes budget warning keys
- **AND** those warnings are included in the agent status block and SSE metadata payload

### Requirement: Structured Compression Handoff
The system SHALL include structured runtime metadata in compression handoff payloads so compressed context preserves execution state, budget pressure, and validation evidence.

#### Scenario: Compression preserves runtime state
- **WHEN** context compression runs during a task with runtime metadata
- **THEN** the handoff payload includes a runtime metadata snapshot
- **AND** compression statistics are recorded back into runtime metadata

### Requirement: Manifest Outcome Validation
The system SHALL validate generated task manifests against produced slide artifacts and expose missing files, failed pages, and completion counts as structured runtime metadata.

#### Scenario: Completion checks generated files
- **WHEN** a task reaches completion polling or final result emission
- **THEN** the runtime metadata includes manifest validation counts and missing-file identifiers
- **AND** the task result does not rely only on natural-language agent output to infer file completeness

