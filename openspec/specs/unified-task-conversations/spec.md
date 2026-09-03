# Unified Task Conversations

## Purpose

Keep every workbench interaction in one durable, user-owned task from the first message through PPT delivery.

## Requirements

### Requirement: First workbench message creates a task conversation
The system SHALL create a user-owned task conversation before routing the first non-empty workbench message when no `task_id` is supplied. The response SHALL include that task ID and the task SHALL be visible in the task list without starting PPT planning or rendering.

#### Scenario: First general message
- **WHEN** an authenticated user sends a non-empty workbench message without a task ID
- **THEN** the system creates one `conversation` task, persists the user message, routes it, and returns the same task ID with the route result

#### Scenario: First explicit PPT request
- **WHEN** an authenticated user sends an explicit PPT request without a task ID
- **THEN** the system creates one task conversation and returns its task ID without requiring the frontend to create a second task

### Requirement: Subsequent messages use task-scoped context
The system SHALL require that subsequent workbench messages target an owned task ID and SHALL provide a bounded chronological history of that task to intent routing and generation preparation.

#### Scenario: Follow-up delegates unspecified details
- **WHEN** a task contains a prior message describing a 青甘大环线旅游项目介绍 and the user later says “你决定主题和风格吧”
- **THEN** routing and generation preparation retain the earlier project subject and treat the later message as a constraint delegation rather than a new unrelated topic

#### Scenario: Foreign task access
- **WHEN** a user sends a message with a task ID owned by another user
- **THEN** the system rejects the request and does not disclose or append that task's context

### Requirement: Intent changes task phase without replacing its identity
The system SHALL use the same task ID as a workbench conversation moves between chat, clarification, planning, generation, completion, and failure. Only an explicit generation action SHALL start the Planner/rendering workflow.

#### Scenario: Clarification followed by creation
- **WHEN** a task is in conversation or planning phase and the user subsequently requests PPT generation
- **THEN** the system starts generation for the existing task ID using its persisted context

#### Scenario: Chat does not render
- **WHEN** a routed task message has chat intent
- **THEN** the system persists the exchange and keeps the task in conversation phase without invoking Planner, Reviewer, or rendering

### Requirement: Workbench clients keep the returned task identity
The workbench client SHALL store the task ID returned for the first message, pass it on every follow-up, and transition the same selected task into the appropriate view or generation action.

#### Scenario: Follow-up from the task workspace
- **WHEN** the user sends a second message in an existing workbench conversation
- **THEN** the client sends the selected task ID and displays both turns under that task rather than creating a new list item

#### Scenario: Start generation from a conversation task
- **WHEN** the route result directs the client to start PPT creation
- **THEN** the client starts generation for the returned task ID and does not call the legacy create endpoint to allocate another task

### Requirement: Chat replies integrate material and degrade honestly

The system SHALL answer a routed chat message directly. When verified web or image material is available, it SHALL synthesize the material and render only valid, deduplicated source references; when a requested capability is unavailable, it SHALL state that limitation without claiming a search or preview succeeded.

#### Scenario: User requests material for an existing topic

- **WHEN** a conversation about a travel project is followed by “补充些材料和图片参考” and usable web evidence is available
- **THEN** the chat reply summarizes topic-relevant material before listing valid web sources, and it only renders image references that were actually returned

#### Scenario: Generic route fallback with retrieval evidence

- **WHEN** the intent router supplies a generic chat fallback but usable retrieval evidence exists
- **THEN** the final chat reply uses evidence-backed content rather than returning only the generic fallback

#### Scenario: Invalid source from an augmentation

- **WHEN** an augmentation contains a blank, non-HTTP(S), malformed, or duplicate source URL
- **THEN** the final chat reply omits that source link while retaining any other valid source references
