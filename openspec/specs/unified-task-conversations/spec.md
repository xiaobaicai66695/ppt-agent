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
