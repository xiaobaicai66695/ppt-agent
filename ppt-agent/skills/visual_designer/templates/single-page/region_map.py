TEMPLATE = {
    "type": "region_map",
    "name": "区域版图页",
    "description": "地图式布局，适合展示全国/区域业务覆盖、市场分布或战略布局。左侧用抽象地图区域表示不同区域，右侧面板展示区域业绩数据。",
    "layout_hint": "map + right_panel",
    "elements": {
        "kicker": {
            "font_size": "12pt",
            "color": "text_muted",
            "text": "{领域标签}",
            "margin_bottom": "0.1in"
        },
        "title": {
            "position": "top",
            "font_size": "32pt",
            "font_weight": "bold",
            "alignment": "left",
            "margin_bottom": "0.2in",
            "max_chars": 30
        },
        "subtitle": {
            "font_size": "14pt",
            "color": "text_muted",
            "margin_bottom": "0.3in",
            "max_chars": 40
        },
        "map_regions": [
            {
                "name": "{区域1}",
                "shape": "irregular_polygon",
                "color": "primary",
                "value": "{份额}%",
                "trend": "{增长率}"
            },
            {
                "name": "{区域2}",
                "shape": "irregular_polygon",
                "color": "secondary",
                "value": "{份额}%",
                "trend": "{增长率}"
            },
            {
                "name": "{区域3}",
                "shape": "irregular_polygon",
                "color": "accent",
                "value": "{份额}%",
                "trend": "{增长率}"
            },
            {
                "name": "{区域4}",
                "shape": "irregular_polygon",
                "color": "light_bg",
                "value": "{份额}%",
                "trend": "{增长率}"
            }
        ],
        "right_panel": {
            "title": "区域业绩概览",
            "regions_data": [
                {
                    "name": "{区域1}",
                    "value": "{份额}%",
                    "trend": "{增长率}",
                    "detail": "{经济圈/定位描述}"
                },
                {
                    "name": "{区域2}",
                    "value": "{份额}%",
                    "trend": "{增长率}",
                    "detail": "{经济圈/定位描述}"
                },
                {
                    "name": "{区域3}",
                    "value": "{份额}%",
                    "trend": "{增长率}",
                    "detail": "{经济圈/定位描述}"
                },
                {
                    "name": "{区域4}",
                    "value": "{份额}%",
                    "trend": "{增长率}",
                    "detail": "{经济圈/定位描述}"
                },
                {
                    "name": "{区域5}",
                    "value": "{份额}%",
                    "trend": "{增长率}",
                    "detail": "{经济圈/定位描述}"
                },
                {
                    "name": "{区域6}",
                    "value": "{份额}%",
                    "trend": "{增长率}",
                    "detail": "{经济圈/定位描述}"
                }
            ]
        }
    },
    "visual_elements": [
        "左侧地图区域用不规则多边形表示（可抽象化处理），比实际地图简化",
        "每个区域用不同深浅的色块填充：primary（最深）、secondary（中等）、accent、light_bg（最浅）",
        "区域形状带轻微随机抖动模拟手绘感",
        "区域名称标注在区域中心位置，深色区域用白色字，浅色区域用深色字",
        "右侧面板用 background 色背景，顶部 primary 色强调线（3pt）",
        "面板内每个区域用一行展示：区域名（加粗）+ 份额（大号primary色）+ 增长率（小号secondary色）+ 描述（text_muted）",
        "每行左侧有优先级/重要度色条（3pt宽）",
        "整体布局左右比例约 55%:45%"
    ],
    "example": {
        "kicker": "战略布局",
        "title": "全国业务版图",
        "subtitle": "战略布局覆盖全国主要经济区域",
        "regions": [
            {"name": "华东", "value": "35%", "trend": "+18%"},
            {"name": "华北", "value": "28%", "trend": "+12%"},
            {"name": "华南", "value": "18%", "trend": "+25%"},
            {"name": "西南", "value": "6%", "trend": "+10%"},
            {"name": "东北", "value": "5%", "trend": "+3%"},
            {"name": "华中", "value": "8%", "trend": "+15%"}
        ],
        "regions_detail": [
            {"name": "华东地区", "value": "35%", "trend": "+18%", "detail": "长三角核心区"},
            {"name": "华北地区", "value": "28%", "trend": "+12%", "detail": "京津冀协同发展"},
            {"name": "华南地区", "value": "18%", "trend": "+25%", "detail": "粤港澳大湾区"},
            {"name": "西南地区", "value": "6%", "trend": "+10%", "detail": "成渝双城经济圈"},
            {"name": "东北地区", "value": "5%", "trend": "+3%", "detail": "老工业基地"},
            {"name": "华中地区", "value": "8%", "trend": "+15%", "detail": "中部崛起战略"}
        ]
    },
    "example_2": {
        "kicker": "{领域标签}",
        "title": "{版图标题}",
        "subtitle": "{说明}",
        "regions": [
            {"name": "{区域1}", "value": "{份额}%", "trend": "{增长率}"},
            {"name": "{区域2}", "value": "{份额}%", "trend": "{增长率}"},
            {"name": "{区域3}", "value": "{份额}%", "trend": "{增长率}"},
            {"name": "{区域4}", "value": "{份额}%", "trend": "{增长率}"}
        ],
        "regions_detail": [
            {"name": "{区域1}", "value": "{份额}%", "trend": "{增长率}", "detail": "{经济圈描述}"},
            {"name": "{区域2}", "value": "{份额}%", "trend": "{增长率}", "detail": "{经济圈描述}"},
            {"name": "{区域3}", "value": "{份额}%", "trend": "{增长率}", "detail": "{经济圈描述}"},
            {"name": "{区域4}", "value": "{份额}%", "trend": "{增长率}", "detail": "{经济圈描述}"}
        ]
    },
    "when_to_use": [
        "全国/区域业务布局展示",
        "市场份额分布",
        "市场覆盖战略",
        "区域业绩对比"
    ],
    "never": [
        "禁止地图过于复杂（用抽象简化版）",
        "禁止区域超过8个",
        "禁止右侧面板数据不完整",
        "禁止区域颜色重复导致难以区分"
    ]
}
