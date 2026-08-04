## 1. Idempotent conversation recovery

- [x] 1.1 Add the latest SSE event boundary to conversation responses and cover it with backend tests
- [x] 1.2 Cache each task's conversation messages, streaming answer, and last applied event ID in Dashboard
- [x] 1.3 Resume running tasks from their cached cursor without clearing restored UI state
- [x] 1.4 Add frontend tests for cursor selection and duplicate-safe conversation merging

## 2. Fast and reliable manifest updates

- [x] 2.1 Implement a DeepAgent manifest tool for structured batch page updates and atomic validation
- [x] 2.2 Register the manifest tool and remove generic `edit_file` from the main Agent tool set
- [x] 2.3 Rewrite the master prompt to use one batch planning update and backend file reconciliation instead of per-page status writes
- [x] 2.4 Add manifest tool tests for valid batch updates and rollback on invalid input

## 3. Live user status

- [x] 3.1 Derive a stable user-facing activity state from phase, tool, progress, retry, connection, and terminal events
- [x] 3.2 Render the live activity line in the unified conversation composer with accessible running and terminal states
- [x] 3.3 Keep raw tool arguments and repetitive execution narration out of the conversation presentation

## 4. Verification and records

- [x] 4.1 Run focused Go tests and backend build
- [x] 4.2 Run frontend unit tests and production build
- [x] 4.3 Validate the OpenSpec change and update iteration/TODO records with results
