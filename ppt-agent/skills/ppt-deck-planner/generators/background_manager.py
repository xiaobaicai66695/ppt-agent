"""Compatibility layer for the removed local background catalog.

The current PPT pipeline plans real images through search tools and passes
downloaded files via local_path. Legacy background fields are tolerated but no
longer resolved from an offline bundle.
"""

from __future__ import annotations

from typing import Optional


BACKGROUND_ROOT = None


def load_background_manifest() -> dict:
    return {"version": 0, "themes": []}


def _manifest_theme_mapping() -> dict[str, dict]:
    return {}


def _manifest_palette_mapping() -> dict[str, str]:
    return {}


# Kept as exported compatibility constants for older generator imports.
THEME_MAPPING = _manifest_theme_mapping()
BACKGROUND_PALETTE_MAP: dict[str, str] = _manifest_palette_mapping()


def get_background_dir():
    return BACKGROUND_ROOT


def scan_backgrounds() -> list[dict]:
    return []


def validate_background_manifest(min_images_per_theme: int = 4) -> list[str]:
    _ = min_images_per_theme
    return []


def get_background(
    theme: Optional[str] = None,
    scenario: Optional[str] = None,
    random_select: bool = False,
) -> Optional[str]:
    _ = theme, scenario, random_select
    return None


def list_themes() -> list[dict]:
    return scan_backgrounds()


def get_palette_for_background(background_theme: str) -> str:
    return BACKGROUND_PALETTE_MAP.get(background_theme, "ocean_soft")
