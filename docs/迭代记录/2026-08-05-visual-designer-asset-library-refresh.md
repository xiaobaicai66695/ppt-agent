# Visual Designer 素材库与图文页迭代记录

## 背景

Visual Designer 原有素材量偏少，图标语义覆盖集中在工程概念，背景主题轮换容量不足。图文解说页在没有真实图片时还会用图标、纹理和“动态字号/本地素材”等内部标签拼出假图片区，生成效果生硬且不方便用户继续编辑。

## 本轮目标

- 全量刷新离线图标、编辑背景和纹理资产，并补齐背景主题轮换容量。
- 用 manifest 统一记录语义、尺寸、来源、许可、署名和维护地址。
- 新增独立的分类内容图片库，为图文解说页提供可替换的默认图片。
- 保持 PPT 生成运行时完全离线，外部网络只用于维护脚本同步素材。

## 实现摘要

- `assets/manifest.json` 升级为 v2：登记 79 个 Icons8 Fluency Systems Regular 图标、14 张分类内容图、6 张编辑背景和 4 个纹理。
- `assets/photos/<category>/` 新增商务、科技、教育、城市、自然、运动、交通、文化、创意、农业等分类。
- `background_templates/manifest.json` 统一登记 6 个主题、适用场景、推荐 palette、图片来源和许可；每个主题至少 4 张图。
- `asset_manager.py` 使用最长关键词、命中数量和优先级做确定性语义匹配；未知图标省略，内容图片回退到固定通用办公图。
- `sync_external_assets.py` 支持全量同步和 `--photos-only` 增量同步；下载先进入 staging，通过图片校验后再替换目录。
- `image_text_generator.py` 删除图标/纹理假图片区，左右图文和横向条带统一插入一张中心裁切的 PowerPoint Picture 对象。优先级为显式 `image_path`、语义分类图、任务背景、固定通用图。
- SlideExecutor prompt、单页模板、SKILL、README 和 generator 参考文档同步了 `image_path` 与默认图片契约。

## 关键决策

- Bing 图片搜索只作为发现入口，不直接作为来源或许可依据。
- Icons8 和 Unsplash/Picsum 素材保留来源、许可与 attribution 元数据；运行时不联网。
- 图文页缺图时允许使用默认图片，但必须是真实且可替换的图片对象，不能用形状组、图标或内部标签伪装。
- 内容图片与页面背景分离：图文页使用纯色页面底，避免同一页同时出现大背景图和假图片区造成视觉冲突。

## 验证

- `python -m unittest discover -s tests -p "test_asset_library.py" -v`：9 项通过。
- Visual Designer 全量 generator 与同步脚本 `py_compile` 通过。
- `go test ./pkg/prompts/...` 通过。
- 生成右图、左图、横向条带 3 页 smoke deck，并通过 LibreOffice 转 PDF、Poppler 转 PNG 全页检查。
- 3 页均只有 1 个可替换图片对象；画布越界、文本互相重叠、文本与图片重叠均为 0。
- 14 张分类内容图和代表性图标/背景 contact sheet 已人工检查语义与裁切。

## 关联产物

- OpenSpec：`openspec/changes/archive/2026-08-05-visual-designer-asset-library-refresh/`
- TODO 归档：`docs/issues/done.md` 中 `PPT-QUALITY-011`
- 维护入口：`ppt-agent/skills/visual_designer/README.md`
