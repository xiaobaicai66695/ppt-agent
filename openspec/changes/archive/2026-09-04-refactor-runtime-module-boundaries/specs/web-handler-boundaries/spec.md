## ADDED Requirements

### Requirement: Web handlers SHALL be organized by business responsibility
The Web package SHALL keep authentication, task, conversation, delivery, credential, and administration handlers in independently readable units while preserving the existing route paths, methods, ownership checks, status codes, and response shapes.

#### Scenario: Existing route contract is preserved after file moves
- **WHEN** the server is built and its route table is registered
- **THEN** every existing API path and HTTP method resolves to the same behavior and response contract as before the refactor

### Requirement: HTTP handlers SHALL delegate runtime work
HTTP handlers SHALL be limited to request decoding, authorization/context extraction, service invocation, and response mapping; workflow execution, task state mutation, and model construction SHALL remain outside route functions.

#### Scenario: Task creation remains asynchronous
- **WHEN** a client submits a valid task creation request
- **THEN** the handler returns the existing task response and delegates execution without embedding the workflow implementation in the route function

### Requirement: Web boundary helpers SHALL have one owner
Path safety, error mapping, task ownership, and response projection helpers SHALL have a single package owner and SHALL NOT be duplicated across handler groups.

#### Scenario: Unauthorized resource access is consistent
- **WHEN** an authenticated user requests another user's task, file, conversation, or feedback
- **THEN** the corresponding handler group returns the same authorization result through the shared ownership boundary
