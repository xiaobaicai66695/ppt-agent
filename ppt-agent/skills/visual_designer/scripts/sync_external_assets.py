#!/usr/bin/env python3
"""Synchronize the curated offline visual asset library.

The generator runtime never downloads files. Run this script only when the
curated source list changes, then commit the normalized files and manifests.
"""

from __future__ import annotations

import argparse
from dataclasses import dataclass
from datetime import date
from io import BytesIO
import json
from pathlib import Path
import shutil
import tempfile
import time
from typing import Iterable
from urllib.error import URLError
from urllib.request import Request, urlopen

from PIL import Image, ImageOps


ROOT = Path(__file__).resolve().parents[1]
ASSET_ROOT = ROOT / "assets"
BACKGROUND_ROOT = ROOT / "background_templates"
BING_DISCOVERY_URL = (
    "https://cn.bing.com/images/search?q=%e8%83%8c%e6%99%af+%e5%9b%be%e7%89%87+%e5%a4%a7%e5%85%a8"
    "&qpvt=%e8%83%8c%e6%99%af%e5%9b%be%e7%89%87%e5%a4%a7%e5%85%a8&form=IGRE&first=1"
)
ICON_STYLE = "fluency-systems-regular"
ICON_COLOR = "344054"
ICON_URL = "https://img.icons8.com/{style}/512/{color}/{slug}.png"
PICSUM_URL = "https://picsum.photos/id/{photo_id}/1920/1080"


@dataclass(frozen=True)
class IconSpec:
    asset_id: str
    slug: str
    keywords: tuple[str, ...]
    tags: tuple[str, ...] = ()
    priority: int = 100


@dataclass(frozen=True)
class PhotoSpec:
    asset_id: str
    photo_id: int
    author: str
    source_url: str
    tags: tuple[str, ...]
    recommended_templates: tuple[str, ...]


@dataclass(frozen=True)
class ContentPhotoSpec:
    asset_id: str
    category: str
    photo_id: int
    author: str
    source_url: str
    keywords: tuple[str, ...]
    tags: tuple[str, ...] = ()
    priority: int = 100


@dataclass(frozen=True)
class PatternSpec:
    asset_id: str
    filename: str
    source_url: str
    tags: tuple[str, ...]


@dataclass(frozen=True)
class ThemePhotoSpec:
    theme: str
    filename: str
    photo_id: int
    author: str
    source_url: str


def _icon(asset_id: str, slug: str, keywords: str, *tags: str, priority: int = 100) -> IconSpec:
    return IconSpec(
        asset_id=asset_id,
        slug=slug,
        keywords=tuple(part.strip() for part in keywords.split("|") if part.strip()),
        tags=tuple(tags),
        priority=priority,
    )


