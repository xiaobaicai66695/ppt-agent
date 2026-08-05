# Visual Designer Asset Library

`visual_designer` is the offline PPT rendering skill used by the Agent. Its
generators never download assets at runtime: icons, editable content photos,
editorial backgrounds, patterns, and themed background photos are committed
with versioned metadata.

## Directory Contract

```text
visual_designer/
├── assets/
│   ├── manifest.json                 # v2 asset metadata and provenance
│   ├── icons/core/                   # 512x512 transparent PNG icons
│   ├── photos/<category>/            # 1920x1080 editable content photos
│   ├── backgrounds/editorial/        # 1920x1080 JPG photos
│   └── patterns/subtle/              # 1920x1080 transparent PNG patterns
├── background_templates/
│   ├── manifest.json                 # theme, palette, scenario, image metadata
│   └── <theme>/images/*.jpg
├── generators/
│   ├── asset_manager.py              # asset lookup and semantic icon/photo matching
│   └── background_manager.py         # manifest-backed theme selection
└── scripts/sync_external_assets.py   # curated download and normalization
```

Current baseline:

- 79 Icons8 Fluency Systems Regular icons covering structure, business,
  technology, government, geography, nature, data, and closing-page concepts.
- 14 categorized content photos, 6 editorial photo backgrounds, and 4 subtle
  texture patterns.
- 6 background themes with 4-5 images per theme.

## Runtime Rules

- Use an explicit icon id when the content plan already knows the concept.
- Otherwise call `icon_id_from_text`; matching comes from manifest `keywords`
  and `tags`, using longest keyword and priority ranking.
- Unknown semantics return an empty id. Do not restore generic `layout`,
  `primitive`, `review`, or abbreviation placeholders.
- Use `photo_id_from_text` for `image_text` defaults. Explicit local paths or
  registered photo ids win; otherwise semantic matching chooses a categorized
  local photo and falls back to `photo_business_work`.
- Content photos must be inserted as a single replaceable PowerPoint picture.
  Do not synthesize an image area from icons, patterns, labels, or shape groups.
- Use a theme id such as `minimalist_blue` when any image in a theme is valid.
- Use `<theme>/images/<file>.jpg` when adjacent-page de-duplication matters.
- The runtime must remain offline. Network URLs are maintenance metadata only.

## Synchronize Sources

Run from `ppt-agent/skills/visual_designer` with Pillow installed:

```powershell
python scripts/sync_external_assets.py
python scripts/sync_external_assets.py --photos-only  # only categorized content photos
```

The script downloads every curated source into a temporary staging directory,
validates it as an image, normalizes dimensions, then replaces the four
`assets` subdirectories and regenerates both manifests. A failed download does
not leave a partially replaced asset library.

The Bing image URL in manifest metadata is a discovery entry, not a license
source. Do not import a Bing thumbnail or a reposted image unless the original
page and reuse terms are known.

## Validate

```powershell
python -m py_compile generators/*.py scripts/sync_external_assets.py
python -m unittest discover -s tests -p "test_asset_library.py"
```

For a visual change, also generate representative `title_slide`,
`section_divider`, `icon_grid`, `image_text`, `quote_slide`, and
`summary_slide` pages, render them, and inspect icon meaning, crop, contrast,
overlap, and unresolved placeholders.

## Attribution And Licenses

- Icons8: `Icons by Icons8`, subject to the
  [Icons8 License](https://icons8.com/license). Confirm the deployment's paid
  or free license obligations before removing attribution.
- Photos: each external photo records the Unsplash page, author, download URL,
  and `Unsplash License` in the relevant manifest.
- Patterns: `Transparent Textures`, recorded as CC BY 3.0 in the asset manifest.
- Existing project backgrounds without recoverable provenance are explicitly
  marked `project-existing`; they are not presented as externally licensed.

When a generated deck is distributed under a license that requires credit,
include the recorded attribution in the deck credits or accompanying material.

## Add Or Replace An Asset

1. Add a curated spec to `scripts/sync_external_assets.py`.
2. Include semantic Chinese/English keywords and a stable source page.
3. Run the synchronization script; do not hand-edit generated dimensions.
4. Run the manifest tests and a focused rendering smoke test.
5. Update `SKILL.md` or `references/generators.md` when selection behavior or
   the public generator contract changes.
