"""Chart slide generator using native python-pptx charts."""
from typing import Optional, List, Dict
from pptx import Presentation
from pptx.util import Inches, Pt, Emu
from pptx.enum.text import PP_ALIGN
from pptx.enum.chart import XL_CHART_TYPE, XL_LEGEND_POSITION
from pptx.chart.data import CategoryChartData
from pptx.dml.color import RGBColor

from .base import (
    add_source_line,
    new_presentation,
    PALETTES, add_text, add_rect, add_round_rect,
    set_slide_background,
    resolve_background, set_image_background,
)


def _get_color_rgb(palette: str, color_key: str) -> RGBColor:
    """Convert palette color key to RGBColor."""
    colors = PALETTES.get(palette, PALETTES["ocean_soft"])
    color_map = {
        "primary": colors.get("primary", "#2563EB"),
        "secondary": colors.get("secondary", "#64748B"),
        "accent": colors.get("accent", "#F59E0B"),
        "light_bg": colors.get("light_bg", "#F1F5F9"),
        "background": colors.get("background", "#FFFFFF"),
        "text": colors.get("text", "#1E293B"),
        "text_muted": colors.get("text_muted", "#64748B"),
        "divider": colors.get("divider", "#E2E8F0"),
    }
    hex_color = color_map.get(color_key, "#2563EB")
    if hex_color.startswith("#"):
        hex_color = hex_color[1:]
        r = int(hex_color[0:2], 16)
        g = int(hex_color[2:4], 16)
        b = int(hex_color[4:6], 16)
        return RGBColor(r, g, b)
    return RGBColor(37, 99, 235)


def _hex_to_rgb(hex_color: str) -> tuple:
    """Convert hex color to RGB tuple."""
    if hex_color.startswith("#"):
        hex_color = hex_color[1:]
    r = int(hex_color[0:2], 16)
    g = int(hex_color[2:4], 16)
    b = int(hex_color[4:6], 16)
    return (r, g, b)


