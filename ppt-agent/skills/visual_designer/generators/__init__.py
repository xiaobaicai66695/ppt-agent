"""PPT slide generators - exports all generators."""
from .base import PALETTES, new_presentation, set_slide_background, save_presentation, save_slide
from .title_slide_generator import generate as generate_title_slide
from .section_divider_generator import generate as generate_section_divider
from .content_slide_generator import generate as generate_content_slide
from .stat_slide_generator import generate as generate_stat_slide
from .quote_slide_generator import generate as generate_quote_slide
from .card_grid_generator import generate as generate_card_grid
from .timeline_generator import generate as generate_timeline
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

__all__ = [
    "PALETTES",
    "new_presentation",
    "set_slide_background",
    "save_presentation",
    "save_slide",
    "generate_title_slide",
    "generate_section_divider",
    "generate_content_slide",
    "generate_stat_slide",
    "generate_quote_slide",
    "generate_card_grid",
    "generate_timeline",
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
]
