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

def generate(prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "", kicker: str = "", title: str = "", lede: str = "", context_block: str = "", solution_block: str = "", metrics: list[dict] = None, takeaway: str = "", background: str = None, glass_colors: dict = None, **kwargs: Any) -> Presentation:
    components = [component("paragraph", title="??", text=context_block), component("feature_card", title="??", body=solution_block), component("callout", title="??", body=takeaway)]
    for metric in metrics or []:
        components.append(component("kpi_metric", title=metric.get("label"), text=metric.get("value"), body=metric.get("trend"), data=metric))
    components.extend(kwargs.get("components") or [])
    return render_component_slide(prs, palette, source, title, lede, kicker, components, "example_detail", "", background)
