# Examples

These examples are maintained as executable smoke fixtures for the standalone `ppt-planner` skill.

Run from the `ppt-agent` repository root:

```bash
python skills/ppt-planner/generators/validate_ppt.py --work-dir skills/ppt-planner/examples/minimal --skills-dir skills
python skills/ppt-planner/generators/render_ppt.py --work-dir skills/ppt-planner/examples/minimal --skills-dir skills --output ppt.pptx
```

Expected behavior:

- `minimal/` renders a one-slide text/card ppt with no image dependencies.
- `image_text_with_local_image/` embeds `sample_scene.png` from its own directory.
- `chart_slide/` renders a one-slide bar chart from structured chart data.

Generated `.pptx`, `.pdf`, and `.png` files are outputs, not source fixtures.
