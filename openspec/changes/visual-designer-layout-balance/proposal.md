# Proposal: Visual Designer Layout Balance

## Problem

Visual Designer generators still place some text blocks too high or too low, use fixed font sizes for varied content lengths, and leave several templates visually biased or sparse. The previous asset pass improved visual anchors, but did not provide a consistent text-box balancing layer across generators.

## Goals

- Add reusable layout/text intelligence helpers for dynamic font sizing, line estimation, vertical centering, and content-band positioning.
- Retrofit high-impact single-page generators so text boxes self-balance inside their available regions.
- Improve placeholder-heavy visual areas such as `image_text` and timeline node rendering.
- Keep public `generate_*` signatures stable for SlideExecutor compatibility.
- Generate and render every single-page template as a smoke deck, then iterate until there are no obvious visual issues.

## Non-Goals

- No online QA model loop.
- No external image search or paid asset generation.
- No broad backend or frontend changes unless generator contracts require them.
