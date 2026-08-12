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

def generate(prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "", title: str = "", subtitle: str = "", author: str = "", date: str = "", kicker: str = "", layout_variant: str = "", background: str = None, glass_colors: dict = None, **kwargs: Any) -> Presentation:
    components = []
    if author:
        components.append(component("callout", title="???", body=author))
    if date:
        components.append(component("callout", title="??", body=date))
    components.extend(kwargs.get("components") or [])
    return render_component_slide(prs, palette, source, title, subtitle, kicker, components, "title_slide", layout_variant, background)
