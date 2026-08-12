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

def generate(prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "", kicker: str = "", title: str = "", subtitle: str = "", headers: list[str] = None, rows: list[list[str]] = None, recommendation: str = "", background: str = None, glass_colors: dict = None, **kwargs: Any) -> Presentation:
    components = [component("table", title=title, data={"headers": headers or [], "rows": rows or []})]
    if recommendation:
        components.append(component("callout", title="??", body=recommendation))
    components.extend(kwargs.get("components") or [])
    return render_component_slide(prs, palette, source, title, subtitle, kicker, components, "comparison_table", "", background)
