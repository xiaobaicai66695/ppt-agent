## Context

The user feedback identified three concrete design failures:

- Text is concentrated in fixed regions even when a centered or more balanced composition would be better.
- Sparse content does not dynamically scale, leaving slides visually hollow.
- Repeated simple primitives make the deck feel machine-made and low energy.

The project already confirmed local primitives first, selective background use, and no default external image search.

## Goals / Non-Goals

**Goals:**

- Add a local asset library that is safe to use offline.
- Add reusable helpers so generator improvements are consistent.
- Improve a representative set of generators without rewriting the entire skill.
- Preserve existing generator function signatures.

**Non-Goals:**

- Do not introduce external image search by default.
- Do not require online LLM or image generation to produce slides.
- Do not change the backend task schema in this slice.
- Do not rewrite all templates or all generator functions at once.

## Decisions

### Decision 1: Asset Manifest First

Assets are tracked through `assets/manifest.json`. Generators select by id, tag, or type instead of hard-coding individual files everywhere.

### Decision 2: Generated Local Assets Are Acceptable

The first local asset pack is generated offline into PNG files with a consistent low-saturation style. This keeps the skill self-contained and avoids licensing uncertainty.

### Decision 3: Density-Aware Behavior Is Opt-In Per Generator

Shared helpers calculate density and dynamic sizes, but each generator decides how to apply them to avoid unintended global layout regressions.

### Decision 4: Use Existing Python Generator Surface

The public `generate_*` function signatures remain compatible. New behavior is driven by existing fields such as `title`, `bullets`, `cards`, `icons`, and `background`.

## Risks / Trade-offs

- [Risk] Generated assets may still feel generic. -> Mitigation: keep assets tagged and replaceable; use them as a first local style layer.
- [Risk] Dynamic sizing can create wrapping issues. -> Mitigation: clamp font sizes and validate with generated sample decks.
- [Risk] Adding helper imports can break packaging. -> Mitigation: keep helpers inside `generators/` and run `py_compile`.
