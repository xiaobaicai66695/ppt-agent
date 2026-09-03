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

def generate(prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "", title: str = "", layout: str = "", image_path: str = "", header: str = "", paragraph: str = "", bullets: list[str] = None, kicker: str = "", sub_header: str = "", layout_variant: str = "", background: str = None, glass_colors: dict = None, **kwargs: Any) -> Presentation:
    components = [
        component("image", title=header or sub_header, local_path=image_path),
        component("paragraph", title=header, text=paragraph or " / ".join(bullets or [])),
    ]
    components = prefer_explicit_components(kwargs, components)
    return render_component_slide(prs, palette, source, title, paragraph[:80], kicker, components, "image_text", layout_variant or layout, background)
