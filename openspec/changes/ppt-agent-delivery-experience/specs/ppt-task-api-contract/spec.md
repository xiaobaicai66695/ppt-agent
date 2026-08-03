## ADDED Requirements

### Requirement: Task resource ownership
The backend SHALL authorize every task-specific API operation against the authenticated user before returning or mutating task data.

#### Scenario: Task owner accesses a task resource
- **WHEN** an authenticated user requests their own task detail, stream, file, thumbnail, cancellation, deletion, continuation, or conversation
- **THEN** the backend processes the request normally

#### Scenario: Another user requests a task resource
- **WHEN** an authenticated non-admin user requests a task owned by another user
- **THEN** the backend returns a not-found response without exposing task metadata or files

#### Scenario: Administrator accesses a task resource
- **WHEN** an authenticated administrator requests a task resource for support or diagnosis
- **THEN** the backend permits access

### Requirement: Consistent authenticated user context
Authenticated handlers SHALL read user identity from the request authentication context established by middleware.

#### Scenario: Feedback and recommendation endpoints use authenticated identity
- **WHEN** an authenticated user submits feedback or requests insights or recommendations
- **THEN** the handler resolves the same user identity used by task and profile endpoints
- **AND** it does not fail because an unrelated Gin key is absent

### Requirement: Profile summary contract
The backend SHALL expose the registered profile-summary endpoint expected by the frontend and SHALL accept correctly typed profile updates.

#### Scenario: User opens preference settings
- **WHEN** the frontend requests `/api/users/me/profile/summarize`
- **THEN** the backend returns the stored preference summary, task count, and update time

#### Scenario: User saves list preferences
- **WHEN** the user edits comma-separated preference values and saves them
- **THEN** the frontend sends JSON arrays to list-valued fields
- **AND** it only reports success after a successful HTTP response

### Requirement: Frontend API error handling
The frontend API client SHALL reject non-success HTTP responses with a useful error instead of treating error JSON as successful domain data.

#### Scenario: Task mutation fails
- **WHEN** create, cancel, delete, continue, profile, template, or AI endpoint returns a non-2xx response
- **THEN** the calling UI receives a rejected promise with the backend error message when available

### Requirement: API contract tests
The project SHALL include focused tests for task ownership and user-context behavior.

#### Scenario: Authorization regression is introduced
- **WHEN** backend package tests run
- **THEN** tests cover owner, non-owner, administrator, and authenticated user-context cases
