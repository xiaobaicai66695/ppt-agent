## 1. Task conversation foundation

- [x] 1.1 Add persistent conversation/generation phase support to task metadata and ensure non-generating tasks survive restart recovery.
- [x] 1.2 Add TaskManager operations to create a lightweight user-owned conversation task, append/recover messages, and start generation for that same task ID.
- [x] 1.3 Add focused task and database tests for conversation task lifecycle, ownership, deletion, and restart behavior.

## 2. Context-aware message API

- [x] 2.1 Extend `/api/messages` to create or validate a task ID before routing, persist turns, return the task ID, and build bounded task context.
- [x] 2.2 Make route classification and generation preparation consume task context; preserve earlier subject when a follow-up delegates topic/style decisions.
- [x] 2.3 Add an owned-task generation-start endpoint/path that reuses the conversation task rather than allocating a second task.
- [x] 2.4 Add web API tests for first-message creation, follow-up context, cross-user rejection, no-render chat, and same-ID generation handoff.

## 3. Workbench integration

- [x] 3.1 Update API types and Dashboard/Home flow to retain the returned task ID and use it for all follow-ups.
- [x] 3.2 Update task list/detail UI to display conversation and generation phases without treating conversation tasks as failed renders.
- [x] 3.3 Add focused frontend tests for task-ID reuse and generation handoff, then build the frontend.

## 4. Evaluation, delivery, and records

- [x] 4.1 Add a router benchmark regression covering subject description → delegated topic/style → same-task generation preparation.
- [x] 4.2 Run focused backend/frontend/benchmark validation and OpenSpec strict validation.
- [x] 4.3 Build Linux artifacts, deploy to `remote-dev`, run low-cost cross-turn smoke tests, clean temporary records/files, and record deployment evidence.
