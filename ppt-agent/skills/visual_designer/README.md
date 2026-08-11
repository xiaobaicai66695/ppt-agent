# Visual Designer 素材库

`visual_designer` 是 Agent 使用的离线 PPT 渲染能力包。运行时不下载任何素材；图标、可替换内容图片、编辑类背景、纹理和主题背景图片都随仓库交付，并通过带版本的元数据登记。

## 目录契约

```text
visual_designer/
├── assets/
│   ├── manifest.json                 # v2 素材元数据与来源记录
│   ├── icons/core/                   # 512x512 透明 PNG 图标
│   ├── photos/<category>/            # 1920x1080 可替换内容图片
│   ├── backgrounds/editorial/        # 1920x1080 JPG 编辑类背景
│   └── patterns/subtle/              # 1920x1080 透明 PNG 纹理
├── background_templates/
│   ├── manifest.json                 # 主题、配色、适用场景和图片元数据
│   └── <theme>/images/*.jpg
├── generators/
│   ├── asset_manager.py              # 素材查找与语义图标/图片匹配
│   └── background_manager.py         # 基于 manifest 的背景主题选择
└── scripts/sync_external_assets.py   # 精选素材下载与规格归一化脚本
```

## 当前基线

- 79 个 Icons8 Fluency Systems Regular 图标，覆盖结构、商业、科技、党政、地理、自然、数据和收尾页概念。
- 14 张分类内容图片、6 张编辑类照片背景，以及 4 张轻纹理素材。
- 10 类主题背景：`party_government` 5 张，其余 9 类各 4 张。
- 背景图片统一放在 `<theme>/images/<序号>.jpg`，序号从 `1.jpg` 开始，目录名应准确表达主题。

## 运行规则

- 当 `content_plan` 已经明确概念时，优先使用显式图标 id。
- 否则调用 `icon_id_from_text`，根据 manifest 中的 `keywords` 和 `tags` 做匹配，并按最长关键词和优先级排序。
- 未知语义返回空 id。不要恢复 `layout`、`primitive`、`review` 或缩写类占位图标。
- `image_text` 的默认图片使用 `photo_id_from_text`。显式本地路径或已登记 photo id 优先；否则按语义选择本地分类图片，并回退到 `photo_business_work`。
- 内容图片必须作为一张可替换的 PowerPoint 图片插入。不要用图标、纹理、标签或组合形状拼出图片区域。
- 任一主题内的图片都适用时，可以使用 `minimalist_blue` 这类主题 id。
- 需要避免相邻页面重复图片时，使用 `<theme>/images/<file>.jpg` 指定具体背景。
- 运行时必须保持离线。网络 URL 只作为维护元数据，不作为运行依赖。

## 同步来源

在 `ppt-agent/skills/visual_designer` 目录下运行，环境中需要安装 Pillow：

```powershell
python scripts/sync_external_assets.py
python scripts/sync_external_assets.py --photos-only  # 只同步分类内容图片
```

脚本会把精选来源下载到临时暂存目录，校验图片有效性，归一化尺寸，然后替换 4 个 `assets` 子目录并重新生成 manifest。下载失败时不会留下半替换状态的素材库。

manifest 元数据中的 Bing 图片 URL 只是发现入口，不是授权来源。除非能确认原始页面和复用条款，否则不要导入 Bing 缩略图或转载图片。

## 验证

```powershell
python -m py_compile generators/*.py scripts/sync_external_assets.py
python -m unittest discover -s tests -p "test_asset_library.py"
```

涉及视觉变更时，还要生成有代表性的 `title_slide`、`section_divider`、`icon_grid`、`image_text`、`quote_slide` 和 `summary_slide` 页面，渲染后检查图标语义、裁剪、对比度、重叠和未解析占位符。

## 署名与许可证

- Icons8：`Icons by Icons8`，受 [Icons8 License](https://icons8.com/license) 约束。移除署名前必须确认当前部署使用的是付费授权还是免费授权，以及对应义务。
- 照片：每张外部照片都在相关 manifest 中记录 Unsplash 页面、作者、下载 URL 和 `Unsplash License`。
- 纹理：`Transparent Textures`，在素材 manifest 中记录为 CC BY 3.0。
- 无法恢复来源的既有项目背景统一标记为 `project-existing`，不得把它们描述成外部授权素材。

当生成的演示文稿在需要署名的许可证下分发时，应在 PPT 致谢页或随附材料中包含已记录的署名信息。

## 新增或替换素材

1. 在 `scripts/sync_external_assets.py` 中新增精选素材规格。
2. 补充中英文语义关键词和稳定来源页面。
3. 运行同步脚本；不要手工修改生成出的尺寸字段。
4. 运行 manifest 测试和聚焦渲染冒烟。
5. 当选择行为或公开生成契约变化时，同步更新 `SKILL.md` 或 `references/generators.md`。
