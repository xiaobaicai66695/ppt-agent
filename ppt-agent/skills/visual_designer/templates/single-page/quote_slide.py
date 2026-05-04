TEMPLATE = {
    "type": "quote_slide",
    "name": "金句引言页",
    "description": "高亮一条引言/金句，配合引言者信息。可用于演讲中的关键引用、数据金句、名人名言等。",
    "layout_hint": "center",
    "elements": {
        "kicker": {
            "font_size": "12pt",
            "color": "text_muted",
            "text": "金句"
        },
        "quote": {
            "position": "center",
            "font_size": "28-36pt",
            "font_weight": "bold",
            "font_style": "italic",
            "color": "text",
            "line_spacing": "1.4x",
            "max_chars": 60,
            "alignment": "center"
        },
        "attribution": {
            "position": "below_quote",
            "font_size": "14pt",
            "color": "secondary",
            "format": "—— {引言者}，{来源/职位}"
        },
        "quote_mark": {
            "type": "「" or """,
            "font_size": "72pt",
            "color": "primary",
            "opacity": 0.15,
            "position": "top_left 和 bottom_right"
        }
    },
    "visual_elements": [
        "页面大面积留白，引言居中",
        "左上角和右下角放置大号引号「」（primary色，透明度15%）作为装饰",
        "引言正文用加粗衬线或黑体，28-36pt",
        "引言者信息用次要色，置于引言下方，可配合细线分隔",
        "背景可用极浅的渐变色（background→light_bg）或纯色",
        "可添加左侧竖线装饰（primary色，3pt宽）作为视觉锚点"
    ],
    "example": {
        "kicker": "金句",
        "quote": "预测未来的最好方式，就是创造未来。",
        "attribution": "—— Alan Kay，计算机科学家",
        "decoration": {
            "quote_mark_top_left": "「",
            "quote_mark_bottom_right": "」",
            "vertical_bar": true
        }
    },
    "example_2": {
        "kicker": "核心洞察",
        "quote": "数据是新时代的石油，但只有精炼后才能释放价值。",
        "attribution": "—— {高管姓名}，{公司名称} CEO",
        "decoration": {
            "vertical_bar": true,
            "background_gradient": "light_bg → background"
        }
    },
    "when_to_use": [
        "演讲中需要用一句有力的话点明主题",
        "引用权威观点、数据金句、关键结论",
        "章节过渡时用金句承接上下内容",
        "需要用视觉力量强化核心信息时"
    ],
    "never": [
        "禁止引言字号小于24pt",
        "禁止没有引言者信息",
        "禁止引言超过2行——超过则拆分到多页",
        "禁止在引言页放bullet列表",
        "禁止引言者信息字号大于引言正文"
    ]
}