ICONS = [
    _icon("overview", "overview-pages-1", "overview|概览|总览|介绍|目录|首页", "structure"),
    _icon("section", "bookmark", "section|chapter|章节|篇章|部分|阶段", "structure", priority=120),
    _icon("thanks", "applause", "thanks|thank you|感谢|致谢|谢谢|聆听", "closing", priority=130),
    _icon("contact", "contact-card", "contact|email|phone|联系|联系方式|邮箱|电话", "closing", priority=125),
    _icon("question", "help", "question|q&a|faq|问题|提问|问答|答疑", "closing"),
    _icon("map", "map", "map|region|geography|地图|区域|地域|地理|版图", "geography", priority=125),
    _icon("location", "marker", "location|address|place|位置|地点|地址|坐标", "geography"),
    _icon("forest", "forest", "forest|woodland|森林|林业|树林", "nature"),
    _icon("ecology", "leaf", "ecology|green|sustainable|生态|绿色|可持续|环保", "nature", priority=120),
    _icon("protection", "shield", "protection|protect|guard|保护|保障|守护|防护", "nature", "security"),
    _icon("challenge", "high-priority", "challenge|difficulty|issue|挑战|难点|困境|痛点", "risk"),
    _icon("policy", "rules", "policy|regulation|guideline|政策|规则|制度|规划", "civic", priority=125),
    _icon("government", "parliament", "government|civic|政务|政府|机关|行政", "civic"),
    _icon("meeting", "conference-call", "meeting|conference|forum|会议|大会|论坛|两会", "civic", "business", priority=125),
    _icon("people", "user-group-man-man", "people|population|team|audience|人群|人口|团队|受众", "people"),
    _icon("project", "project", "project|initiative|项目|专项|工程", "business"),
    _icon("task", "task", "task|todo|action|任务|行动|待办", "business"),
    _icon("goal", "goal", "goal|target|objective|目标|愿景|方向", "business"),
    _icon("budget", "budget", "budget|cost|expense|预算|成本|费用", "finance"),
    _icon("product", "product", "product|service|solution|产品|服务|方案", "business"),
    _icon("market", "commercial", "market|business|commerce|市场|商业|行业", "business"),
    _icon("customer", "customer-support", "customer|client|user|客户|顾客|用户", "business"),
    _icon("research", "microscope", "research|study|experiment|研究|调研|实验|探索", "science"),
    _icon("architecture", "mind-map", "architecture|framework|system|架构|框架|系统", "technology"),
    _icon("database", "database", "database|data store|storage|数据库|数据存储|存储", "technology"),
    _icon("security", "security-checked", "security|privacy|compliance|安全|隐私|合规", "technology"),
    _icon("mountain", "mountain", "mountain|peak|山地|山脉|雪山|高峰", "nature"),
    _icon("water", "water", "water|ocean|river|水资源|海洋|河流|湖泊", "nature"),
    _icon("climate", "temperature", "climate|weather|temperature|气候|天气|温度", "nature"),
    _icon("party", "star", "party|党建|党委|党组织|党员", "civic"),
    _icon("flag", "flag", "flag|milestone|国旗|旗帜|里程碑", "civic"),
    _icon("law", "law", "law|legal|legislation|法律|法治|立法", "civic"),
    _icon("runtime", "activity-history", "runtime|observability|status|运行|可观测|状态", "agent"),
    _icon("timeline", "timeline", "timeline|history|event|时间线|轨迹|事件|历史", "structure"),
    _icon("tool", "toolbox", "tool|execution|operation|工具|执行|操作", "agent"),
    _icon("llm", "artificial-intelligence", "llm|model|agent|ai|大模型|模型|智能体|人工智能", "agent"),
    _icon("file", "file", "file|artifact|output|文件|产物|输出", "document"),
    _icon("report", "business-report", "report|briefing|报告|汇报|简报", "document"),
    _icon("contract", "agreement", "contract|schema|agreement|契约|协议|结构约束", "document"),
    _icon("capacity", "resize", "capacity|limit|scale|容量|上限|规模", "layout"),
    _icon("layout", "dashboard-layout", "layout|composition|布局|版式|排版", "layout"),
    _icon("split", "split", "split|pagination|overflow|拆分|分页|溢出", "layout"),
    _icon("align", "align-center", "align|center|balance|对齐|居中|平衡", "layout"),
    _icon("density", "content", "density|content|text|密度|内容|文字", "layout"),
    _icon("chart", "combo-chart", "chart|graph|data|图表|数据图|可视化", "data"),
    _icon("kpi", "performance", "kpi|metric|indicator|指标|绩效|度量", "data"),
    _icon("trend", "growth", "trend|growth|increase|趋势|增长|提升", "data"),
    _icon("source", "source-code", "source|citation|reference|来源|引用|参考", "document"),
    _icon("warning", "error", "warning|risk|danger|警告|风险|危险|异常", "risk"),
    _icon("fix", "maintenance", "fix|repair|improve|修复|整改|优化", "agent"),
    _icon("template", "template", "template|preset|模板|预设|母版", "layout"),
    _icon("background", "image", "background|backdrop|背景|底图", "layout"),
    _icon("card", "stack", "card|module|grid|卡片|模块|网格", "layout"),
    _icon("flow", "workflow", "flow|process|workflow|流程|工作流|步骤", "structure"),
    _icon("review", "inspection", "review|inspect|qa|审查|评审|检查|质检", "agent"),
    _icon("calendar", "calendar", "calendar|date|schedule|日历|日期|计划|排期", "time"),
    _icon("clock", "clock", "clock|time|duration|时间|时长|周期", "time"),
    _icon("idea", "idea", "idea|insight|inspiration|想法|洞察|灵感|观点", "creative"),
    _icon("innovation", "innovation", "innovation|creative|创新|创意|突破", "creative"),
    _icon("education", "school", "education|school|training|教育|学校|培训|学习", "domain"),
    _icon("technology", "electronics", "technology|digital|tech|科技|技术|数字化", "domain"),
    _icon("cloud", "cloud", "cloud|cloud computing|云|云计算", "technology"),
    _icon("network", "network", "network|connection|internet|网络|连接|互联网", "technology"),
    _icon("finance", "bank", "finance|bank|investment|金融|银行|投资", "domain"),
    _icon("healthcare", "health-book", "health|medical|healthcare|健康|医疗|医药", "domain"),
    _icon("energy", "lightning-bolt", "energy|electricity|power|能源|电力|能耗", "domain"),
    _icon("manufacturing", "factory", "manufacturing|factory|industry|制造业|制造|工厂|工业", "domain", priority=110),
    _icon("agriculture", "farm", "agriculture|farm|rural|农业|农场|乡村", "domain"),
    _icon("transport", "train", "transport|traffic|logistics|交通|运输|物流", "domain"),
    _icon("city", "city", "city|urban|building|城市|城区|建筑", "geography"),
    _icon("globe", "globe", "globe|world|international|全球|世界|国际", "geography"),
    _icon("award", "prize", "award|honor|achievement|奖项|荣誉|成就", "people"),
    _icon("strategy", "strategy-board", "strategy|roadmap|strategy plan|战略|路线图|策略", "business"),
    _icon("teamwork", "collaboration", "teamwork|cooperation|collaboration|协作|合作|共创", "people"),
    _icon("communication", "chat", "communication|conversation|message|沟通|交流|对话", "people"),
    _icon("check", "checkmark", "check|complete|success|完成|成功|确认|达成", "status"),
    _icon("summary", "summary-list", "summary|conclusion|takeaway|总结|结论|要点|回顾", "structure"),
    _icon("quote", "quote-left", "quote|quotation|saying|引用语|金句|引言|语录", "structure"),
    _icon("presentation", "presentation", "presentation|slide|ppt|演示|幻灯片|演讲", "structure"),
]


