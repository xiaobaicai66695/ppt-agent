## 1. Chat quality contract and deterministic behavior

- [x] 1.1 Validate and deduplicate chat web/image source URLs before supplement Markdown is rendered.
- [x] 1.2 Ensure retrieval-backed replies take precedence over generic chat/PPT fallbacks and unavailable capabilities are reported honestly.
- [x] 1.3 Add focused tests for contextual material requests, malformed sources, image-search degradation and streamed Markdown safety.

## 2. Chat benchmark suite

- [x] 2.1 Add a Go test chat benchmark with a production chat-reply adapter, case schema, real-model output persistence and coverage checks.
- [x] 2.2 Add at least ten independent test chat cases, rubric and offline fixture/coverage validation.
- [x] 2.3 Add at least ten independent validation holdout chat cases and enforce test/validation request isolation.

## 3. Verification and delivery

- [x] 3.1 Run focused Go tests, chat test/validation benchmark checks, frontend validation where affected, build and OpenSpec strict validation.
- [x] 3.2 Build Linux deliverables, deploy with backup, restart, smoke chat retrieval/degradation behavior, clean temporary data and record release evidence.
- [x] 3.3 Commit changes by feature, push `master`, update completion archive and document the final deployment evidence.
