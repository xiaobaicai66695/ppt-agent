## 1. Task API Contract

- [x] 1.1 Add task ownership middleware for every task-specific route and register the missing profile summary route
- [x] 1.2 Normalize authenticated user lookup for feedback, insights, and recommendations handlers
- [x] 1.3 Normalize frontend API error handling and preference form array serialization
- [x] 1.4 Add focused backend tests for owner, non-owner, administrator, and authenticated user-context behavior

## 2. SSE And Progress State

- [x] 2.1 Add monotonically increasing SSE event IDs and Last-Event-ID incremental replay
- [x] 2.2 Complete frontend event types and render system-step and thumbnail lifecycle events without duplicate logs
- [x] 2.3 Synchronize polling terminal status into task summaries and make planning/elapsed progress reactive
- [x] 2.4 Add focused tests for event IDs, replay filtering, and task terminal-state recovery helpers

## 3. Progressive Preview Delivery

- [x] 3.1 Add a TaskManager file-ready callback that prepares thumbnails asynchronously and broadcasts ready/error events
- [x] 3.2 Make thumbnail generation recheck cache after locking to avoid duplicate conversion
- [x] 3.3 Remove full-PPTX preloading and implement stable per-slide thumbnail loading, retry, ready, and failure states
- [x] 3.4 Add focused thumbnail and file-ready callback tests that do not require a live model

## 4. Template And Compose Experience

- [x] 4.1 Add a repeatable local preset-thumbnail generation script and real preview assets for every full-deck preset
- [x] 4.2 Derive layout display names and required-field guidance from the template API instead of a partial hardcoded map
- [x] 4.3 Make Compose responsive across desktop, tablet, and phone layouts while preserving all editing actions

## 5. Dashboard Workbench Experience

- [x] 5.1 Make user progress primary, move RuntimeMeta/Timeline into a collapsible diagnostics region, and remove disabled-QA copy
- [x] 5.2 Add responsive Dashboard navigation drawer, single-column narrow layout, and full-width mobile chat
- [x] 5.3 Replace touched non-semantic click targets and undersized icon controls with keyboard and touch accessible controls

## 6. Verification And Tracking

- [x] 6.1 Run focused Go tests/build, frontend type/build checks, and preset asset/URL validation
- [x] 6.2 Validate the OpenSpec change strictly and update `docs/issues/todo.md` with implementation outputs and final status
