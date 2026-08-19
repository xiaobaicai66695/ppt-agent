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

def generate(prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "", title: str = "", left_header: str = "", right_header: str = "", left_bullets: list[str] = None, right_bullets: list[str] = None, left_intro: str = "", right_intro: str = "", left_sections: dict = None, right_sections: dict = None, left_items: list[dict] = None, right_items: list[dict] = None, kicker: str = "", layout_variant: str = "", background: str = None, glass_colors: dict = None, **kwargs: Any) -> Presentation:
    components = []
    components.append(component("feature_card", title=left_header or "左侧", body=left_intro or " / ".join(left_bullets or []), relation="left"))
    components.append(component("feature_card", title=right_header or "右侧", body=right_intro or " / ".join(right_bullets or []), relation="right"))
    for item in left_items or []:
        components.append(component("feature_card", title=item.get("title"), body=item.get("desc"), relation="left"))
    for item in right_items or []:
        components.append(component("feature_card", title=item.get("title"), body=item.get("desc"), relation="right"))
    components = prefer_explicit_components(kwargs, components)
    return render_component_slide(prs, palette, source, title, "", kicker, components, "two_column", layout_variant, background)
