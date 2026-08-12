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

def generate(prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "", title: str = "", subtitle: str = "", kicker: str = "", layout_variant: str = "", background: str = None, glass_colors: dict = None, **kwargs: Any) -> Presentation:
    components = list(kwargs.get("components") or [])
    for key, value in kwargs.items():
        if key not in {"components"} and value not in (None, "", [], {}):
            components.append(component("feature_card", title=key, body=str(value)))
    return render_component_slide(prs, palette, source, title, subtitle, kicker, components, "swot_analysis", layout_variant, background)
