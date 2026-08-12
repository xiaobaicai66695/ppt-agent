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

def generate(prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "", kicker: str = "", title: str = "", lede: str = "", left_header: str = "", key_points: list[str] = None, analysis: list[str] = None, right_header: str = "", case_example: list[str] = None, data_evidence: list[str] = None, supplement: list[str] = None, background: str = None, glass_colors: dict = None, **kwargs: Any) -> Presentation:
    components = [component("feature_card", title=left_header or "????", body=" / ".join(key_points or [])), component("feature_card", title="??", body=" / ".join(analysis or [])), component("feature_card", title=right_header or "????", body=" / ".join((case_example or []) + (data_evidence or [])))]
    if supplement:
        components.append(component("callout", title="??", body=" / ".join(supplement)))
    components.extend(kwargs.get("components") or [])
    return render_component_slide(prs, palette, source, title, lede, kicker, components, "deep_dive", "", background)