ASSET_BACKGROUNDS = [
    PhotoSpec("editorial_workspace", 60, "Vadim Sherbakov", "https://unsplash.com/photos/Hi9GSwWkCJk", ("title", "hero", "business", "workspace"), ("title_slide", "image_hero")),
    PhotoSpec("editorial_sky", 53, "J Duclos", "https://unsplash.com/photos/6qORI5j_6n8", ("title", "quote", "soft", "sky"), ("title_slide", "quote_slide")),
    PhotoSpec("editorial_city", 43, "Oleg Chursin", "https://unsplash.com/photos/IoCWq07GaG4", ("section", "hero", "city", "architecture"), ("section_divider", "image_hero")),
    PhotoSpec("editorial_creative", 58, "Tony Naccarato", "https://unsplash.com/photos/-kEr-QltARg", ("title", "section", "creative", "color"), ("title_slide", "section_divider")),
    PhotoSpec("editorial_heritage", 78, "Paul Evans", "https://unsplash.com/photos/CtkDsu4w-Rs", ("quote", "section", "heritage", "texture"), ("quote_slide", "section_divider")),
    PhotoSpec("editorial_nature", 95, "Kundan Ramisetti", "https://unsplash.com/photos/87TJNWkepvI", ("summary", "quote", "nature", "quiet"), ("summary_slide", "quote_slide")),
]


