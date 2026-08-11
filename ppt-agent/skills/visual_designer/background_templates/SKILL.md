# Background Templates

## 背景模板库

本目录包含可用的 PPT 背景模板，按主题分类。图片均为 16:9 兼容 JPG（至少 1280x720），生成时统一裁切为 1920x1080。

`manifest.json` 是主题、场景、推荐 palette、图片路径、来源和许可的唯一元数据来源；`generators/background_manager.py` 只在 manifest 缺失时进行旧目录扫描。

每个主题目录下的图片统一放在 `images/`，命名为数字文件名，例如 `1.jpg`、`2.jpg`。每类至少 4 张，可超过 4 张；图片内容必须与目录主题一致、同目录内风格相近，并避免黑白配色图片进入 manifest。

## 目录结构

```
background_templates/
├── SKILL.md                    # 本文件
├── manifest.json               # 主题和图片元数据
├── party_government/           # 党政办公
├── minimalist_blue/            # 商务科技蓝
├── business_gradient/          # 商务渐变
├── ink_wash_mountain/          # 彩墨山水
├── vintage_chinese/            # 复古中国风
├── education_warm/             # 教育暖阳
├── medical_clean/              # 医疗清新
├── eco_nature/                 # 生态自然
├── snowy_mountain/             # 雪山风景
└── artistic/                   # 艺术创意
```

## 主题说明

| 主题标识 | 主题名称 | 适用场景 |
|----------|----------|----------|
| `party_government` | 党政办公 | 党建、政府汇报 |
| `minimalist_blue` | 商务科技蓝 | 商务、科技 |
| `business_gradient` | 商务渐变 | 经营、咨询、投标 |
| `ink_wash_mountain` | 彩墨山水 | 中国风、文艺 |
| `vintage_chinese` | 复古中国风 | 传统文化 |
| `education_warm` | 教育暖阳 | 课程、培训 |
| `medical_clean` | 医疗清新 | 医疗、健康 |
| `eco_nature` | 生态自然 | 环保、可持续 |
| `snowy_mountain` | 雪山风景 | 自然、户外 |
| `artistic` | 艺术创意 | 艺术、创意、时尚 |

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
       {"value": "minimalist_blue", "label": "商务科技蓝"},
       {"value": "ink_wash_mountain", "label": "彩墨山水"}
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
    "themes": ["", "party_government", "minimalist_blue", "business_gradient"],
    "labels": ["不使用背景", "党政办公", "商务科技蓝", "商务渐变"]
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

### 4. 随机选择与防重复

任务规划阶段会把主题 id 转换为主题内具体图片引用，例如：

```json
{
  "background": "party_government/images/3.jpg"
}
```

启用背景时，整套 PPT 必须使用同一个主题目录下的图片；不同页面可以在该目录内轮换具体图片，避免跨主题混用造成风格漂移。每个主题至少 4 张图片，规划阶段应尽量避免相邻页面重复。

生成器仍兼容旧写法：

```python
background="party_government"
```

此时会在主题下随机选择图片，但跨独立 Python 进程无法保证相邻页不重复；需要严格防重复时，应使用具体图片引用。

### 5. 亮度与蒙版

背景亮度由具体生成器按页面类型自动选择，目标是在图片主题可辨识的同时保证标题和正文对比度。直接调用 `set_image_background` 时可通过 `brightness` 参数调整：

```python
# brightness 范围 0.0-1.0
set_image_background(slide, bg_path, brightness=0.92)  # 轻微柔化
set_image_background(slide, bg_path, brightness=0.98)  # 明亮清晰
```

所有图片背景都会自动叠加浅色磨砂玻璃蒙版。标题页不得默认使用过暗背景；正文密集时应增加局部卡片或面板，而不是继续压暗整张图。

## 动态扩充

添加新背景：

1. 确认图片内容与目录名一致、同目录内风格相近，不使用黑白配色图。
2. 将图片归一化到 `<theme>/images/<数字>.jpg`，每类至少保留 4 张可用图片。
3. 重建 manifest，并运行背景 manifest 校验和代表性页面 smoke test。

```powershell
python scripts/sync_external_assets.py
python -m unittest discover -s tests -p "test_asset_library.py"
```

不要再修改 Python `THEME_MAPPING` 作为正常维护路径。Bing 图片搜索可用于发现素材，但入库前必须回到原始页面确认许可；来源不明、转载、水印图不得登记为可复用背景。
