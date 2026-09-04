## 1. Benchmark coverage

- [x] 1.1 Expand the test and validation fixtures so each router, planner, reviewer and fixer suite contains ten independent valid cases.
- [x] 1.2 Add no-model benchmark loader tests for per-suite counts, non-empty unique IDs and test/validation isolation; document the 10-case baseline.

## 2. Delivery-feedback backend

- [x] 2.1 Add the task-feedback persistence model, automatic migration, task response projection and owner-scoped lookup.
- [x] 2.2 Add a validated, idempotent owner-only feedback API for completed deliveries and focused API/persistence tests.

## 3. Delivery-feedback workbench

- [x] 3.1 Add frontend feedback API/types and a responsive lower-left Dashboard feedback control with rating, optional suggestion, restore and update states.

## 4. Verification and delivery

- [x] 4.1 Run focused Go tests, `go build ./...`, frontend build and OpenSpec strict validation.
- [x] 4.2 Build Linux deliverables, deploy, restart, run the smallest feedback smoke test, clean test data and record deployment evidence.