CONTENT_PHOTOS = [
    ContentPhotoSpec(
        "photo_technology_device", "technology", 0, "Alejandro Escamilla",
        "https://unsplash.com/photos/yC-Yzbqy7PY",
        ("technology", "digital", "device", "computer", "科技", "技术", "数字化", "电脑", "设备", "系统", "研发"),
        ("technology", "device", "workspace"), 118,
    ),
    ContentPhotoSpec(
        "photo_technology_workspace", "technology", 60, "Vadim Sherbakov",
        "https://unsplash.com/photos/Hi9GSwWkCJk",
        ("data", "platform", "terminal", "software", "数据", "平台", "终端", "软件", "人工智能", "大模型", "ai"),
        ("technology", "data", "workspace"), 116,
    ),
    ContentPhotoSpec(
        "photo_business_work", "business", 1, "Alejandro Escamilla",
        "https://unsplash.com/photos/LNRyGwIJr5c",
        ("business", "office", "team", "work", "商务", "办公", "企业", "团队", "协作", "工作"),
        ("business", "people", "workspace"), 115,
    ),
    ContentPhotoSpec(
        "photo_business_desk", "business", 180, "Galymzhan Abdugalimov",
        "https://unsplash.com/photos/ICW6QYOcdlg",
        ("plan", "report", "strategy", "operation", "方案", "报告", "规划", "战略", "运营", "总结"),
        ("business", "planning", "workspace"), 108,
    ),
    ContentPhotoSpec(
        "photo_education_design", "education", 20, "Aleks Dorohovich",
        "https://unsplash.com/photos/nJdwUHmaY8A",
        ("education", "learning", "training", "design", "教育", "学习", "培训", "课程", "设计", "创意"),
        ("education", "creative", "workspace"), 114,
    ),
    ContentPhotoSpec(
        "photo_education_book", "education", 24, "Alejandro Escamilla",
        "https://unsplash.com/photos/cZhUxIQjILg",
        ("book", "reading", "knowledge", "research", "书籍", "阅读", "知识", "研究", "文化", "文献"),
        ("education", "book", "knowledge"), 112,
    ),
    ContentPhotoSpec(
        "photo_city_architecture", "city", 43, "Oleg Chursin",
        "https://unsplash.com/photos/IoCWq07GaG4",
        ("city", "architecture", "region", "finance", "government", "城市", "建筑", "区域", "金融", "政务", "政府"),
        ("city", "architecture", "civic"), 110,
    ),
    ContentPhotoSpec(
        "photo_nature_outdoor", "nature", 54, "Nicholas Swanson",
        "https://unsplash.com/photos/d19by2PLaPc",
        ("outdoor", "travel", "mountain", "tourism", "户外", "旅行", "山地", "旅游", "风景"),
        ("nature", "outdoor", "travel"), 107,
    ),
    ContentPhotoSpec(
        "photo_nature_ecology", "nature", 95, "Kundan Ramisetti",
        "https://unsplash.com/photos/87TJNWkepvI",
        ("nature", "ecology", "environment", "forest", "自然", "生态", "环保", "森林", "绿色", "可持续"),
        ("nature", "ecology", "environment"), 119,
    ),
    ContentPhotoSpec(
        "photo_sports_baseball", "sports", 73, "Jon Eckert",
        "https://unsplash.com/photos/umLpP7uCZs0",
        ("sports", "fitness", "competition", "baseball", "体育", "运动", "健身", "竞赛", "棒球"),
        ("sports", "people", "action"), 110,
    ),
    ContentPhotoSpec(
        "photo_transport_bicycle", "transport", 76, "Alexander Shustov",
        "https://unsplash.com/photos/OxzhYtL-00Y",
        ("transport", "mobility", "bicycle", "traffic", "交通", "出行", "自行车", "运输", "物流"),
        ("transport", "mobility", "city"), 109,
    ),
    ContentPhotoSpec(
        "photo_culture_heritage", "culture", 78, "Paul Evans",
        "https://unsplash.com/photos/CtkDsu4w-Rs",
        ("culture", "heritage", "history", "tradition", "文化", "遗产", "历史", "传统", "建筑"),
        ("culture", "heritage", "history"), 113,
    ),
    ContentPhotoSpec(
        "photo_creative_camera", "creative", 91, "Jennifer Trovato",
        "https://unsplash.com/photos/baRYCsjO6z4",
        ("creative", "media", "photo", "camera", "创意", "媒体", "摄影", "相机", "影像"),
        ("creative", "media", "photography"), 111,
    ),
    ContentPhotoSpec(
        "photo_agriculture_field", "agriculture", 107, "Lukas Schweizer",
        "https://unsplash.com/photos/9VWOr22LhVI",
        ("agriculture", "farm", "crop", "rural", "农业", "农田", "粮食", "乡村", "种植"),
        ("agriculture", "nature", "rural"), 117,
    ),
]


