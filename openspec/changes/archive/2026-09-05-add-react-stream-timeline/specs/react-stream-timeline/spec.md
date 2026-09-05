## ADDED Requirements

### Requirement: Task streams expose ordered ReAct timeline events

The task stream SHALL emit safe public `thought`, `tool_call`, `tool_result`, `final_answer`, and `error` events using the existing SSE endpoint. Every emitted event SHALL receive a monotonically increasing task-local ID before it is made visible to listeners or replay clients.

#### Scenario: Multi-step tool workflow

- **WHEN** a task produces a safe thought summary, invokes a tool, receives its result, invokes another tool, and produces an answer
- **THEN** the stream exposes those events in the same chronological order without grouping all tools after the answer

#### Scenario: Reconnect after a partial turn

- **WHEN** a client reconnects with the last received event ID during an active turn
- **THEN** the server replays only later events in their original order and the client does not render a duplicate timeline entry

### Requirement: Tool calls and results remain separate correlated backend events

The backend SHALL emit `tool_call` immediately when a visible invocation begins and SHALL later emit a separate `tool_result` correlated by `tool_call_id`. It MUST NOT delay the call until completion, rewrite the result as a call, or relocate either event in replay history.

#### Scenario: Slow tool execution

- **WHEN** a tool takes observable time to complete
- **THEN** the client receives and renders the running `tool_call` before the tool finishes, followed later by its `tool_result`

#### Scenario: Repeated tool name

- **WHEN** two invocations use the same tool name in one turn
- **THEN** each result is associated with the correct call ID and both call/result pairs retain source order

#### Scenario: Missing provider result callback

- **WHEN** the agent run ends with a call that has no provider completion callback
- **THEN** the backend emits exactly one fallback `tool_result` for that still-pending call with an honest success or error status

### Requirement: Thought events expose safe summaries only

The system SHALL use `thought` events for user-visible process narration or deterministic stage summaries and MUST NOT expose private chain-of-thought, model `ReasoningContent`, hidden prompts, credentials, or unredacted tool payloads.

#### Scenario: Model has hidden reasoning metadata

- **WHEN** a provider response includes private reasoning content or a reasoning preview
- **THEN** that content is excluded from the public SSE thought payload while a safe stage summary may still be shown

#### Scenario: Retrieval starts

- **WHEN** a chat request requires web or image retrieval
- **THEN** the stream emits a concise safe thought summary before the corresponding tool call

### Requirement: Final answers stream incrementally and remain durable

The system SHALL emit user-facing answer text as incremental `final_answer` events grouped by a stable segment ID. Only final answer text SHALL be persisted as the durable assistant conversation turn; transient thought and tool trace content SHALL remain outside the durable chat transcript.

#### Scenario: Native model stream

- **WHEN** the chat model yields multiple answer chunks
- **THEN** the client updates one final-answer bubble as each chunk arrives instead of waiting for the full answer

#### Scenario: Conversation reload after completion

- **WHEN** a completed conversation is loaded from its durable snapshot
- **THEN** user and final assistant messages are restored without private thought text or duplicated tool payloads

### Requirement: The workbench renders a flat accessible execution timeline

The workbench SHALL render thoughts, paired tool call/result cards, and final answers in arrival order. A `tool_call` SHALL create one visible tool card immediately; a matching `tool_result` SHALL update that same card's state and result panel instead of creating a detached result row. Thought and tool cards SHALL support keyboard-accessible expansion, and status meaning SHALL be conveyed by text in addition to color or icon.

#### Scenario: Tool result arrives

- **WHEN** a `tool_result` follows a matching `tool_call`
- **THEN** the existing tool card updates to show the returned content and terminal state, with no second tool row and no bottom-piled result card

#### Scenario: Long result content

- **WHEN** a safe tool result exceeds the compact preview limit
- **THEN** the full safe result remains available inside the expanded tool card's bounded scrollable result panel without content truncation

#### Scenario: Manual scroll-up during streaming

- **WHEN** the user scrolls away from the bottom while new events arrive
- **THEN** the timeline preserves the user's scroll position and resumes automatic following only after the user returns near the bottom

#### Scenario: Narrow viewport and reduced motion

- **WHEN** the timeline is viewed at a narrow supported width or with reduced motion enabled
- **THEN** all content and disclosure controls remain reachable, and nonessential entry motion is removed

### Requirement: Legacy task lifecycle remains compatible

The stream endpoint, task identifiers, event cursor, heartbeat, delivery fields, and explicit terminal event semantics SHALL remain unchanged. The frontend SHALL continue accepting legacy `answer`, `system_step`, and merged tool payloads during rolling deployment.

#### Scenario: Legacy answer event

- **WHEN** the frontend receives an `answer` event from an older backend
- **THEN** it appends that content to a final-answer segment without duplicating content received through the new protocol

#### Scenario: Planner answer end

- **WHEN** the stream receives `answer_end` during PPT generation
- **THEN** the client closes the current text segment but keeps the SSE connection open until an explicit terminal lifecycle event arrives
