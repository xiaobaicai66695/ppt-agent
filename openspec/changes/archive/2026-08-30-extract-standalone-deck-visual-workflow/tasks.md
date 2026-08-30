## 1. Skill contract and standalone workflow

- [x] 1.1 Move platform-neutral visual policy and materialization workflow from the backend Planner prompt into the standalone skill and component contract.
- [x] 1.2 Extend the standalone Unsplash resolver to hydrate both visual intents and foreground image components atomically.
- [x] 1.3 Enforce visual asset validation from standalone validation and full-deck render entry points; update examples and references.

## 2. Backend Planner adaptation

- [x] 2.1 Remove duplicate Planner-directed image download behavior while preserving complete visual query planning and deterministic backend materialization.
- [x] 2.2 Update focused Prompt and visual-policy tests for the new responsibility boundary.

## 3. Quality evidence and delivery

- [x] 3.1 Run standalone unit/render smoke tests, backend focused tests, and offline gold DeckSpec benchmark.
- [x] 3.2 Run a real Planner first-draft benchmark when configured; otherwise record the concrete credential/configuration block.
- [x] 3.3 Build and deploy the runtime change, run a minimal production smoke test, clean temporary artifacts, and record deployment evidence before archiving.
