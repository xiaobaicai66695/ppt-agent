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

def generate(prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "", kicker: str = "", title: str = "", context: str = "", problem: str = "", solution: str = "", results: list[dict] = None, background: str = None, glass_colors: dict = None, **kwargs: Any) -> Presentation:
    components = [component("paragraph", title="??", text=context), component("callout", title="??", body=problem), component("feature_card", title="??", body=solution)]
    for r in results or []:
        components.append(component("kpi_metric", title=r.get("metric"), text=r.get("value"), body=r.get("comparison"), data=r))
    components.extend(kwargs.get("components") or [])
    return render_component_slide(prs, palette, source, title, context, kicker, components, "case_study", "", background)
