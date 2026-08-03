"""Local visual asset helpers for visual_designer generators."""

from __future__ import annotations

import json
from functools import lru_cache
from pathlib import Path
from typing import Iterable

from pptx.util import Inches

from .base import add_rect, set_image_background


ASSET_ROOT = Path(__file__).resolve().parents[1] / "assets"
MANIFEST_PATH = ASSET_ROOT / "manifest.json"


@lru_cache(maxsize=1)
def load_manifest() -> list[dict]:
    if not MANIFEST_PATH.exists():
        return []
    data = json.loads(MANIFEST_PATH.read_text(encoding="utf-8"))
    return data.get("assets", [])


def asset_path(asset_id: str) -> str | None:
    for asset in load_manifest():
        if asset.get("id") == asset_id:
            path = ASSET_ROOT / asset.get("path", "")
            return str(path) if path.exists() else None
    return None


def find_asset(kind: str, tags: Iterable[str] = (), preferred_id: str = "") -> dict | None:
    assets = [a for a in load_manifest() if a.get("type") == kind]
    if preferred_id:
        for asset in assets:
            if asset.get("id") == preferred_id:
                return asset
    tag_set = {t for t in tags if t}
    if tag_set:
        ranked = []
        for asset in assets:
            score = len(tag_set.intersection(set(asset.get("tags", []))))
            score += len(tag_set.intersection(set(asset.get("recommended_templates", []))))
            if score > 0:
                ranked.append((score, asset.get("id", ""), asset))
        if ranked:
            ranked.sort(reverse=True)
            return ranked[0][2]
    return assets[0] if assets else None


def resolve_asset_background(background: str | None, role: str = "") -> str | None:
    if background and background.startswith("asset:"):
        return asset_path(background.split(":", 1)[1])
    if background:
        direct = asset_path(background)
        if direct:
            return direct
        return None
    asset = find_asset("background", tags=[role], preferred_id=default_background_for_role(role))
    if asset:
        return asset_path(asset.get("id", ""))
    return None


def default_background_for_role(role: str) -> str:
    return {
        "title": "editorial_blue",
        "section": "editorial_gold",
        "hero": "editorial_blue",
        "quote": "editorial_lavender",
        "summary": "editorial_sage",
    }.get(role, "")


def apply_asset_background(slide, background: str | None, palette: str, role: str, brightness: float = 0.96):
    path = resolve_asset_background(background, role)
    if path:
        return set_image_background(slide, path, brightness=brightness, palette=palette)
    return None


def icon_id_from_text(text: str, fallback: str = "primitive") -> str:
    text = (text or "").lower()
    keyword_map = [
        ("runtime", "runtime"), ("运行", "runtime"), ("可观测", "runtime"),
        ("timeline", "timeline"), ("轨迹", "timeline"), ("事件", "timeline"),
        ("tool", "tool"), ("工具", "tool"),
        ("llm", "llm"), ("模型", "llm"), ("agent", "llm"),
        ("file", "file"), ("文件", "file"), ("输出", "file"),
        ("report", "report"), ("报告", "report"),
        ("contract", "contract"), ("契约", "contract"), ("schema", "contract"),
        ("capacity", "capacity"), ("容量", "capacity"),
        ("layout", "layout"), ("布局", "layout"),
        ("split", "split"), ("拆", "split"), ("分页", "split"),
        ("align", "align"), ("对齐", "align"), ("居中", "align"),
        ("density", "density"), ("密度", "density"),
        ("chart", "chart"), ("图表", "chart"),
        ("kpi", "kpi"), ("指标", "kpi"),
        ("trend", "trend"), ("趋势", "trend"), ("增长", "trend"),
        ("source", "source"), ("来源", "source"),
        ("warning", "warning"), ("风险", "warning"), ("预算", "warning"),
        ("fix", "fix"), ("修复", "fix"),
        ("template", "template"), ("模板", "template"),
        ("background", "background"), ("背景", "background"),
        ("card", "card"), ("卡片", "card"),
        ("flow", "flow"), ("流程", "flow"),
        ("review", "review"), ("qa", "review"), ("审查", "review"),
    ]
    for keyword, icon_id in keyword_map:
        if keyword in text:
            return icon_id
    return fallback


def add_local_icon(
    slide,
    icon_id: str,
    left: float,
    top: float,
    size: float,
    palette: str = "ocean_soft",
    with_badge: bool = False,
    badge_color: str = "light_bg",
):
    path = asset_path(icon_id) or asset_path("primitive")
    if not path:
        return None
    if with_badge:
        add_rect(
            slide,
            left=left - size * 0.08,
            top=top - size * 0.08,
            width=size * 1.16,
            height=size * 1.16,
            fill_color=badge_color,
            palette=palette,
            line_color="divider",
            line_width=0.5,
        )
    return slide.shapes.add_picture(path, Inches(left), Inches(top), width=Inches(size), height=Inches(size))


def add_pattern_overlay(slide, pattern_id: str, left: float, top: float, width: float, height: float, opacity_backdrop: bool = True, palette: str = "ocean_soft"):
    path = asset_path(pattern_id)
    if not path:
        return None
    if opacity_backdrop:
        add_rect(slide, left=left, top=top, width=width, height=height, fill_color=(255, 255, 255, 170), palette=palette)
    return slide.shapes.add_picture(path, Inches(left), Inches(top), width=Inches(width), height=Inches(height))
