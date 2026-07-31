## Why

The Visual Designer skill asks for rich, data-backed, visually structured PPT pages, but the template JSON schemas, generator docs, and actual generator parameters drift apart. This causes agents and frontend composition to fall back to weak fields such as `title` and `bullets`, reducing quality without requiring online QA.

## What Changes

- Add template metadata for capacity, required fields, best/avoid usage, and overflow strategy.
- Confirm a default background policy: title/section/hero can use backgrounds; information-dense pages default to clean backgrounds.
- Add local icon/shape primitive guidance without external image search.
- Start by upgrading representative high-use templates instead of rewriting every generator.

## Capabilities

### New Capabilities

- `visual-designer-contract`: Template schema and design contract metadata for stable PPT generation without default online QA.

### Modified Capabilities

<!-- No existing OpenSpec visual designer spec exists yet. -->

## Impact

- `ppt-agent/skills/visual_designer/templates/single-page/*.json`: add contract metadata to selected core templates.
- `ppt-agent/skills/visual_designer/references/generators.md`: document contract metadata and background/icon policy.
- `docs/issues/todo.md` and `docs/xq-todo.md`: track confirmed decisions and remaining questions.
