## Why

The Planner, Reviewer/Refiner, and Fixer benchmarks use fixed local fixtures, so they do not continuously reflect production requests, real follow-up corrections, or the exact stage artifacts that led to a delivery outcome. Production task metadata alone cannot recreate those cases because draft plans, review advice, scoped repairs, and their versioned inputs live in task work directories or transient workflow state.

## What Changes

- Capture immutable, privacy-gated evaluation snapshots at Planner, Reviewer, PlannerRefiner, and Fixer stage boundaries for production PPT tasks.
- Persist structured stage inputs, outputs, provenance, feedback signals, and content-addressed artifacts independently from operational task records and disposable work directories.
- Provide an admin-only candidate review and dataset publishing workflow that turns approved snapshots into versioned `active-dev` or hidden `validation` benchmark bundles.
- Provide a checksum-verified local pull workflow and benchmark dataset resolution for imported Planner, Reviewer, and Fixer cases.
- Preserve the existing Router benchmark fixtures and exclude Router from the online snapshot/promotion workflow.
- Enforce redaction, withdrawal, split isolation, and immutable versioning so production content is not copied wholesale into development environments.

## Capabilities

### New Capabilities

- `online-agent-evaluation-snapshots`: Capture, retain, redact, and withdraw stage-level production evidence for Planner, Reviewer/Refiner, and Fixer.
- `versioned-benchmark-dataset-sync`: Review candidate evidence, publish immutable development or validation datasets, and import verified bundles into a local benchmark registry.

### Modified Capabilities

- `ppt-agent-quality-eval`: Extend benchmark dataset resolution to support imported, versioned Planner/Reviewer/Fixer case bundles while retaining the fixed Router suite.

## Impact

- Affected backend areas: task workflow and persistence, database schema, admin-only evaluation APIs, artifact storage, and `cmd/pptbench` dataset loading.
- Adds production-side evaluation metadata and artifact storage; no production user database or task output directory is directly exposed to development machines.
- Adds local private benchmark imports under an ignored directory and a local registry; fixed repository fixtures remain the regression baseline.