PATTERNS = [
    PatternSpec("pattern_cubes", "cubes.png", "https://www.transparenttextures.com/patterns/cubes.png", ("geometry", "subtle", "business")),
    PatternSpec("pattern_diagonal", "diagmonds-light.png", "https://www.transparenttextures.com/patterns/diagmonds-light.png", ("diagonal", "subtle", "editorial")),
    PatternSpec("pattern_grid", "grid-me.png", "https://www.transparenttextures.com/patterns/grid-me.png", ("grid", "subtle", "technology")),
    PatternSpec("pattern_dots", "subtle-dots.png", "https://www.transparenttextures.com/patterns/subtle-dots.png", ("dots", "subtle", "clean")),
]


THEME_ADDITIONS = [
    ThemePhotoSpec("artistic", "2.jpg", 56, "Sebastian Muller", "https://unsplash.com/photos/VLdaxYyXJvw"),
    ThemePhotoSpec("artistic", "3.jpg", 58, "Tony Naccarato", "https://unsplash.com/photos/-kEr-QltARg"),
    ThemePhotoSpec("artistic", "4.jpg", 82, "Rula Sibai", "https://unsplash.com/photos/-vq7mi4oF0s"),
    ThemePhotoSpec("minimalist_blue", "2.jpg", 43, "Oleg Chursin", "https://unsplash.com/photos/IoCWq07GaG4"),
    ThemePhotoSpec("minimalist_blue", "3.jpg", 53, "J Duclos", "https://unsplash.com/photos/6qORI5j_6n8"),
    ThemePhotoSpec("minimalist_blue", "4.jpg", 74, "Isaak Dury", "https://unsplash.com/photos/YhZbnxqtooM"),
    ThemePhotoSpec("vintage_chinese", "2.jpg", 24, "Alejandro Escamilla", "https://unsplash.com/photos/cZhUxIQjILg"),
    ThemePhotoSpec("vintage_chinese", "3.jpg", 76, "Alexander Shustov", "https://unsplash.com/photos/OxzhYtL-00Y"),
    ThemePhotoSpec("vintage_chinese", "4.jpg", 78, "Paul Evans", "https://unsplash.com/photos/CtkDsu4w-Rs"),
    ThemePhotoSpec("snowy_mountain", "4.jpg", 29, "Go Wild", "https://unsplash.com/photos/V0yAek6BgGk"),
]


THEMES = {
    "party_government": {
        "name_cn": "党政办公",
        "scenarios": ["党建", "政府", "政务", "机关", "党委", "党支部", "红色"],
        "priority": 10,
        "recommended_palette": "government_red",
        "existing": ["1.jpg", "2.jpg", "3.jpg", "4.jpg", "5.jpg"],
    },
    "ink_wash_mountain": {
        "name_cn": "水墨山水",
        "scenarios": ["水墨", "山水", "艺术", "自然"],
        "priority": 8,
        "recommended_palette": "sage_calm",
        "existing": ["1.jpg", "2.jpg", "3.jpg", "4.jpg"],
    },
    "vintage_chinese": {
        "name_cn": "复古中国风",
        "scenarios": ["中国风", "传统", "文化", "国风", "古风", "复古", "文艺"],
        "priority": 7,
        "recommended_palette": "warm_terracotta",
        "existing": ["background.jpg"],
    },
    "artistic": {
        "name_cn": "艺术涂鸦",
        "scenarios": ["艺术", "创意", "涂鸦", "个性", "时尚", "现代艺术"],
        "priority": 6,
        "recommended_palette": "berry_cream",
        "existing": ["background.jpg"],
    },
    "minimalist_blue": {
        "name_cn": "简约蓝白",
        "scenarios": ["商务", "企业", "科技", "现代", "简约", "专业", "会议", "方案", "产品"],
        "priority": 5,
        "recommended_palette": "ocean_soft",
        "existing": ["background.jpg"],
    },
    "snowy_mountain": {
        "name_cn": "雪山风景",
        "scenarios": ["自然", "风景", "户外", "雪山", "山川", "旅行", "环保"],
        "priority": 4,
        "recommended_palette": "charcoal_light",
        "existing": ["1.jpg", "2.jpg", "3.jpg"],
    },
}


