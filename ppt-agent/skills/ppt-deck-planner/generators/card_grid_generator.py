from __future__ import annotations

from typing import Any, Optional
from pptx import Presentation

from .component_layout import component, prefer_explicit_components, render_component_slide


def _clean_items(values):
    result = []
    for value in values or []:
        if isinstance(value, dict):
            result.append(value)
        elif value:
            result.append({"text": str(value)})
    return result

def generate(prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "", title: str = "", cards: list[dict] = None, layout: str = "", layout_variant: str = "", kicker: str = "", subtitle: str = "", background: str = None, glass_colors: dict = None, **kwargs: Any) -> Presentation:
    components = [component("feature_card", title=c.get("header") or c.get("title"), body=c.get("body") or c.get("desc"), emphasis=c.get("emphasis", "")) for c in (cards or [])]
    components = prefer_explicit_components(kwargs, components)
    return render_component_slide(prs, palette, source, title, subtitle, kicker, components, "card_grid", layout_variant or layout, background)
