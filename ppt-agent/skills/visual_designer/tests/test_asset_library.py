from __future__ import annotations

from pathlib import Path
import sys
import unittest


SKILL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SKILL_ROOT))

from generators.asset_manager import (  # noqa: E402
    asset_path,
    icon_id_from_text,
    load_manifest,
    load_manifest_data,
    photo_id_from_text,
    resolve_photo,
    validate_manifest,
)
from generators.background_manager import (  # noqa: E402
    get_background,
    get_palette_for_background,
    list_themes,
    validate_background_manifest,
)
from generators.base import resolve_background  # noqa: E402
from generators.image_text_generator import generate as generate_image_text  # noqa: E402


class AssetManifestTests(unittest.TestCase):
    def test_manifest_is_complete_and_valid(self):
        self.assertEqual([], validate_manifest())
        data = load_manifest_data()
        self.assertEqual(2, data["version"])
        counts = {
            kind: len([asset for asset in load_manifest() if asset["type"] == kind])
            for kind in ("icon", "background", "photo", "pattern")
        }
        self.assertGreaterEqual(counts["icon"], 70)
        self.assertGreaterEqual(counts["background"], 6)
        self.assertGreaterEqual(counts["photo"], 14)
        self.assertGreaterEqual(counts["pattern"], 4)

    def test_registered_assets_resolve_offline(self):
        for asset in load_manifest():
            with self.subTest(asset=asset["id"]):
                resolved = asset_path(asset["id"])
                self.assertIsNotNone(resolved)
                self.assertTrue(Path(resolved).is_file())

    def test_semantic_icons_cover_priority_domains(self):
        cases = {
            "2026 年两会政策解读": "policy",
            "森林生态保护行动": "ecology",
            "中国区域分布与地理格局": "map",
            "感谢聆听": "thanks",
            "联系我们": "contact",
            "数据安全与隐私合规": "security",
            "制造业数字化转型": "manufacturing",
        }
        for text, expected in cases.items():
            with self.subTest(text=text):
                self.assertEqual(expected, icon_id_from_text(text))

    def test_unknown_semantics_are_omitted(self):
        self.assertEqual("", icon_id_from_text("没有登记过的全新抽象词汇"))
        self.assertEqual("presentation", icon_id_from_text("没有登记过的全新抽象词汇", fallback="presentation"))
        self.assertEqual("", icon_id_from_text("没有登记过的全新抽象词汇", fallback="missing-icon"))

    def test_semantic_photos_cover_content_categories(self):
        cases = {
            "企业团队协作与办公方式": "photo_business_work",
            "人工智能平台与数字化系统": "photo_technology_workspace",
            "书籍阅读与知识文献研究": "photo_education_book",
            "绿色生态与森林保护": "photo_nature_ecology",
            "乡村农业与粮食种植": "photo_agriculture_field",
            "摄影媒体与创意影像": "photo_creative_camera",
        }
        for text, expected in cases.items():
            with self.subTest(text=text):
                self.assertEqual(expected, photo_id_from_text(text))
        self.assertEqual("photo_business_work", photo_id_from_text("未登记主题"))
        self.assertTrue(Path(resolve_photo(text="未登记主题")).is_file())


class ImageTextGeneratorTests(unittest.TestCase):
    def test_missing_image_uses_replaceable_semantic_photo(self):
        prs = generate_image_text(
            title="绿色生态建设",
            header="森林保护行动",
            paragraph="通过长期监测、生态修复和社区共治提升森林系统的稳定性。",
            image_path="",
            background="",
        )
        slide = prs.slides[-1]
        pictures = [shape for shape in slide.shapes if shape.shape_type == 13]
        self.assertEqual(1, len(pictures))
        self.assertTrue(pictures[0].name.startswith("Replaceable photo - photo_nature_ecology"))
        visible_text = "\n".join(
            shape.text for shape in slide.shapes if getattr(shape, "has_text_frame", False)
        )
        for forbidden in ("动态字号", "本地素材", "结构化内容", "图片占位"):
            self.assertNotIn(forbidden, visible_text)

    def test_explicit_photo_asset_wins_and_strip_keeps_one_picture(self):
        prs = generate_image_text(
            title="学习方式升级",
            layout_variant="photo_strip",
            header="从阅读到实践",
            paragraph="课程通过阅读、演示和动手练习形成完整的知识闭环。",
            image_path="asset:photo_education_book",
        )
        pictures = [shape for shape in prs.slides[-1].shapes if shape.shape_type == 13]
        self.assertEqual(1, len(pictures))
        self.assertTrue(pictures[0].name.startswith("Replaceable photo - photo_education_book"))
        self.assertGreater(pictures[0].crop_top + pictures[0].crop_bottom, 0)


class BackgroundManifestTests(unittest.TestCase):
    def test_all_themes_have_rotation_capacity(self):
        self.assertEqual([], validate_background_manifest())
        themes = list_themes()
        self.assertEqual(6, len(themes))
        for theme in themes:
            with self.subTest(theme=theme["theme"]):
                self.assertGreaterEqual(len(theme["images"]), 4)

    def test_theme_and_direct_reference_remain_compatible(self):
        theme_path = get_background(theme="minimalist_blue")
        self.assertIsNotNone(theme_path)
        self.assertTrue(Path(theme_path).is_file())
        direct_path = resolve_background("minimalist_blue/images/2.jpg")
        self.assertIsNotNone(direct_path)
        self.assertTrue(Path(direct_path).is_file())
        self.assertEqual("ocean_soft", get_palette_for_background("minimalist_blue"))


if __name__ == "__main__":
    unittest.main()
