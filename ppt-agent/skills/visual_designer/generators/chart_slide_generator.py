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

def generate(prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "", kicker: str = "", title: str = "", subtitle: str = "", chart_type: str = "bar", data: dict = None, show_legend: bool = True, analysis: str = None, background: str = None, glass_colors: dict = None, **kwargs: Any) -> Presentation:
    chart_data = data or {"labels": ["A", "B", "C"], "datasets": [{"name": "指标", "values": [3, 5, 4]}]}
    chart_data.setdefault("chart_type", chart_type)
    components = [component("chart", title=title, data=chart_data)]
    if analysis:
        components.append(component("insight", title="分析", body=analysis))
    components = prefer_explicit_components(kwargs, components)
    return render_component_slide(prs, palette, source, title, subtitle, kicker, components, "chart_slide", chart_type, background)
