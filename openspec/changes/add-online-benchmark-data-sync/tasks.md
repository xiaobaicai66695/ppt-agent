## 1. Evaluation persistence foundation

- [x] 1.1 Add evaluation session, stage-run, artifact, candidate, case revision, dataset, and export persistence models with migrations and focused repository tests.
- [x] 1.2 Add a content-addressed evaluation artifact store and stage-capture service that records canonical JSON evidence, provenance, redaction state, and capture failures without interrupting delivery.

## 2. Production workflow capture

- [x] 2.1 Wire Planner, Reviewer, PlannerRefiner, and Fixer workflow boundaries to persist their exact structured input/output snapshots and parent/attempt relationships.
- [x] 2.2 Persist final task outcome and later feedback signals into the evaluation session, and add explicit withdrawal handling that blocks future publication.

## 3. Candidate review and dataset publication

- [x] 3.1 Add authenticated admin APIs for listing/redacting/approving candidates, saving explicit case revisions, withdrawing evidence, and inspecting immutable provenance.
- [x] 3.2 Add dataset version publication and checksum-verified bundle export with split-isolation rules for core, active-dev, and validation roles.

## 4. Local benchmark import and resolution

- [x] 4.1 Add `pptbench dataset pull` and private local registry/import storage with manifest checksum verification.
- [x] 4.2 Extend `pptbench` dataset selection to resolve imported Planner, Reviewer, and Fixer bundles while retaining Router's fixture-only rule.

## 5. Verification and delivery

- [x] 5.1 Add focused unit/integration tests for capture, privacy gates, dataset split isolation, export/import integrity, and imported-case execution.
- [x] 5.2 Update benchmark operations documentation and the completion record with dataset lifecycle, deployment evidence, and retention/withdrawal constraints.
- [x] 5.3 Run focused tests and build, commit the change, deploy the Linux delivery, execute a minimal production smoke capture and bundle pull, clean smoke data, and record the result.
