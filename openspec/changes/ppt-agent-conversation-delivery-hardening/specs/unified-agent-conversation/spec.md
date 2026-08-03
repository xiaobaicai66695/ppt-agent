## ADDED Requirements

### Requirement: Unified Conversation Composer
The Dashboard SHALL use one composer for new task creation, feedback queued during generation, and continuation of a completed task.

#### Scenario: No task is selected
- **WHEN** the user submits text with no selected task
- **THEN** the composer creates a new PPT task

#### Scenario: A task is selected
- **WHEN** the user submits text while a task is selected
- **THEN** the message is appended to that task conversation
- **AND** it is accepted immediately or queued according to task status

### Requirement: Template Selection for the Next Task
The unified composer SHALL allow the user to select no template, a preset template, or a custom outline draft for the next newly created task.

#### Scenario: Custom outline returns from Compose
- **WHEN** a user saves a customized outline in the template editor
- **THEN** the Dashboard composer shows that outline as the selected template
- **AND** submitting the next new task sends the validated outline to the task creation API

#### Scenario: Continuing an existing task
- **WHEN** the selected task already exists
- **THEN** template selection is not applied to the continuation message

### Requirement: Recoverable Structured Conversation
The backend SHALL return ordered structured user and assistant messages for an active or persisted task, and the frontend SHALL not clear visible history while loading it.

#### Scenario: Completed task is reopened
- **WHEN** the user reopens a completed task after a server restart
- **THEN** prior user and assistant turns are restored in chronological order

#### Scenario: Legacy task has no structured assistant messages
- **WHEN** an existing task only has `full_answer` or `conversation_content`
- **THEN** the frontend presents the available content as a readable legacy assistant history instead of an empty conversation

### Requirement: Markdown-Preserving Streaming
The system SHALL preserve model-provided Markdown and line boundaries while combining SSE answer chunks into one assistant turn, and SHALL render the result with escaped HTML.

#### Scenario: Markdown answer arrives in arbitrary chunks
- **WHEN** headings, list markers, or code spans are split across multiple SSE answer chunks
- **THEN** the chunks are concatenated in arrival order without punctuation-based sentence splitting
- **AND** the completed turn renders as Markdown

#### Scenario: Raw HTML is included
- **WHEN** an assistant message contains HTML markup
- **THEN** the markup is escaped and is not executed

### Requirement: Accepted Continuation Contract
The continuation POST endpoint SHALL acknowledge accepted or queued work promptly, and continuation events SHALL be delivered through the existing task SSE stream.

#### Scenario: Completed task continuation is submitted
- **WHEN** a valid continuation message is posted for a completed task
- **THEN** the endpoint returns an accepted JSON response without waiting for generation
- **AND** answer and completion events are observable on the task stream

