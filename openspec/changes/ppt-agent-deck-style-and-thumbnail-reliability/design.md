## Context

The selected preset or recommendation produces a deck template, palette, and explicit background policy before DeepAgent planning starts. Today the master prompt can still infer a different background from the topic, and the SlideExecutor prompt then calls `get_palette_for_background`, replacing the manifest theme on individual pages. Separately, completion reconciliation only accepts an exact `output_file`, while the model sometimes reconstructs filenames from titles and introduces whitespace differences. The renderer succeeds, but the UI requests a nonexistent manifest filename and labels the result as a conversion failure.

## Goals / Non-Goals

**Goals:**

- Preserve one deck-wide palette selected before generation.
- Keep background usage equal to the backend's explicit outline/recommendation decision.
- Make filename drift recoverable when the page identity is unambiguous.
- Keep existing task previews usable even before their artifacts are normalized.

**Non-Goals:**

- Redesign the visual generator layouts or add new background assets.
- Guess among multiple PPTX files for the same page.
- Run the retired online QA/Reviewer flow.

## Decisions

1. The manifest `theme` is the source of truth for every generator call. Background selection affects image treatment and contrast but never selects another palette. This keeps a preset visually coherent and also keeps recommended mode deterministic.
2. `TemplateData` carries `OutlineUseBackground` and `OutlineBackground`. The master prompt may preserve or apply that explicit choice, but may not infer a different background from the topic. Existing per-page backgrounds supplied by a user-authored outline remain intact when background use is disabled.
3. SlideExecutor must use `tasks.json.theme`, preserve an empty `background`, and save to the task's `output_file` verbatim. Prompt examples will demonstrate these invariants and remove `get_palette_for_background` guidance.
4. Completion reconciliation first checks the exact output path. If absent, it searches for a unique PPTX candidate with the same page prefix, atomically renames it to the declared path, and renames an existing `qa_images/<stem>.jpg` in the same operation sequence. Multiple candidates remain unresolved and visible as an error.
5. Frontend delivery merging uses the actual ready filename when a legacy artifact can be associated with a manifest page by page prefix. This is a compatibility fallback, not the primary normalization path.

## Risks / Trade-offs

- [A background may be less harmonious with the locked palette] -> Background choice is made together with the palette by recommendation; preset mode only uses explicitly enabled backgrounds, and generators retain contrast adaptation.
- [Filename normalization could rename the wrong artifact] -> Rename only when exactly one page-prefixed PPTX candidate exists and never overwrite an existing target.
- [PPTX rename succeeds but JPG rename fails] -> Keep the PPTX canonical, report the JPG rename error, and allow thumbnail regeneration from the canonical file.
- [Older tasks use unusual page prefixes] -> Retain exact filename matching first and keep unresolved cases visible rather than guessing.

## Migration Plan

1. Deploy prompt and reconciliation changes with focused tests.
2. Reconcile the reported task once to normalize pages 6 and 9 and their JPGs.
3. Restart the backend and verify health, task delivery metadata, and thumbnail endpoints.
4. Roll back application code if needed; normalized filenames remain valid because they match the manifest contract.

## Open Questions

None for this change.
