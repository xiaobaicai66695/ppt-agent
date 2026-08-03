# Background Templates

## 背景模板库

本目录包含可用的PPT背景模板，按主题分类。所有图片为 **JPG格式 (1920x1080)**，适配16:9横屏PPT。

## 目录结构

```
background_templates/
├── SKILL.md                    # 本文件
├── party_government/          # 党政办公 (5张)
├── ink_wash_mountain/         # 水墨山水 (4张)
├── vintage_chinese/           # 复古中国风 (1张)
├── minimalist_blue/           # 简约蓝白 (1张)
├── snowy_mountain/           # 雪山风景 (3张)
└── artistic/                 # 艺术涂鸦 (1张)
```

## 主题说明

| 主题标识 | 主题名称 | 适用场景 |
|----------|----------|----------|
| `party_government` | 党政办公 | 党建、政府汇报 |
| `ink_wash_mountain` | 水墨山水 | 中国风、文艺 |
| `vintage_chinese` | 复古中国风 | 传统文化 |
| `minimalist_blue` | 简约蓝白 | 商务、科技 |
| `snowy_mountain` | 雪山风景 | 自然、户外 |
| `artistic` | 艺术涂鸦 | 艺术、创意、时尚 |

## 使用方式

### 1. 单页模板 (atomic)

在 `templates/single-page/*.json` 中添加 `background` 字段：

```json
{
  "name": "title_slide",
  "display_name": "封面页",
  "fields": [
    {"name": "title", "label": "主标题", "type": "text"},
    {"name": "background", "label": "背景图片", "type": "select",
     "options": [
       {"value": "", "label": "不使用背景"},
       {"value": "party_government", "label": "党政办公"},
       {"value": "minimalist_blue", "label": "简约蓝白"},
       {"value": "ink_wash_mountain", "label": "水墨山水"}
     ]}
  ]
}
```

### 2. 全局模板 (preset)

在 `templates/full-decks/*.json` 的根级别添加 `background_options` 字段：

```json
{
  "name": "generic",
  "display_name": "通用模板",
  "background_options": {
    "themes": ["", "party_government", "minimalist_blue"],
    "labels": ["不使用背景", "党政办公", "简约蓝白"]
  },
  "default_slides": [...]
}
```

### 3. Python 代码

```python
from generators.background_manager import get_background

# 按主题获取
get_background(theme="party_government")

# 按场景获取
get_background(scenario="党建汇报")
```

### 4. 亮度设置

背景亮度由具体生成器按页面类型自动选择，目标是在图片主题可辨识的同时保证标题和正文对比度。直接调用 `set_image_background` 时可通过 `brightness` 参数调整：

```python
# brightness 范围 0.0-1.0
set_image_background(slide, bg_path, brightness=0.8)  # 稍暗
set_image_background(slide, bg_path, brightness=0.95)  # 明亮清晰
```

## 动态扩充

添加新背景：

1. 在 `background_templates/` 下创建主题目录，放入 JPG 图片
2. 更新 `generators/background_manager.py` 的 `THEME_MAPPING`：

```python
THEME_MAPPING = {
    "tech_blue": {               # 新主题目录名
        "name_cn": "科技蓝",
        "scenarios": ["科技", "技术"],
        "priority": 6,
    },
}
```
