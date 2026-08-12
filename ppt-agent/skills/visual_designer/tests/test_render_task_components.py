import unittest
from pathlib import Path
import sys


SKILL_ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SKILL_ROOT))

from generators import render_task


class RenderTaskComponentsTest(unittest.TestCase):
    def test_extract_cards_prefers_semantic_components(self):
        plan = {
            "summary": "三层能力矩阵支撑端到端交付",
            "components": [
                {"id": "headline_1", "type": "headline", "text": "三层能力矩阵支撑端到端交付"},
                {"id": "feature_card_1", "type": "feature_card", "title": "数据接入", "body": "统一连接业务库、文件和接口数据", "emphasis": "primary"},
                {"id": "feature_card_2", "type": "feature_card", "title": "智能分析", "body": "自动识别趋势、异常和关键指标", "emphasis": "normal"},
            ],
        }

        items = render_task.extract_items(plan, {"description": "旧描述"})
        cards = render_task.extract_cards(plan, items)

        self.assertIn("三层能力矩阵支撑端到端交付", items)
        self.assertEqual(cards[0]["header"], "数据接入")
        self.assertEqual(cards[0]["emphasis"], "primary")
        self.assertEqual(cards[1]["body"], "自动识别趋势、异常和关键指标")

    def test_chart_and_kpi_components_map_to_generator_params(self):
        chart_plan = {
            "components": [
                {
                    "type": "chart",
                    "data": {
                        "labels": ["Q1", "Q2"],
                        "datasets": [{"name": "收入", "values": [120, 180]}],
                    },
                }
            ]
        }
        kpi_plan = {
            "components": [
                {"type": "kpi_metric", "title": "转化率", "body": "较上月提升", "data": {"value": "38%", "label": "转化率", "delta": "+6pp", "baseline": "较上月"}},
            ]
        }

        self.assertEqual(render_task.chart_data(chart_plan, [])["labels"], ["Q1", "Q2"])
        self.assertEqual(render_task.kpis(kpi_plan, [])[0]["value"], "38%")


if __name__ == "__main__":
    unittest.main()
