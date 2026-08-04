## Why

Generated decks can drift between incompatible visual systems when a page background causes the SlideExecutor to replace the deck palette selected by the user or recommender. Thumbnail delivery also reports false conversion failures when the generated PPTX filename differs from the manifest by whitespace or punctuation even though both the PPTX and rendered JPG exist.

## What Changes

- Make the task manifest theme the immutable deck-wide palette for every slide, including slides with image backgrounds.
- Pass the backend's explicit background decision into the planning prompt and prevent the model from inventing backgrounds when the selected mode did not enable them.
- Require SlideExecutor and continuation generation to preserve empty backgrounds and use the manifest `output_file` verbatim.
- Reconcile uniquely identifiable page artifacts whose filenames drifted from the manifest, including matching rendered JPG files, before marking pages complete.
- Add frontend delivery fallback so legacy tasks use the actual ready artifact filename when it can be matched to the same page.
- Add focused prompt, manifest reconciliation, and frontend merge tests.

## Capabilities

### New Capabilities

- `ppt-deck-style-integrity`: Keeps the selected template palette and background policy consistent across every generated slide.
- `ppt-slide-artifact-reconciliation`: Reconciles generated PPTX and thumbnail filenames with the task manifest and preserves legacy task preview availability.

### Modified Capabilities

None.

## Impact

- Backend prompt data and DeepAgent instruction construction.
- Master and SlideExecutor prompt templates.
- Manifest output reconciliation and its filesystem behavior.
- Frontend slide delivery merge behavior for existing tasks.
- Existing task files may be atomically renamed only when one unambiguous page candidate exists.
