"""Small content-density helpers for Visual Designer generators."""

from __future__ import annotations

import re
from typing import Iterable


def weighted_text_len(text: str) -> int:
    text = text or ""
    total = 0
    for ch in text.strip():
        if ch.isspace():
            continue
        total += 2 if "\u4e00" <= ch <= "\u9fff" else 1
    return total


def number_count(text: str) -> int:
    return len(re.findall(r"\d+(?:\.\d+)?%?", text or ""))


def density_level(items: Iterable[str] = (), title: str = "", body: str = "") -> str:
    values = list(items or [])
    score = weighted_text_len(title) * 0.25 + weighted_text_len(body)
    score += sum(weighted_text_len(v) for v in values)
    score += len(values) * 18
    if len(values) <= 2 and score < 110:
        return "sparse"
    if len(values) <= 4 and score < 230:
        return "normal"
    return "dense"


def title_font_size(title: str, base: int = 36, sparse_boost: int = 6, max_size: int = 48) -> int:
    length = weighted_text_len(title)
    if length <= 18:
        return min(max_size, base + sparse_boost)
    if length <= 32:
        return base
    if length <= 48:
        return max(30, base - 4)
    return max(26, base - 8)


def focal_font_size(text: str, base: int = 32, max_size: int = 54, min_size: int = 22) -> int:
    length = weighted_text_len(text)
    if length <= 18:
        return max_size
    if length <= 32:
        return min(max_size, base + 10)
    if length <= 60:
        return base
    return max(min_size, base - 6)


def body_font_size(items: Iterable[str], base: int = 15) -> int:
    values = list(items or [])
    avg = sum(weighted_text_len(v) for v in values) / max(len(values), 1)
    if len(values) <= 3 and avg <= 32:
        return base + 3
    if avg > 70 or len(values) >= 6:
        return max(12, base - 2)
    return base


def card_layout_for_count(count: int, requested: str = "") -> str:
    if requested and requested != "auto":
        return requested
    if count <= 2:
        return "1x2"
    if count <= 4:
        return "2x2"
    return "2x3"


def alignment_for_density(level: str, role: str = "content") -> str:
    if role in {"quote", "title", "section"}:
        return "center"
    if level == "sparse":
        return "center"
    return "left"


def short_items(items: Iterable[str], max_items: int) -> list[str]:
    return [str(item).strip() for item in list(items or [])[:max_items] if str(item).strip()]


def balanced_band_top(region_top: float, region_height: float, content_height: float, min_top: float | None = None) -> float:
    top = region_top + max(0.0, (region_height - content_height) / 2)
    if min_top is not None:
        return max(min_top, top)
    return top


def gap_for_count(count: int, region_height: float, item_height: float, min_gap: float = 0.18, max_gap: float = 0.52) -> float:
    if count <= 1:
        return 0.0
    remaining = max(0.0, region_height - count * item_height)
    return max(min_gap, min(max_gap, remaining / (count - 1)))


def fitted_grid_rows(count: int, cols: int) -> int:
    if cols <= 0:
        return 1
    return max(1, (max(count, 1) + cols - 1) // cols)
