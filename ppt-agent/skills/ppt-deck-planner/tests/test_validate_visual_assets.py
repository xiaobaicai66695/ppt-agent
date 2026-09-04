import json
import sys
import tempfile
import unittest
from pathlib import Path

SKILL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SKILL_ROOT))

from generators.validate_visual_assets import validate_visual_manifest_file
from generators.validate_deck import validate_manifest_file


PNG_1X1 = bytes.fromhex(
    "89504e470d0a1a0a0000000d4948445200000001000000010802000000907753de"
    "0000000c4944415408d763f8ffff3f0005fe02fea73581e40000000049454e44ae426082"
)


class ValidateVisualAssetsTest(unittest.TestCase):
    def test_required_policy_rejects_query_only_image_plan(self):
        with tempfile.TemporaryDirectory() as tmp:
            work_dir = Path(tmp)
            manifest_path = work_dir / "tasks.json"
            manifest_path.write_text(json.dumps({
                "title": "智能制造",
                "visual_policy": {
                    "mode": "required",
                    "min_image_pages": 1
                },
                "tasks": [
                    {
                        "page_index": 1,
                        "title": "智能工厂现场",
                        "content_type": "image_text",
                        "content_plan": {
                            "visual_intent": {
                                "asset_purpose": "scene",
                                "asset_query": "smart factory robotic arm",
                                "asset_subject": "robotic arm on factory line",
                                "composition": "wide landscape with copy space on left",
                                "orientation": "landscape",
                                "provider": "unsplash",
                                "search_status": "planned"
                            }
                        },
                    }
                ],
            }, ensure_ascii=False), encoding="utf-8")

            result = validate_visual_manifest_file(manifest_path, work_dir)

        self.assertFalse(result["ok"])
        self.assertIn("unmaterialized_visual_asset", {error["code"] for error in result["errors"]})
        self.assertIn("too_few_image_pages", {error["code"] for error in result["errors"]})

    def test_required_policy_accepts_materialized_image_with_attribution(self):
        with tempfile.TemporaryDirectory() as tmp:
            work_dir = Path(tmp)
            image_dir = work_dir / "assets" / "images"
            image_dir.mkdir(parents=True)
            (image_dir / "factory.png").write_bytes(PNG_1X1)
            manifest_path = work_dir / "tasks.json"
            manifest_path.write_text(json.dumps({
                "title": "智能制造",
                "tasks": [
                    {
                        "page_index": 1,
                        "title": "智能工厂现场",
                        "content_type": "image_text",
                        "content_plan": {
                            "visual_intent": {
                                "asset_purpose": "scene",
                                "asset_query": "smart factory robotic arm",
                                "asset_subject": "robotic arm on factory line",
                                "composition": "wide landscape with copy space on left",
                                "orientation": "landscape",
                                "local_path": "assets/images/factory.png",
                                "source_url": "https://unsplash.com/photos/factory",
                                "attribution": "Photo by Test Photographer on Unsplash",
                                "provider": "unsplash",
                                "search_status": "resolved"
                            }
                        },
                    }
                ],
            }, ensure_ascii=False), encoding="utf-8")

            result = validate_visual_manifest_file(manifest_path, work_dir)

        self.assertTrue(result["ok"], result)
        self.assertEqual(result["image_page_count"], 1)

    def test_deck_preflight_accepts_deck_without_background_images(self):
        with tempfile.TemporaryDirectory() as tmp:
            work_dir = Path(tmp)
            manifest_path = work_dir / "tasks.json"
            manifest_path.write_text(json.dumps({
                "title": "智能制造",
                "tasks": [{
                    "task_id": "slide_01",
                    "page_index": 1,
                    "title": "智能工厂现场",
                    "content_type": "content_slide",
                    "content_plan": {
                        "summary": "无背景图片时仍可验证并渲染结构化语义页面。",
                        "components": [{
                            "type": "headline",
                            "text": "背景图片是可选视觉增强，不是渲染前置条件"
                        }]
                    }
                }]
            }, ensure_ascii=False), encoding="utf-8")
            result = validate_manifest_file(manifest_path, SKILL_ROOT / "templates" / "component_contracts.json", work_dir)

        self.assertTrue(result["ok"], result)

    def test_same_content_type_must_reuse_one_materialized_background(self):
        with tempfile.TemporaryDirectory() as tmp:
            work_dir = Path(tmp)
            image_dir = work_dir / "assets" / "images"
            image_dir.mkdir(parents=True)
            (image_dir / "one.png").write_bytes(PNG_1X1)
            (image_dir / "two.png").write_bytes(PNG_1X1)
            manifest_path = work_dir / "tasks.json"
            manifest_path.write_text(json.dumps({
                "title": "背景复用",
                "tasks": [
                    {
                        "page_index": 1,
                        "content_type": "content_slide",
                        "content_plan": {"visual_intent": {
                            "asset_purpose": "background", "local_path": "assets/images/one.png",
                            "provider": "unsplash", "search_status": "resolved"
                        }}
                    },
                    {
                        "page_index": 2,
                        "content_type": "content_slide",
                        "content_plan": {"visual_intent": {
                            "asset_purpose": "background", "local_path": "assets/images/two.png",
                            "provider": "unsplash", "search_status": "resolved"
                        }}
                    },
                    {
                        "page_index": 3,
                        "content_type": "chart_slide",
                        "content_plan": {"visual_intent": {
                            "asset_purpose": "background", "local_path": "assets/images/two.png",
                            "provider": "unsplash", "search_status": "resolved"
                        }}
                    }
                ]
            }, ensure_ascii=False), encoding="utf-8")

            result = validate_visual_manifest_file(manifest_path, work_dir)

        self.assertFalse(result["ok"])
        reuse_errors = [error for error in result["errors"] if error["code"] == "background_not_reused_by_content_type"]
        self.assertEqual(2, len(reuse_errors), reuse_errors)


if __name__ == "__main__":
    unittest.main()
