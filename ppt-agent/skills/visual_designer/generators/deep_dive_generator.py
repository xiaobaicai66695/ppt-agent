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

def generate(prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "", kicker: str = "", title: str = "", lede: str = "", left_header: str = "", key_points: list[str] = None, analysis: list[str] = None, right_header: str = "", case_example: list[str] = None, data_evidence: list[str] = None, supplement: list[str] = None, background: str = None, glass_colors: dict = None, **kwargs: Any) -> Presentation:
    components = [component("fact_card", title=left_header or "关键事实", body=" / ".join(key_points or [])), component("insight", title="分析判断", body=" / ".join(analysis or [])), component("case_snapshot", title=right_header or "案例证据", body=" / ".join((case_example or []) + (data_evidence or [])))]
    if supplement:
        components.append(component("callout", title="补充说明", body=" / ".join(supplement)))
    components = prefer_explicit_components(kwargs, components)
    return render_component_slide(prs, palette, source, title, lede, kicker, components, "deep_dive", "", background)
