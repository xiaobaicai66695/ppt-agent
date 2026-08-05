## Why

当前 Visual Designer 的离线素材库仅覆盖少量工程流程概念，背景主题数量和单主题图片数量也不均衡，导致生成器频繁回退到 `layout`、`primitive` 等无关图标并重复使用同一背景。需要建立一套语义覆盖更广、来源可追踪、可由生成器稳定选择的本地素材契约，降低 PPT 的占位感和风格漂移。

## What Changes

- 全量替换 `skills/visual_designer/assets` 中现有的图标、编辑型背景和纹理素材，统一视觉风格、尺寸、透明通道和离线加载方式。
- 扩充 `background_templates` 的六个主题，使每个主题具备多张可轮换图片，并记录来源、许可和适用场景。
- 扩展素材 manifest 的语义标签、别名、来源和许可字段，增加素材完整性校验。
- 新增可按内容语义选择的 `photo` 素材类型和分类目录，供图文解说页在没有显式图片时放置可替换的默认图片。
- 调整图标语义匹配、主题选择和 fallback 规则，未知语义不再默认渲染无关工程图标。
- 更新 Visual Designer 的 README/SKILL、背景模板说明和 generator 参考文档，明确目录结构、素材契约、选择规则及维护流程。

## Capabilities

### New Capabilities

- `visual-designer-asset-library`: 定义本地图标、背景和纹理资产的目录、元数据、语义选择、许可追踪及生成器消费契约。

### Modified Capabilities


## Impact

- 影响 `ppt-agent/skills/visual_designer/assets`、`background_templates`、`generators/asset_manager.py`、使用本地图标或背景的页面生成器及相关文档。
- 不改变对外 HTTP API；`image_text` 只新增可选的 `image_path` 参数，现有调用保持兼容。
- 仓库体积会因离线图片增加而上升，需要通过尺寸、格式和单文件大小约束控制。
