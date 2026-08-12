from __future__ import annotations

from typing import Any, Optional
from pptx import Presentation

from .component_layout import component, render_component_slide


def _clean_items(values):
    result = []
    for value in values or []:
        if isinstance(value, dict):
            result.append(value)
        elif value:
            result.append({"text": str(value)})
    return result

def generate(prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "", title: str = "", subtitle: str = "", nodes: list[dict] = None, steps: list[dict] = None, direction: str = "", kicker: str = "", background: str = None, glass_colors: dict = None, **kwargs: Any) -> Presentation:
    raw = nodes or steps or []
    components = [component("timeline_node", title=item.get("year") or item.get("num") or item.get("title"), body=item.get("event") or item.get("desc") or item.get("body")) for item in raw]
    components.extend(kwargs.get("components") or [])
    return render_component_slide(prs, palette, source, title, subtitle, kicker, components, "timeline", direction, background)
