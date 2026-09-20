## Context

Production and development run different binaries and retain different task data. The production MySQL database records task, conversation, feedback, and runtime-event metadata, while Planner drafts, review checkpoints, final plans, and visual QA evidence are written to per-task work directories. The existing `pptbench` command reads only repository fixtures, with fixed `test` and `validation` roots.

The new workflow must collect the exact structured context that affects Planner, Reviewer/PlannerRefiner, and Fixer decisions without exposing production databases, credentials, or arbitrary work directories to developer machines. Router remains on its existing static fixture coverage.

## Goals / Non-Goals

**Goals:**

- Durably capture an immutable, redacted stage snapshot after each Planner, Reviewer, PlannerRefiner, and Fixer boundary.
- Preserve the canonical stage input/output and provenance needed to derive a replayable benchmark case, rather than attempting to reconstruct it from a final task state.
- Let an administrator inspect candidates, attach approved case content, freeze a versioned dataset, and export it as a checksum-protected bundle.
- Let `pptbench` import and resolve a named local dataset without changing repository fixtures.
- Keep `core` cases stable, permit prior validation bundles to become `active-dev`, and prevent any published-development case from returning to validation.

**Non-Goals:**

- Replaying production Router data, copying the production MySQL database, or bulk-downloading `weboutput`.
- Automatically treating user feedback or a model output as a golden expected result.
- Exporting raw user uploads, API keys, secret-bearing traces, or full PPTX artifacts by default.
- Building a public-facing evaluation administration UI in the first release; authenticated admin APIs and a CLI-compatible bundle are sufficient.

## Decisions

### Capture structured stage snapshots, not generic logs

The workflow records `planner`, `reviewer`, `planner_refiner`, and `fixer` stages explicitly. Each stage record points to immutable JSON artifacts for canonical input and output, and stores attempt ordering, status, elapsed time, prompt/model/contract provenance, and a parent stage run where applicable.

This is preferred over parsing `RuntimeEventRecord` because timeline events are user-visible diagnostics, have an evolving payload shape, and do not guarantee the complete pre/post plan boundary required by benchmark cases.

### Keep artifact bytes outside MySQL

`eval_artifacts` stores a SHA-256, content type, byte count, redaction status, and a relative storage key. The JSON bytes live under a configured evaluation artifact root. Stage capture writes a temporary artifact, hashes it, atomically publishes it, and only then marks its database record available.

Storing large JSON and visual artifacts in MySQL would inflate operational backups and make retention/deletion unnecessarily expensive. The initial release stores only JSON evidence; rendered slides and user-uploaded files remain out of scope.

### Capture at workflow boundaries and update outcome separately

Capture occurs after the Planner draft is written, after a Reviewer report, after a Refiner patch/review, and before/after a Fixer patch. A later task completion or feedback update augments the session's outcome metadata without mutating the prior stage evidence.

This prevents a final `tasks.json` from erasing the draft and advice that explain a later result. A unique task id plus stage plus attempt makes capture idempotent.

### Separate candidate approval from dataset publication

Snapshots first enter a candidate queue. An admin approves a candidate only after reviewing/redacting the canonical evidence and providing a benchmark case JSON with explicit expected/rubric fields. Dataset versions pin a candidate case revision and a case's stable split group. A validation publication rejects candidates previously published in any development dataset.

This makes the evaluation lifecycle auditable and prevents holdout leakage. A production session may have many stage runs but only an approved, scoped case may be exported.

### Export bundles through an admin-only API and import locally

Publishing creates a manifest containing dataset metadata, each approved case JSON, relative artifact references, and SHA-256 values. The server creates a ZIP bundle in the evaluation artifact root. The local `pptbench dataset pull` command downloads it using an explicit endpoint and admin bearer token, verifies the manifest/files, and writes it under an ignored local import root plus a small local registry JSON.

Direct MySQL access and copying remote work directories are rejected: they unnecessarily expose unrelated production state and cannot prove a downloaded dataset was frozen as evaluated.

## Risks / Trade-offs

- [Stage capture misses a boundary because a workflow exits early] → Capture helpers are invoked around every authoritative write and have focused unit tests; captures are best-effort for delivery but record failure observably.
- [Production content contains personal or sensitive information] → Default snapshots are `pending_redaction`, exports require explicit `approved` redaction state, and withdrawal revokes future downloads.
- [Dataset changes hide a regression] → `core` remains fixed; only frozen validation bundles can promote to `active-dev`, and dataset membership is immutable after publication.
- [Local and Linux execution differ] → Bundle manifests record source build/model/contract provenance; release validation runs use the Linux delivery build as well as local focused tests.
- [Evaluation storage grows indefinitely] → Store JSON only in v1, deduplicate by SHA-256, and use future retention jobs for rejected/withdrawn snapshots.

## Migration Plan

1. Add evaluation models, migrations, artifact store, and capture service without changing task delivery semantics.
2. Wire Planner, Reviewer/Refiner, and Fixer boundaries to create private stage snapshots; verify normal generation remains available if capture storage fails.
3. Add authenticated admin candidate/dataset/export endpoints, then add local bundle pull and dataset resolution.
4. Deploy the backend and run a one- to two-page production smoke task. Inspect the captured Planner/Reviewer stage evidence, publish a minimal development bundle, pull it locally, and run one imported case.
5. If rollback is required, redeploy the preceding commit. Existing evaluation records and artifacts are additive and ignored by the prior runtime.

## Open Questions

- The initial retention period for rejected and withdrawn snapshots will be configured operationally after production volume is observed.
- A later change can add a dedicated admin UI; this change deliberately exposes the minimum API/CLI surface first.
