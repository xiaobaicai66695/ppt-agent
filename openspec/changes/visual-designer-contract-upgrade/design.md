## Context

The user confirmed the current direction for background and assets: use image backgrounds selectively, and prefer local icon/shape primitives over external image search. `content_plan` ownership and quality scoring remain open, so this change avoids locking those decisions.

## Goals / Non-Goals

**Goals:**

- Add contract metadata to representative core templates.
- Document background and local primitive policy.
- Keep changes compatible with existing template loader behavior.

**Non-Goals:**

- Do not rewrite all generators.
- Do not introduce external image search or generated image assets.
- Do not define the final strong `content_plan` schema in this slice.
- Do not add LLM judge scoring.

## Decisions

### Decision 1: Additive JSON Metadata

Template JSON files accept extra keys without changing generator calls. Contract metadata is additive and can be consumed by future layout selectors.

### Decision 2: Start With High-Use Templates

The first slice covers high-use existing template files: `title_slide`, `section_divider`, `agenda`, `quote_slide`, `summary_slide`, `image_text`, `content_slide`, `card_grid`, `two_column`, `icon_grid`, `kpi_dashboard`, and `chart_slide`. `generate_image_hero` exists as a generator function, but this repository does not currently include an `image_hero.json` template file; adding that template is intentionally left for a separate slice.

## Risks / Trade-offs

- [Risk] Metadata is ignored by current code. -> Mitigation: this is still valuable as a stable source for the next layout selector change.
- [Risk] Contracts drift again. -> Mitigation: document the fields in `generators.md` and keep future generator changes synchronized.