def generate(
    prs: Optional[Presentation] = None,
    palette: str = "ocean_soft",
    source: str = "",
    kicker: str = "",
    title: str = "{图表标题}",
    subtitle: str = "",
    chart_type: str = "bar",
    data: Optional[Dict] = None,
    show_legend: bool = True,
    analysis: Optional[str] = None,
    background: str = None,
    glass_colors: dict = None,
) -> Presentation:
    """
    Generate a chart slide using native PowerPoint charts.

    Args:
        prs: Existing Presentation object. If None, creates a new one.
        palette: Color palette name.
        kicker: Small label above title.
        title: Slide title.
        subtitle: Optional subtitle.
        chart_type: "bar", "pie", "line", "doughnut", "stacked_bar".
        data: Dict with "labels" (list) and "datasets" (list of dicts with name/values).
        show_legend: Whether to show chart legend.
        analysis: Optional data analysis text to display beside the chart.

    Returns:
        The Presentation object.
    """
    if data is None:
        data = {
            "labels": ["Q1", "Q2", "Q3", "Q4"],
            "datasets": [
                {"name": "2025年", "values": [1200, 1500, 1800, 2200]},
                {"name": "2024年", "values": [900, 1100, 1300, 1600]}
            ]
        }

    if prs is None:
        prs = new_presentation(palette=palette)

    blank_layout = prs.slide_layouts[6]
    slide = prs.slides.add_slide(blank_layout)
    bg_path = resolve_background(background) if background else None
    if bg_path:
        colors = set_image_background(slide, bg_path, brightness=0.95, palette=palette)
    else:
        set_slide_background(slide, palette)
        colors = PALETTES.get(palette, PALETTES["ocean_soft"])

    # Kicker
    y_offset = 0.4
    if kicker:
        add_text(
            slide,
            text=kicker,
            left=0.6, top=0.15, width=12.0, height=0.3,
            font_size=12, bold=False,
            color="secondary", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_offset = 0.35

    # Title with left accent bar
    add_rect(
        slide,
        left=0.5, top=y_offset + 0.05, width=0.08, height=0.55,
        fill_color="primary", palette=palette,
    )
    add_text(
        slide,
        text=title,
        left=0.7, top=y_offset, width=11.5, height=0.65,
        font_size=30, bold=True,
        color="text", alignment="left",
        palette=palette,
        colors=colors,
    )
    y_offset += 0.7

    # Subtitle
    if subtitle:
        add_text(
            slide,
            text=subtitle,
            left=0.7, top=y_offset, width=11.5, height=0.35,
            font_size=14, bold=False,
            color="text_muted", alignment="left",
            palette=palette,
            colors=colors,
        )
        y_offset += 0.4

    # Determine chart layout based on whether analysis is provided
    has_analysis = analysis and len(analysis.strip()) > 0

    if has_analysis:
        # Chart on left (60%), analysis panel on right (35%)
        chart_left = 0.7
        chart_top = y_offset + 0.15
        chart_width = 7.5
        chart_height = 4.6

        # Analysis panel on the right
        panel_left = 8.5
        panel_top = y_offset + 0.15
        panel_width = 4.3
        panel_height = 4.6

        # Panel background
        add_round_rect(
            slide,
            left=panel_left, top=panel_top, width=panel_width, height=panel_height,
            fill_color="light_bg", palette=palette,
        )

        # Panel header
        add_text(
            slide,
            text="数据分析",
            left=panel_left + 0.2, top=panel_top + 0.15, width=panel_width - 0.4, height=0.4,
            font_size=14, bold=True,
            color="primary", alignment="left",
            palette=palette,
            colors=colors,
        )

        # Divider line under header
        add_rect(
            slide,
            left=panel_left + 0.2, top=panel_top + 0.55, width=panel_width - 0.4, height=0.02,
            fill_color="divider", palette=palette,
        )

        # Analysis content
        add_text(
            slide,
            text=analysis,
            left=panel_left + 0.2, top=panel_top + 0.7, width=panel_width - 0.4, height=panel_height - 0.9,
            font_size=11, bold=False,
            color="text", alignment="left",
            palette=palette,
            colors=colors,
        )
    else:
        # Full width chart
        chart_left = 0.7
        chart_top = y_offset + 0.15
        chart_width = 12.0
        chart_height = 5.0

    labels = data.get("labels", [])
    datasets = data.get("datasets", [])

    # Create native PowerPoint chart
    chart = _create_chart(
        slide, chart_type, palette,
        chart_left, chart_top, chart_width, chart_height,
        labels, datasets
    )

    # Configure legend
    if chart and show_legend:
        chart.has_legend = True
        chart.legend.position = XL_LEGEND_POSITION.RIGHT
        chart.legend.include_in_layout = False

        # Style legend text
        if chart.legend.font:
            chart.legend.font.size = Pt(10)
            chart.legend.font.color.rgb = _get_color_rgb(palette, "text")

    add_source_line(slide, source, palette)
    return prs


def _create_chart(slide, chart_type: str, palette: str,
                  left: float, top: float, width: float, height: float,
                  labels: List, datasets: List):
    """Create a native PowerPoint chart based on chart type."""
    chart_data = CategoryChartData()
    chart_data.categories = labels

    # Add series
    for ds in datasets:
        chart_data.add_series(ds.get("name", "数据"), tuple(ds.get("values", [])))

    # Determine chart type
    chart_types = {
        "bar": XL_CHART_TYPE.COLUMN_CLUSTERED,
        "stacked_bar": XL_CHART_TYPE.COLUMN_STACKED,
        "pie": XL_CHART_TYPE.PIE,
        "line": XL_CHART_TYPE.LINE,
        "doughnut": XL_CHART_TYPE.DOUGHNUT,
    }

    xl_chart_type = chart_types.get(chart_type, XL_CHART_TYPE.COLUMN_CLUSTERED)

    # Add chart to slide
    chart = slide.shapes.add_chart(
        xl_chart_type,
        Inches(left), Inches(top), Inches(width), Inches(height),
        chart_data
    ).chart

    # Apply palette colors (simplified to avoid API issues)
    _apply_palette_colors(chart, palette, chart_type, len(datasets))

    return chart


def _apply_palette_colors(chart, palette: str, chart_type: str, num_series: int):
    """Apply palette colors to chart series."""
    palette_data = PALETTES.get(palette, PALETTES["ocean_soft"])
    palette_colors = [
        palette_data.get("primary", "#2563EB"),
        palette_data.get("secondary", "#64748B"),
        palette_data.get("accent", "#F59E0B"),
        palette_data.get("primary", "#2563EB"),
    ]

    plot = chart.plots[0]

    try:
        if chart_type in ["pie", "doughnut"]:
            # For pie/doughnut: color points individually
            for idx, point in enumerate(plot.series[0].points):
                color_idx = idx % len(palette_colors)
                color_hex = palette_colors[color_idx]
                r, g, b = _hex_to_rgb(color_hex)
                point.format.fill.solid()
                point.format.fill.fore_color.rgb = RGBColor(r, g, b)
        elif chart_type == "line":
            # Line charts - apply colors to lines and markers
            for idx, series in enumerate(plot.series):
                color_idx = idx % len(palette_colors)
                color_hex = palette_colors[color_idx]
                r, g, b = _hex_to_rgb(color_hex)

                # Set line color
                series.format.line.color.rgb = RGBColor(r, g, b)
                series.format.line.width = Pt(2)

                # Set marker style
                series.marker.style = 1  # XL_MARKER_STYLE.CIRCLE
                series.marker.size = 6
                series.marker.format.fill.solid()
                series.marker.format.fill.fore_color.rgb = RGBColor(r, g, b)
                series.marker.format.line.color.rgb = RGBColor(r, g, b)
        else:
            # For bar/stacked_bar: set fill color properly
            for idx, series in enumerate(plot.series):
                color_idx = idx % len(palette_colors)
                color_hex = palette_colors[color_idx]
                r, g, b = _hex_to_rgb(color_hex)

                # Apply fill color to the entire series
                series.format.fill.solid()
                series.format.fill.fore_color.rgb = RGBColor(r, g, b)

                # Also try to set the actual data points
                try:
                    for point_idx, point in enumerate(series.points):
                        point.format.fill.solid()
                        point.format.fill.fore_color.rgb = RGBColor(r, g, b)
                except Exception:
                    pass

        # Configure data labels - hide when too many to avoid clutter
        # For charts with many data points, labels cause visual clutter
        num_data_points = len(labels) * len(datasets)
        show_data_labels = num_data_points <= 12  # Only show if <= 12 data points

        if show_data_labels:
            plot.has_data_labels = True
            if plot.data_labels:
                plot.data_labels.show_value = True
                plot.data_labels.show_category_name = False
                plot.data_labels.show_series_name = False
                plot.data_labels.font.size = Pt(9)
                text_rgb = _get_color_rgb(palette, "text")
                plot.data_labels.font.color.rgb = text_rgb
        else:
            plot.has_data_labels = False

        # Configure axis for better readability
        _configure_axis(chart, chart_type, len(labels), labels)

    except Exception:
        # Silently ignore color formatting errors
        pass


def _configure_axis(chart, chart_type: str, num_labels: int = 0, labels: list = None):
    """Configure chart axis for better readability - optimize Y-axis ticks and X-axis labels."""
    try:
        value_axis = chart.value_axis
        category_axis = chart.category_axis

        # Configure category axis (X-axis) - handle many labels by reducing density
        if labels and len(labels) > 6:
            # Show only every Nth label to avoid crowding
            step = max(2, len(labels) // 5)
            try:
                category_axis.tick_label_spacing = step
            except Exception:
                pass
        else:
            try:
                category_axis.tick_label_spacing = 1
            except Exception:
                pass

        # Configure value axis (Y-axis) - force clean tick count
        all_values = []
        for series in chart.series:
            all_values.extend([v for v in series.values if v is not None])

        if all_values:
            min_val = min(all_values)
            max_val = max(all_values)
            data_range = max_val - min_val

            # Target 5-6 ticks for clean appearance
            target_ticks = 5

            if data_range > 0:
                # Calculate ideal unit
                ideal_unit = data_range / target_ticks
                nice_unit = _calculate_nice_unit(ideal_unit)

                # Calculate axis bounds with padding (15% on each side)
                padding = max(data_range * 0.12, nice_unit * 0.5)
                y_min = max(0, min_val - padding)
                y_max = max_val + padding

                # Snap to nice units
                axis_min = _floor_to_nice(y_min, nice_unit)
                axis_max = _ceil_to_nice(y_max, nice_unit)

                # Count ticks and adjust if needed
                tick_count = int((axis_max - axis_min) / nice_unit) + 1
                while tick_count > 6 and nice_unit < axis_max:
                    # Increase unit to reduce tick count
                    nice_unit = nice_unit * 2
                    axis_min = _floor_to_nice(y_min, nice_unit)
                    axis_max = _ceil_to_nice(y_max, nice_unit)
                    tick_count = int((axis_max - axis_min) / nice_unit) + 1
                    if nice_unit >= axis_max:
                        break

                # Set axis explicitly
                value_axis.minimum_scale = axis_min
                value_axis.maximum_scale = axis_max
                value_axis.major_unit = nice_unit

            else:
                # All values are the same
                if max_val == 0:
                    value_axis.minimum_scale = 0
                    value_axis.maximum_scale = 10
                    value_axis.major_unit = 2
                else:
                    value_axis.minimum_scale = 0
                    value_axis.maximum_scale = max_val * 1.5
                    value_axis.major_unit = max_val * 0.3

        # Force tick mark style - only show major ticks
        try:
            value_axis.major_tick_mark = 1  # XL_TICK_MARK.OUTSIDE
            value_axis.minor_tick_mark = 0  # XL_TICK_MARK.NONE
            category_axis.major_tick_mark = 1
            category_axis.minor_tick_mark = 0
        except Exception:
            pass

        # Style the gridlines to be subtle
        try:
            value_axis.major_gridlines.format.line.width = 0.5
            value_axis.major_gridlines.format.line.color.rgb = RGBColor(220, 220, 220)
        except Exception:
            pass

        # Format tick labels - smaller and lighter color
        try:
            value_axis.tick_labels.font.size = Pt(9)
            value_axis.tick_labels.font.color.rgb = RGBColor(100, 100, 100)
            category_axis.tick_labels.font.size = Pt(8)
            category_axis.tick_labels.font.color.rgb = RGBColor(100, 100, 100)
        except Exception:
            pass

    except Exception:
        pass


def _calculate_nice_unit(raw_unit: float) -> float:
    """Calculate a nice unit value for axis ticks."""
    if raw_unit <= 0:
        return 1

    # Find the magnitude (power of 10)
    if raw_unit >= 1:
        magnitude = 10 ** (len(str(int(raw_unit))) - 1)
    else:
        magnitude = 0.1

    normalized = raw_unit / magnitude

    # Round to nearest nice number: 1, 2, 5, 10
    if normalized <= 1:
        result = 1 * magnitude
    elif normalized <= 2:
        result = 2 * magnitude
    elif normalized <= 5:
        result = 5 * magnitude
    else:
        result = 10 * magnitude

    return max(result, 0.001)


def _floor_to_nice(value: float, unit: float) -> float:
    """Floor value to nearest nice unit."""
    if unit == 0:
        return value
    return (int(value / unit)) * unit


def _ceil_to_nice(value: float, unit: float) -> float:
    """Ceiling value to nearest nice unit."""
    if unit == 0:
        return value
    return (int((value + unit - 0.001) / unit)) * unit


def _draw_bar_chart(slide, palette, left, top, width, height, labels, datasets):
    """Legacy function - now uses native charts via _create_chart."""
    pass


def _draw_pie_chart(slide, palette, left, top, width, height, labels, dataset):
    """Legacy function - now uses native charts via _create_chart."""
    pass


def _draw_line_chart(slide, palette, left, top, width, height, labels, datasets):
    """Legacy function - now uses native charts via _create_chart."""
    pass


def _draw_doughnut_chart(slide, palette, left, top, width, height, labels, dataset):
    """Legacy function - now uses native charts via _create_chart."""
    pass
