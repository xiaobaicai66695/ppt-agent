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

def generate(prs: Optional[Presentation] = None, palette: str = "ocean_soft", source: str = "", title: str = "总结", key_points: list[str] = None, thank_you: str = "感谢聆听", contact: str = "", kicker: str = "", background: str = None, glass_colors: dict = None, **kwargs: Any) -> Presentation:
    components = [component("key_point", title=f"{i+1:02d}", body=item) for i, item in enumerate(key_points or [])]
    if thank_you:
        components.append(component("callout", title=thank_you, body=contact))
    components = prefer_explicit_components(kwargs, components)
    return render_component_slide(prs, palette, source, title, "", kicker or "总结", components, "summary_slide", "", background)
