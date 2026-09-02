## 1. Task approval model and durable state

- [x] 1.1 Add approval modes, waiting/recovering task states, trusted approval reason codes and backward-compatible task metadata.
- [x] 1.2 Persist pending approval requests and provide idempotent owner-only approve, adjust-scope and reject transitions.
- [ ] 1.3 Emit runtime/SSE status events for approval pauses, decisions and normalized recovery actions; add focused state transition tests.

## 2. Candidate deck revisions

- [ ] 2.1 Add immutable revision records and workspace snapshot helpers for active and candidate deck artifacts.
- [ ] 2.2 Route completed-task modifications through candidate creation without overwriting the active revision.
- [ ] 2.3 Add owner-only candidate accept, reject and rollback transitions with artifact/manifest integrity tests.

## 3. Workbench controls

- [ ] 3.1 Add account-default and task-level approval mode controls to the creation flow and API types.
- [x] 3.2 Render an actionable approval interruption panel with reason, scope, consequence and decision controls.
- [ ] 3.3 Render candidate revision summary/history controls and distinguish approval, confirmation, recovery and terminal task states.

## 4. Verification and delivery

- [ ] 4.1 Add backend tests for auto/sensitive/manual policies, pending approval recovery, authorization and revision transitions.
- [ ] 4.2 Run focused backend tests, `go build ./...`, frontend build and OpenSpec strict validation.
- [ ] 4.3 Build Linux deliverables, deploy, restart, run the smallest approval/revision smoke test, clean test data and record evidence.
