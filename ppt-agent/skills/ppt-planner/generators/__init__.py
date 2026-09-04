"""PPT slide generators - exports all generators."""
from .base import (
    PALETTES, new_presentation, set_slide_background, set_image_background,
    save_presentation, save_slide, resolve_background,
    add_text_boxed, apply_slide_transition, apply_presentation_transitions,
)
from .title_slide_generator import generate as generate_title_slide
from .section_divider_generator import generate as generate_section_divider
from .content_slide_generator import generate as generate_content_slide
from .quote_slide_generator import generate as generate_quote_slide
from .card_grid_generator import generate as generate_card_grid
from .comparison_table_generator import generate as generate_comparison_table
from .image_text_generator import generate as generate_image_text
from .agenda_generator import generate as generate_agenda
from .kpi_dashboard_generator import generate as generate_kpi_dashboard
from .kanban_generator import generate as generate_kanban
from .brand_focus_generator import generate as generate_brand_focus
from .timeline_generator import generate as generate_timeline
from .chart_slide_generator import generate as generate_chart_slide
from .swot_analysis_generator import generate as generate_swot_analysis

__all__ = [
    "PALETTES",
    "new_presentation",
    "set_slide_background",
    "set_image_background",
    "save_presentation",
    "save_slide",
    "apply_slide_transition",
    "apply_presentation_transitions",
    "resolve_background",
    "add_text_boxed",
    "generate_title_slide",
    "generate_section_divider",
    "generate_content_slide",
    "generate_quote_slide",
    "generate_card_grid",
    "generate_image_text",
    "generate_agenda",
    "generate_kpi_dashboard",
    "generate_comparison_table",
    "generate_kanban",
    "generate_brand_focus",
    "generate_timeline",
    "generate_chart_slide",
    "generate_swot_analysis",
]
