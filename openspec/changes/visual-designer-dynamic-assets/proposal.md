## Why

The current Visual Designer generators can create valid PPT slides, but many slides still feel rigid: text is locked to fixed regions, sparse content does not scale up, and visual elements rely on simple cards, bars, circles, or text glyphs. This creates an obvious AI-generated feeling even when the content is correct.

## What Changes

- Add a local visual asset library for icons, editorial backgrounds, and subtle patterns.
- Add shared generator helpers for content density, dynamic font sizing, and alignment/layout intent.
- Use local icons and soft backgrounds in representative high-use generators.
- Keep external image search and online generative assets out of the default flow.

## Capabilities

### New Capabilities

- `visual-designer-dynamic-assets`: Local assets and density-aware generator behavior for richer PPT slides.

### Modified Capabilities

- `visual-designer-contract`: Uses existing template contract decisions to drive concrete generator behavior.

## Impact

- `ppt-agent/skills/visual_designer/assets/`: new local icon/background/pattern assets and manifest.
- `ppt-agent/skills/visual_designer/generators/asset_manager.py`: local asset selection and placement helpers.
- `ppt-agent/skills/visual_designer/generators/layout_intelligence.py`: content density and dynamic sizing helpers.
- Selected high-use generators: title, section divider, content, card grid, icon grid, quote, summary.
- `references/generators.md` and `SKILL.md`: document the new local asset and dynamic layout policy.
