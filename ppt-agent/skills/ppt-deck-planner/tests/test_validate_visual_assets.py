import json
import sys
import tempfile
import unittest
from pathlib import Path

SKILL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SKILL_ROOT))

from generators.validate_visual_assets import validate_visual_manifest_file


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
                    "min_image_pages": 1,
                    "required_roles": ["scene_evidence"],
                },
                "tasks": [
                    {
                        "page_index": 1,
                        "title": "智能工厂现场",
                        "content_type": "image_text",
                        "content_plan": {
                            "components": [
                                {
                                    "type": "image",
                                    "asset_purpose": "scene",
                                    "asset_query": "smart factory robotic arm",
                                    "search_status": "planned",
                                }
                            ]
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
                "visual_policy": {
                    "mode": "required",
                    "min_image_pages": 1,
                    "required_roles": ["scene_evidence"],
                },
                "tasks": [
                    {
                        "page_index": 1,
                        "title": "智能工厂现场",
                        "content_type": "image_text",
                        "content_plan": {
                            "components": [
                                {
                                    "type": "image",
                                    "asset_purpose": "scene",
                                    "asset_query": "smart factory robotic arm",
                                    "local_path": "assets/images/factory.png",
                                    "source_url": "https://example.com/factory",
                                    "search_status": "downloaded",
                                }
                            ]
                        },
                    }
                ],
            }, ensure_ascii=False), encoding="utf-8")

            result = validate_visual_manifest_file(manifest_path, work_dir)

        self.assertTrue(result["ok"], result)
        self.assertEqual(result["image_page_count"], 1)


if __name__ == "__main__":
    unittest.main()
