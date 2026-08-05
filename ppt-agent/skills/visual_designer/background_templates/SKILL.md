# Background Templates

## 背景模板库

本目录包含可用的 PPT 背景模板，按主题分类。图片均为 16:9 兼容 JPG（至少 1280x720），生成时统一裁切为 1920x1080。

`manifest.json` 是主题、场景、推荐 palette、图片路径、来源和许可的唯一元数据来源；`generators/background_manager.py` 只在 manifest 缺失时进行旧目录扫描。

## 目录结构

```
background_templates/
├── SKILL.md                    # 本文件
├── manifest.json               # 主题和图片元数据
├── party_government/           # 党政办公 (5张)
├── ink_wash_mountain/          # 水墨山水 (4张)
├── vintage_chinese/            # 复古中国风 (4张)
├── minimalist_blue/            # 简约蓝白 (4张)
├── snowy_mountain/             # 雪山风景 (4张)
└── artistic/                   # 艺术创意 (4张)
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

### 4. 随机选择与防重复

任务规划阶段会把主题 id 转换为主题内具体图片引用，例如：

```json
{
  "background": "party_government/images/3.jpg"
}
```

同一主题连续出现在多个视觉页时，必须随机选择图片，并避免相邻视觉页重复同一张。当前每个主题至少 4 张图片，规划阶段仍应写入具体图片引用以保证跨进程防重复。

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

1. 优先在 `scripts/sync_external_assets.py` 中登记原始页面、下载地址、作者和许可。
2. 运行同步脚本，将图片归一化到 `<theme>/images/` 并重建 manifest。
3. 运行背景 manifest 校验和代表性页面 smoke test。

```powershell
python scripts/sync_external_assets.py
python -m unittest discover -s tests -p "test_asset_library.py"
```

不要再修改 Python `THEME_MAPPING` 作为正常维护路径。Bing 图片搜索可用于发现素材，但入库前必须回到原始页面确认许可；来源不明、转载、水印图不得登记为可复用背景。
