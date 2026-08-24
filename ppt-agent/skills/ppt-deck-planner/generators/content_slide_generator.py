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

def generate(prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "", title: str = "", section_header: str = "", bullets: list[str] = None, kicker: str = "", lede: str = "", highlight_stats: list[dict] = None, layout_variant: str = "", background: str = None, glass_colors: dict = None, **kwargs: Any) -> Presentation:
    components = []
    if section_header:
        components.append(component("subheadline", text=section_header))
    if lede:
        components.append(component("paragraph", text=lede))
    if bullets:
        components.append(component("bullet_list", items=bullets))
    for stat in highlight_stats or []:
        components.append(component("kpi_metric", title=stat.get("label", ""), text=stat.get("value", ""), data=stat))
    components = prefer_explicit_components(kwargs, components)
    return render_component_slide(prs, palette, source, title, lede, kicker, components, "content_slide", layout_variant, background)
