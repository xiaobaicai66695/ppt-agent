## 1. Content Plan And Main Agent Contract

- [x] 1.1 Add tolerant `ContentElement.items` decoding for strings, scalars, objects, and mixed arrays with focused Go tests.
- [x] 1.2 Split outline validation from optional legacy enrichment so task creation never performs a separate model-based template-fill call.
- [x] 1.3 Mark template scaffolds versus user outlines and update the main Agent prompt to rewrite scaffold examples while preserving populated user fields.
- [x] 1.4 Add handler and prompt tests for blank, partially populated, fully populated, and object-item outlines.

## 2. Template Recommendation And Server-Side Resolution

- [x] 2.1 Implement deterministic preset, theme, and background recommendation against current loader resources with pure-function tests.
- [x] 2.2 Extend task creation with backward-compatible `recommended` and `preset` template selections resolved on the server.
- [x] 2.3 Apply recommended backgrounds only to visual page types and expose the selected strategy in task/runtime metadata.
- [x] 2.4 Add API tests for valid preset, invalid preset, recommendation fallback, and no-extra-model task creation.

## 3. Single-Choice Creation Experience

- [x] 3.1 Extend frontend API types and creation calls for server-side template selection.
- [x] 3.2 Rebuild HomePage to load and show smart recommendation, every live preset, and custom composition beside the prompt.
- [x] 3.3 Submit smart/preset choices directly to task creation and route to the created Dashboard task without a second selection.
- [x] 3.4 Make custom composition open an editable outline directly and remove full-deck template selection from that path.
- [x] 3.5 Remove duplicate template controls from the Dashboard conversation composer and cover selection utilities with unit tests.

## 4. Verification And Delivery

- [x] 4.1 Run focused Go tests/build, frontend unit/build checks, and strict OpenSpec validation.
- [x] 4.2 Verify prompt-page template loading, direct task creation, custom composition, and responsive behavior in a browser.
- [x] 4.3 Update iteration records and TODO completion evidence, then redeploy the Linux backend and frontend to `/ppt/ppt-agent` with health/API checks.
