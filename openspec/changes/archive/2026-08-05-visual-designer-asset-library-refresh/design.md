## Context

Visual Designer 当前把 24 个项目内绘制的工程图标、6 张抽象背景和 4 张 pattern 写在 `assets/manifest.json`，中文语义映射则硬编码在 `asset_manager.py`。背景主题另由 `background_manager.py` 的 `THEME_MAPPING` 维护，目录结构、主题元数据和来源信息彼此分离。结果是扩充素材需要同步修改多处，未知内容会回退到 `layout`、`primitive` 或 `review`，并在不相关页面显示成有明确含义的错误图案。

本次需要替换离线 assets，并补齐 `background_templates` 中图片不足的主题。Bing 图片搜索仅作为人工发现入口；正式入库素材必须有稳定下载地址和可追踪来源。Icons8 图标按其许可要求保留 attribution；照片采用 Unsplash 来源并通过 Picsum 的确定性尺寸端点下载；纹理采用 Transparent Textures 并记录来源。

## Goals / Non-Goals

**Goals:**

- 用统一的 Icons8 Fluency Systems Regular 图标集覆盖通用、业务、科技、政务、自然和结束页语义。
- 建立独立于背景图的内容图片分类，覆盖办公、科技、教育、城市、自然、运动、交通、文化、创意和农业等常用图文解说主题。
- 让每个背景主题至少拥有 4 张可轮换图片，并能从 manifest 获取场景、推荐配色、来源和许可。
- 将图标语义匹配从 Python 硬编码迁移到 manifest 的 `keywords`/`tags`，避免无关 fallback。
- 提供可重复执行的同步脚本和离线校验脚本，确保图片尺寸、透明通道、路径和许可元数据完整。
- 保持现有 `background` 参数和主题 id 兼容。

**Non-Goals:**

- 不抓取或镜像 Bing、Icons8 的整个站点，只替换当前 skill 自带资产库所需的精选离线素材。
- 不引入运行时联网下载，PPT 生成仍完全消费本地文件。
- 不调整对外 HTTP API 或模板 id；仅为 `generate_image_text` 增加向后兼容的可选 `image_path` 参数。
- 不在本次建设自动图片搜索、生成式图片服务或素材管理后台。

## Decisions

1. **manifest v2 为素材契约。** `assets/manifest.json` 统一记录 `source_id`、`source_url`、`download_url`、`license`、`attribution`、`keywords`、`dimensions` 和模板推荐。与继续扩展 Python 字典相比，manifest 可被校验、文档和生成器共同消费。

2. **背景主题使用独立 manifest。** 新增 `background_templates/manifest.json`，主题层记录场景、优先级和推荐 palette，图片层记录相对路径及来源。`background_manager.py` 优先读 manifest，缺失时回退到旧目录扫描，以便部署中可平滑升级。

3. **未知语义默认省略图标。** `icon_id_from_text` 默认返回空字符串；章节、金句、总结等结构页使用 `section`、`quote`、`thanks`、`check` 等页面语义明确的回退。`add_local_icon` 找不到素材时返回 `None`，不再画缩写 badge。这样“少一个图标”优于“出现错误含义的图标”。

4. **外部素材通过同步脚本落地。** `scripts/sync_external_assets.py` 保存精选 URL、下载并归一化为 512x512 透明 PNG 图标、1920x1080 JPG 背景和 1920x1080 PNG pattern，同时生成两个 manifest。运行时不依赖网络。

5. **来源与许可可验证。** 外部素材必须存在 `source_id`、`source_url`、`license` 和 `attribution`；项目既有背景可标记为 `project-existing`。Bing 搜索 URL记录为发现入口，不作为图片许可来源。

6. **内容图片与页面背景分离。** manifest v2 新增 `photo` 类型，路径按 `assets/photos/<category>/` 分类。`image_text` 优先使用显式有效的 `image_path`，其次使用任务背景作为图片区候选，最后按标题和正文语义选择本地默认图。缺图时仍输出真正的 PowerPoint 图片对象，便于用户直接更换；不得用图标、纹理面板或内部实现文案伪装图片。

## Risks / Trade-offs

- [Icons8 免费使用通常要求署名] -> 在 manifest、README 和生成器文档中保留 `Icons by Icons8` attribution；商业部署前由项目所有者确认所用 Icons8 许可层级。
- [外部 CDN URL 可能失效] -> 仓库提交归一化后的离线文件；同步脚本采用固定 URL 并进行尺寸/格式校验，失效时明确失败而不写半文件。
- [照片增加仓库体积] -> 背景统一裁剪到 1920x1080 JPEG、质量 88-90，单文件目标小于 900 KB。
- [manifest 关键字匹配仍可能冲突] -> 使用最长关键字、显式 priority 和直接 id 优先，补充聚焦单元测试。
- [既有未注明来源背景无法自动补齐许可] -> 在背景 manifest 标记为 `project-existing`，不伪造外部来源；后续可逐主题替换。

## Migration Plan

1. 运行同步脚本替换 `assets` 三类文件并补充图片不足的背景主题。
2. 提交 v2 assets manifest 和 background templates manifest。
3. 更新 asset/background manager 及结构页 fallback，运行 manifest 与语义测试。
4. 生成代表性封面、章节、图标网格、图文和总结页，渲染检查图标语义、背景裁切和重叠。
5. 部署时整目录同步 `skills/visual_designer`，避免只更新 Python 而遗漏二进制素材。

## Open Questions

- Icons8 在目标商业场景中的最终许可层级由项目所有者确认；当前按需要署名的免费许可保留 attribution。
- `project-existing` 背景的原始来源需在后续素材治理中补录或替换。