SOURCES = [
    {
        "id": "icons8",
        "homepage": "https://icons8.com/icons",
        "license": "Icons8 License",
        "terms_url": "https://icons8.com/license",
        "attribution": "Icons by Icons8",
    },
    {
        "id": "unsplash",
        "homepage": "https://unsplash.com",
        "license": "Unsplash License",
        "terms_url": "https://unsplash.com/license",
        "attribution": "Photo author and source URL are recorded per asset",
    },
    {
        "id": "transparent-textures",
        "homepage": "https://www.transparenttextures.com",
        "license": "CC BY 3.0",
        "terms_url": "https://www.transparenttextures.com",
        "attribution": "Transparent Textures",
    },
]


def _fetch(url: str, attempts: int = 4) -> tuple[bytes, dict[str, str]]:
    last_error: Exception | None = None
    for attempt in range(attempts):
        request = Request(url, headers={
            "User-Agent": "Mozilla/5.0 ppt-agent-asset-sync/2.0",
            "Connection": "close",
        })
        try:
            with urlopen(request, timeout=45) as response:
                data = response.read()
                headers = {key.lower(): value for key, value in response.headers.items()}
            if len(data) < 64:
                raise ValueError(f"remote response is too small: {url}")
            return data, headers
        except (OSError, URLError, ValueError) as exc:
            last_error = exc
            if attempt + 1 < attempts:
                time.sleep(1.5 * (attempt + 1))
    raise RuntimeError(f"download failed after {attempts} attempts: {url}") from last_error


def _open_image(data: bytes, asset_id: str) -> Image.Image:
    try:
        image = Image.open(BytesIO(data))
        image.load()
        return image
    except Exception as exc:
        raise ValueError(f"invalid image for {asset_id}: {exc}") from exc


