## Context

PPT Agent currently persists one task-level lifecycle (`conversation`, `running`, `completed`, `failed`, `cancelled`) and exposes live runtime events. It can route a post-delivery request to a scoped Fixer, but the change writes into the task workspace directly and the user has no durable pre-execution approval or candidate-result decision point. Existing retry, model fallback and recovery code is useful but has no single user-facing recovery state.

The change deliberately excludes the retired runtime QA flow and page-level lifecycle. Benchmark remains a development/release gate and is not part of a user's generation task state machine.

## Goals / Non-Goals

**Goals:**

- Let each user select `auto`, `sensitive`, or `manual` approval for a task, with a persisted default and task-local override.
- Pause a task durably at `waiting_approval` only when its approval policy requires it, with a server-generated explanation of the interruption.
- Create immutable deck revisions so a modification is previewed as a candidate, explicitly accepted or rejected, and can be rolled back without overwriting history.
- Expose automatic retries and fallback as a bounded `recovering` phase with auditable error and recovery events.

**Non-Goals:**

- No `queued` task status, shared-model queue, page-level state machine, or runtime QA/review loop.
- No user approval for every low-impact render operation in the default mode.
- No use of untrusted LLM prose as the authority for approval reasons or scope.
- No new external storage or queue dependency in the first release.

## Decisions

### 1. Separate task lifecycle from revision lifecycle

Task status answers whether the task is executing or waiting for a user. Revision status answers whether a specific output is current or a pending candidate. Task states are `conversation`, `running`, `waiting_approval`, `waiting_confirmation`, `recovering`, `completed`, `failed`, and `cancelled`. Revision states are `active`, `candidate`, `confirmed`, `rejected`, `superseded`, and `discarded`.

`retrying` is recorded as a recovery action while the task is `recovering`; `error` is an event with a failure class, not an additional terminal state. This avoids invalid combinations such as a completed task also being retrying.

The alternative of extending per-page statuses was rejected because the workflow has no page QA lifecycle and pages are only the scope and diff unit of a revision.

### 2. Approval policy controls whether execution pauses

Each task stores an effective `approval_mode`: `auto` (default), `sensitive`, or `manual`. The account setting supplies a default; a new task may override it. `auto` never creates an approval pause. `sensitive` pauses only on server-classified high-impact actions: scope expansion, global visual/outline change, external asset use, overwrite of an active revision, or recovery degradation. `manual` pauses before each change execution action.

The backend creates a durable approval request with a reason code, action summary, affected pages, consequence of rejection and allowed decisions. It derives these fields from trusted route/scope/version data. The frontend is a client of this record, not the source of truth.

### 3. Candidate revisions are copy-on-write workspace snapshots

An initial successful deck becomes revision 1 and is `active`. A continuing modification creates a candidate revision with a parent revision ID, immutable manifest snapshot, affected pages, diff summary and candidate artifact paths. It does not replace the active revision. Acceptance publishes the candidate and marks the prior revision `superseded`; rejection marks only the candidate rejected. Rollback creates a new revision from a prior snapshot and marks it active.

The first delivery stores full manifest snapshots and copies changed artifacts when needed. This favors recoverability and simple rollback over deduplication; later storage optimization can reuse unchanged artifacts by content hash.

### 4. Approval and confirmation are distinct checkpoints

An approval happens before an operation and asks whether the Agent may execute it. A confirmation happens after a candidate revision is rendered and asks whether the user accepts its result. Rejection of approval prevents the operation; rejection of confirmation retains the active version. Both actions are authenticated, idempotent and persisted before workers resume.

### 5. Existing recovery paths are normalized through runtime events

Existing model fallback, recoverable asset errors, Planner draft recovery and scoped Fixer/code fallback emit a common `recovery_started`/`recovery_finished` event with failure kind and action. A recoverable error sets task status to `recovering`, then returns to `running` when successful. Exhaustion records the final error and sets `failed`; actions that require a business decision create an approval request instead.

## Risks / Trade-offs

- [A task is paused while a worker is active] → Persist approval/checkpoint records before exposing the state and resume only through a server-side decision endpoint.
- [Candidate artifact copies consume disk] → Start with scoped copies and add retention/cleanup after correctness is verified.
- [Users encounter excessive approval prompts] → Default to `auto`, use an explicit impact classifier, and offer task-level overrides.
- [A restart loses in-memory state] → Read task, approval and revision records from persistent storage; pending states remain visible after reconnect.
- [A change candidate corrupts the active deck] → Never write to the active revision directory; publish only after confirmation succeeds.

## Migration Plan

1. Add backwards-compatible database fields/tables and default old tasks to `auto` with no revision records.
2. Add task status handling, approval APIs and Dashboard approval UI while keeping existing generation flows automatic by default.
3. Route continuation/Fixer output through candidate revision creation, then expose accept/reject and history rollback.
4. Normalize recovery events, add focused backend/frontend tests and build artifacts.
5. Deploy with compatibility reads for existing tasks. Roll back by disabling approval routing; existing active artifacts and task records remain readable.

## Open Questions

- Initial release uses all-or-nothing candidate acceptance. Page-subset acceptance will reuse the same revision model in a later iteration.
