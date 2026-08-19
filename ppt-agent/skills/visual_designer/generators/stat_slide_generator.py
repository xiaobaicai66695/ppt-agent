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

def generate(prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "", title: str = "", stats: list[dict] = None, kicker: str = "", subtitle: str = "", background: str = None, glass_colors: dict = None, **kwargs: Any) -> Presentation:
    components = [component("kpi_metric", title=s.get("label"), text=str(s.get("number", "")) + str(s.get("unit", "")), body=s.get("trend", ""), data=s) for s in (stats or [])]
    components = prefer_explicit_components(kwargs, components)
    return render_component_slide(prs, palette, source, title, subtitle, kicker, components, "stat_slide", "", background)