def _save_icon(data: bytes, target: Path, asset_id: str, headers: dict[str, str]) -> None:
    if headers.get("not-found-platform", "").lower() == "true":
        raise ValueError(f"Icons8 slug is unavailable for {asset_id}")
    image = _open_image(data, asset_id).convert("RGBA")
    canvas = Image.new("RGBA", (512, 512), (0, 0, 0, 0))
    image.thumbnail((456, 456), Image.Resampling.LANCZOS)
    canvas.alpha_composite(image, ((512 - image.width) // 2, (512 - image.height) // 2))
    target.parent.mkdir(parents=True, exist_ok=True)
    canvas.save(target, "PNG", optimize=True)


def _save_photo(data: bytes, target: Path, asset_id: str) -> None:
    image = _open_image(data, asset_id).convert("RGB")
    normalized = ImageOps.fit(image, (1920, 1080), method=Image.Resampling.LANCZOS)
    target.parent.mkdir(parents=True, exist_ok=True)
    normalized.save(target, "JPEG", quality=89, optimize=True, progressive=True)


def _save_pattern(data: bytes, target: Path, asset_id: str) -> None:
    tile = _open_image(data, asset_id).convert("RGBA")
    canvas = Image.new("RGBA", (1920, 1080), (255, 255, 255, 0))
    for top in range(0, canvas.height, tile.height):
        for left in range(0, canvas.width, tile.width):
            canvas.alpha_composite(tile, (left, top))
    target.parent.mkdir(parents=True, exist_ok=True)
    canvas.save(target, "PNG", optimize=True)


def _dimensions(path: Path) -> list[int]:
    with Image.open(path) as image:
        return [image.width, image.height]


def _asset_entry(
    *,
    asset_id: str,
    kind: str,
    path: Path,
    tags: Iterable[str],
    keywords: Iterable[str],
    recommended_templates: Iterable[str],
    source_id: str,
    source_url: str,
    download_url: str,
    license_name: str,
    attribution: str,
    style: str,
    priority: int = 100,
) -> dict:
    return {
        "id": asset_id,
        "type": kind,
        "path": path.as_posix(),
        "tags": list(dict.fromkeys(tags)),
        "keywords": list(dict.fromkeys(keywords)),
        "style": style,
        "priority": priority,
        "recommended_templates": list(dict.fromkeys(recommended_templates)),
        "dimensions": _dimensions(ASSET_ROOT / path),
        "source_id": source_id,
        "source_url": source_url,
        "download_url": download_url,
        "license": license_name,
        "attribution": attribution,
    }


def _replace_asset_directories(staged_assets: Path) -> None:
    asset_root = ASSET_ROOT.resolve()
    for relative in (Path("icons/core"), Path("backgrounds/editorial"), Path("photos"), Path("patterns/subtle")):
        target = (asset_root / relative).resolve()
        target.relative_to(asset_root)
        if target.exists():
            shutil.rmtree(target)
        shutil.copytree(staged_assets / relative, target)


def _download_content_photos(staged_assets: Path) -> None:
    for spec in CONTENT_PHOTOS:
        url = PICSUM_URL.format(photo_id=spec.photo_id)
        data, _ = _fetch(url)
        _save_photo(data, staged_assets / "photos" / spec.category / f"{spec.asset_id}.jpg", spec.asset_id)


def _content_photo_entries() -> list[dict]:
    entries: list[dict] = []
    for spec in CONTENT_PHOTOS:
        relative = Path("photos") / spec.category / f"{spec.asset_id}.jpg"
        entries.append(_asset_entry(
            asset_id=spec.asset_id,
            kind="photo",
            path=relative,
            tags=(spec.category, *spec.tags),
            keywords=spec.keywords,
            recommended_templates=("image_text", "image_hero"),
            source_id="unsplash",
            source_url=spec.source_url,
            download_url=PICSUM_URL.format(photo_id=spec.photo_id),
            license_name="Unsplash License",
            attribution=f"Photo by {spec.author} on Unsplash",
            style="content-photo",
            priority=spec.priority,
        ))
    return entries


def sync_content_photos(staging_root: Path) -> None:
    """Replace only categorized photos while preserving other valid assets."""
    staged_assets = staging_root / "assets"
    _download_content_photos(staged_assets)

    asset_root = ASSET_ROOT.resolve()
    target = (asset_root / "photos").resolve()
    target.relative_to(asset_root)
    if target.exists():
        shutil.rmtree(target)
    shutil.copytree(staged_assets / "photos", target)

    manifest_path = ASSET_ROOT / "manifest.json"
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    manifest["updated_at"] = date.today().isoformat()
    manifest["assets"] = [
        asset for asset in manifest.get("assets", []) if asset.get("type") != "photo"
    ] + _content_photo_entries()
    manifest_path.write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )


def sync_assets(staging_root: Path) -> None:
    staged_assets = staging_root / "assets"
    for spec in ICONS:
        url = ICON_URL.format(style=ICON_STYLE, color=ICON_COLOR, slug=spec.slug)
        try:
            data, headers = _fetch(url)
        except Exception as exc:
            raise RuntimeError(f"failed to synchronize icon {spec.asset_id}") from exc
        _save_icon(data, staged_assets / "icons" / "core" / f"{spec.asset_id}.png", spec.asset_id, headers)
    for spec in ASSET_BACKGROUNDS:
        url = PICSUM_URL.format(photo_id=spec.photo_id)
        data, _ = _fetch(url)
        _save_photo(data, staged_assets / "backgrounds" / "editorial" / f"{spec.asset_id}.jpg", spec.asset_id)
    _download_content_photos(staged_assets)
    for spec in PATTERNS:
        data, _ = _fetch(spec.source_url)
        _save_pattern(data, staged_assets / "patterns" / "subtle" / spec.filename, spec.asset_id)

    _replace_asset_directories(staged_assets)

    assets: list[dict] = []
    for spec in ICONS:
        relative = Path("icons") / "core" / f"{spec.asset_id}.png"
        download_url = ICON_URL.format(style=ICON_STYLE, color=ICON_COLOR, slug=spec.slug)
        assets.append(_asset_entry(
            asset_id=spec.asset_id,
            kind="icon",
            path=relative,
            tags=(spec.asset_id, *spec.tags),
            keywords=spec.keywords,
            recommended_templates=("icon_grid", "card_grid", "content_slide", "image_text"),
            source_id="icons8",
            source_url=f"https://icons8.com/icon/set/{spec.slug}/{ICON_STYLE}",
            download_url=download_url,
            license_name="Icons8 License",
            attribution="Icons by Icons8",
            style=ICON_STYLE,
            priority=spec.priority,
        ))
    for spec in ASSET_BACKGROUNDS:
        relative = Path("backgrounds") / "editorial" / f"{spec.asset_id}.jpg"
        assets.append(_asset_entry(
            asset_id=spec.asset_id,
            kind="background",
            path=relative,
            tags=spec.tags,
            keywords=spec.tags,
            recommended_templates=spec.recommended_templates,
            source_id="unsplash",
            source_url=spec.source_url,
            download_url=PICSUM_URL.format(photo_id=spec.photo_id),
            license_name="Unsplash License",
            attribution=f"Photo by {spec.author} on Unsplash",
            style="editorial-photo",
        ))
    assets.extend(_content_photo_entries())
    for spec in PATTERNS:
        relative = Path("patterns") / "subtle" / spec.filename
        assets.append(_asset_entry(
            asset_id=spec.asset_id,
            kind="pattern",
            path=relative,
            tags=spec.tags,
            keywords=spec.tags,
            recommended_templates=("content_slide", "card_grid", "chart_slide", "section_divider"),
            source_id="transparent-textures",
            source_url=spec.source_url,
            download_url=spec.source_url,
            license_name="CC BY 3.0",
            attribution="Transparent Textures",
            style="subtle-pattern",
        ))

    manifest = {
        "version": 2,
        "updated_at": date.today().isoformat(),
        "discovery_urls": [BING_DISCOVERY_URL, "https://icons8.com/icons"],
        "sources": SOURCES,
        "assets": assets,
    }
    target = ASSET_ROOT / "manifest.json"
    target.write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def sync_background_templates(staging_root: Path) -> None:
    additions_root = staging_root / "background_templates"
    additions_by_path: dict[str, ThemePhotoSpec] = {}
    for spec in THEME_ADDITIONS:
        data, _ = _fetch(PICSUM_URL.format(photo_id=spec.photo_id))
        staged = additions_root / spec.theme / "images" / spec.filename
        _save_photo(data, staged, f"{spec.theme}/{spec.filename}")
        additions_by_path[f"{spec.theme}/images/{spec.filename}"] = spec

    for relative, spec in additions_by_path.items():
        target = BACKGROUND_ROOT / relative
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(additions_root / relative, target)

    themes: list[dict] = []
    for theme_id, config in THEMES.items():
        images: list[dict] = []
        for filename in config["existing"]:
            relative = f"{theme_id}/images/{filename}"
            target = BACKGROUND_ROOT / relative
            if not target.is_file():
                raise FileNotFoundError(f"missing existing background: {relative}")
            images.append({
                "path": relative,
                "dimensions": _dimensions(target),
                "source_id": "project-existing",
                "source_url": "",
                "download_url": "",
                "license": "project-internal",
                "attribution": "",
            })
        for spec in THEME_ADDITIONS:
            if spec.theme != theme_id:
                continue
            relative = f"{theme_id}/images/{spec.filename}"
            images.append({
                "path": relative,
                "dimensions": _dimensions(BACKGROUND_ROOT / relative),
                "source_id": "unsplash",
                "source_url": spec.source_url,
                "download_url": PICSUM_URL.format(photo_id=spec.photo_id),
                "license": "Unsplash License",
                "attribution": f"Photo by {spec.author} on Unsplash",
            })
        themes.append({
            "id": theme_id,
            "name_cn": config["name_cn"],
            "scenarios": config["scenarios"],
            "priority": config["priority"],
            "recommended_palette": config["recommended_palette"],
            "images": images,
        })

    manifest = {
        "version": 1,
        "updated_at": date.today().isoformat(),
        "discovery_url": BING_DISCOVERY_URL,
        "sources": SOURCES[1:2],
        "themes": themes,
    }
    (BACKGROUND_ROOT / "manifest.json").write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--assets-only", action="store_true", help="Replace only assets/ and its manifest")
    parser.add_argument("--backgrounds-only", action="store_true", help="Update only background_templates/")
    parser.add_argument("--photos-only", action="store_true", help="Replace only categorized content photos")
    args = parser.parse_args()
    if sum((args.assets_only, args.backgrounds_only, args.photos_only)) > 1:
        parser.error("--assets-only, --backgrounds-only and --photos-only are mutually exclusive")

    with tempfile.TemporaryDirectory(prefix="ppt-agent-assets-") as temp_dir:
        staging_root = Path(temp_dir)
        if args.photos_only:
            sync_content_photos(staging_root)
        elif not args.backgrounds_only:
            sync_assets(staging_root)
        if not args.assets_only and not args.photos_only:
            sync_background_templates(staging_root)
    print(f"synchronized visual assets under {ROOT}")


if __name__ == "__main__":
    main()
