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

def generate(prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "", title: str = "????", items: list[str] = None, kicker: str = "??", background: str = None, glass_colors: dict = None, **kwargs: Any) -> Presentation:
    components = [component("feature_card", title=item[:8], body=item) for item in (items or [])]
    components.extend(kwargs.get("components") or [])
    return render_component_slide(prs, palette, source, title, "", kicker, components, "agenda", "", background)
