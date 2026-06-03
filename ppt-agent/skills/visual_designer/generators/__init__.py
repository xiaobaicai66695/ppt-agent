"""PPT slide generators - exports all generators."""
from .base import (
    PALETTES, new_presentation, set_slide_background, set_image_background,
    save_presentation, save_slide, resolve_background, get_background_path,
)
from .background_manager import (
    get_palette_for_background, BACKGROUND_PALETTE_MAP,
)
from .title_slide_generator import generate as generate_title_slide
from .section_divider_generator import generate as generate_section_divider
from .content_slide_generator import generate as generate_content_slide
from .stat_slide_generator import generate as generate_stat_slide
from .quote_slide_generator import generate as generate_quote_slide
from .card_grid_generator import generate as generate_card_grid
from .comparison_table_generator import generate as generate_comparison_table
from .image_hero_generator import generate as generate_image_hero
from .process_flow_generator import generate as generate_process_flow
from .two_column_generator import generate as generate_two_column
from .three_column_generator import generate as generate_three_column
from .summary_slide_generator import generate as generate_summary_slide
from .image_text_generator import generate as generate_image_text
from .example_detail_generator import generate as generate_example_detail
from .deep_dive_generator import generate as generate_deep_dive
from .agenda_generator import generate as generate_agenda
from .case_study_generator import generate as generate_case_study
from .kpi_dashboard_generator import generate as generate_kpi_dashboard
from .kanban_generator import generate as generate_kanban
from .brand_focus_generator import generate as generate_brand_focus
from .region_map_generator import generate as generate_region_map
from .timeline_generator import generate as generate_timeline
from .icon_grid_generator import generate as generate_icon_grid
from .chart_slide_generator import generate as generate_chart_slide
from .swot_analysis_generator import generate as generate_swot_analysis

__all__ = [
    "PALETTES",
    "new_presentation",
    "set_slide_background",
    "set_image_background",
    "save_presentation",
    "save_slide",
    "resolve_background",
    "get_background_path",
    "get_palette_for_background",
    "BACKGROUND_PALETTE_MAP",
    "generate_title_slide",
    "generate_section_divider",
    "generate_content_slide",
    "generate_stat_slide",
    "generate_quote_slide",
    "generate_card_grid",
    "generate_process_flow",
    "generate_two_column",
    "generate_three_column",
    "generate_summary_slide",
    "generate_image_text",
    "generate_example_detail",
    "generate_deep_dive",
    "generate_agenda",
    "generate_case_study",
    "generate_kpi_dashboard",
    "generate_comparison_table",
    "generate_image_hero",
    "generate_kanban",
    "generate_brand_focus",
    "generate_region_map",
    "generate_timeline",
    "generate_icon_grid",
    "generate_chart_slide",
    "generate_swot_analysis",
]
