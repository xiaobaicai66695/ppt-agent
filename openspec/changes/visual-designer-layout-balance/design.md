# Design: Visual Designer Layout Balance

## Approach

1. Extend `layout_intelligence.py` with deterministic text measurement helpers:
   - weighted line estimation for Chinese/English mixed text
   - font-size fitting within width/height boxes
   - vertical centering and content-band placement
   - compact/normal/loose spacing decisions

2. Add safe helpers in `base.py`:
   - `add_text_boxed` for fit-to-box text placement
   - optional vertical anchor support while preserving existing `add_text`
   - support for line spacing and text frame margins

3. Retrofit generators where visual bias is most visible:
   - cover/narrative pages: title, section, quote, summary
   - information pages: content, card grid, two/three column, image text
   - relationship/data pages: timeline, process, KPI/stat, comparison, region/SWOT

4. Verify with a local all-template smoke script:
   - generate 24 single-page PPTX files
   - convert through LibreOffice to PDF
   - rasterize with Poppler to PNG
   - produce a contact sheet and JSON report
   - inspect high-risk pages visually and run bounds/placeholder checks

## Compatibility

Existing generator function signatures remain unchanged. New helpers are internal to `skills/visual_designer/generators`.
