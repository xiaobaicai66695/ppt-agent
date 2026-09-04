## 1. Baseline and guardrails

- [x] 1.1 Capture the current route table, exported response contracts, key SSE terminal events, and package test baseline.
- [x] 1.2 Add a short package boundary map and define the file naming/ownership convention used by the refactor.
- [x] 1.3 Add the approved terminology map, comment convention, and old task-file compatibility rules.

## 2. Web handler boundaries

- [x] 2.1 Move shared path, ownership, response, and request parsing helpers into focused Web boundary files without changing behavior.
- [x] 2.2 Split authentication, task, conversation, and delivery handlers from `handler.go`, preserving `*Server` receivers and route registration.
- [x] 2.3 Split credential, runtime-event, log-analysis, and admin handlers; keep `server.go` responsible for service construction and routes only.
- [x] 2.4 Run Web/task focused tests and add a route-contract regression covering methods, paths, status codes, and representative response shapes.

## 3. Task runtime boundaries

- [x] 3.1 Move stable task types and state/lock transitions into focused files, preserving `TaskState` ownership and public method names.
- [x] 3.2 Separate SSE event buffering, listener management, replay cursors, and answer chunk normalization from general task lifecycle code.
- [x] 3.3 Separate persistence/recovery, conversation state, and delivery reconciliation; preserve callback ordering and database projections.
- [x] 3.4 Run task, database, web, and SSE recovery tests; add a concurrent stream/replay regression for duplicate streams and terminal events.

## 4. Model runtime boundaries

- [x] 4.1 Move model option/configuration and provider resolution into focused files while preserving environment and account-key precedence.
- [x] 4.2 Separate model factory construction, fallback execution, transient-error handling, and per-provider concurrency limiting.
- [x] 4.3 Separate streamed tool-call sanitation, request metadata projection, redaction, and compressor wiring from model construction.
- [x] 4.4 Run model/provider/retry/compressor focused tests and add a fallback/stream-safety regression covering provider order and secret absence.

## 5. Integration and delivery

- [x] 5.1 Remove obsolete duplicate helpers and update package documentation/imports after all three boundaries are split.
- [x] 5.2 Run formatting, `go test ./...`, `go build ./...`, and the affected frontend build; verify no API, SSE, task, or model contract diff.
- [x] 5.3 Build Linux delivery artifacts, deploy with backup, restart, run health and low-cost representative smoke tests, and clean temporary data.
- [x] 5.4 Record implementation/deployment evidence in iteration docs and `done.md`, then validate and archive the OpenSpec change.

## 6. Terminology and fixed-flow simplification

- [x] 6.1 Rename new internal runtime APIs and tools toward `ppt`/`plan`/`page`/`download`/`sync`; preserve exported, JSON, SSE, and historical task compatibility at boundaries.
- [x] 6.2 Replace the fixed render Graph wrapper with a sequential PPT render entry while preserving validation, image preparation, worker-pool, error, and progress contracts.
- [x] 6.3 Add concise responsibility comments to public packages and complex recovery/render functions; remove obsolete terminology from active prompts and docs.
- [x] 6.4 Run contract-focused tests, full build, historical-task compatibility checks, deployment smoke, and archive evidence.
