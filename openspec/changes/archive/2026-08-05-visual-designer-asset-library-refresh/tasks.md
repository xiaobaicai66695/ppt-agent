## 1. Asset Contract And Sources

- [x] 1.1 Add a reproducible external asset synchronization script with curated Icons8, Unsplash/Picsum, and Transparent Textures sources
- [x] 1.2 Replace the existing `assets` icon, editorial background, and pattern files and generate the v2 asset manifest
- [x] 1.3 Add `background_templates/manifest.json` and supplement every underfilled theme to at least four images

## 2. Generator Integration

- [x] 2.1 Refactor `asset_manager.py` to validate manifest v2 and resolve icon semantics from manifest keywords without unrelated fallbacks
- [x] 2.2 Refactor `background_manager.py` to consume the background theme manifest while preserving legacy theme and direct-path compatibility
- [x] 2.3 Replace hardcoded `layout`, `primitive`, `review`, and `align` decorative fallbacks in affected generators with semantic icons or omission
- [x] 2.4 Add categorized `photo` assets and deterministic semantic/default photo resolution for image-text slides

## 3. Documentation And Verification

- [x] 3.1 Update the Visual Designer README/SKILL, background template guide, and generator reference for the new structure, attribution, sync, and validation workflow
- [x] 3.2 Add focused asset manifest, semantic mapping, and background catalog tests and run Python compilation/tests
- [x] 3.3 Generate and render representative slide types, including explicit/default image-text photos, inspect icon semantics/background crops/overlap, and fix visible regressions
- [x] 3.4 Record completion in the project issue archive and iteration log with validation evidence
