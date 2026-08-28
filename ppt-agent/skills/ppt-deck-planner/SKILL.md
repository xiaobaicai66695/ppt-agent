---
name: ppt-deck-planner
description: 规划可执行的 PPT DeckSpec/tasks.json，并用组件化 Python generators 校验和渲染 PPTX。
---

# PPT Deck Planner

本 skill 帮助 Agent 把用户需求规划为可执行的 `tasks.json`，再通过内置 Python generators 生成 PPTX。它是一个独立能力包，不依赖特定后端框架、离线素材库或固定任务目录。

## 什么时候使用

- 用户要求创建或修改一套 PPT、PPTX、演示文稿或 slide deck。
- 需要把自然语言需求、大纲、研究材料或数据转成结构化 DeckSpec。
- 需要根据已有 `tasks.json` 渲染 PPTX，或定位某页内容/布局契约问题。

## 先读哪些资源

- 规划页面时读取 `templates/component_contracts.json` 和 `references/slide_types.md`。
- 调用生成器或排查渲染问题时读取 `references/generators.md`。
- 验证独立 skill 可用性时读取 `references/standalone-validation.md`。
- 需要可执行示例时读取 `examples/README.md`，再按需查看对应示例目录。
- 只有在接入本仓库 `ppt-agent` Go 后端时，才读取 `references/ppt-agent-integration.md`。

## DeckSpec 规划边界

Planner 应尽力一次性填写完整 DeckSpec 内容字段，包括：

- deck 级标题、受众、目标、章节或内容素材；
- 每页 `page_index`、`title`、合法 `content_type`、可选 `section_id/section_title/layout_variant/page_intent/evidence_refs`；
- `content_plan.summary`、`content_plan.slide_intent`、`content_plan.components`；
- 需要事实、数据或图片时的 `source`、图片语义、署名和可追溯信息。

不要把运行态或系统派生字段当成 Planner 内容：

- `task_id`
- `output_file`
- `status`
- `qa_report`
- `fix_attempts`
- `capacity_hint`
- `reviewer_status`

这些字段可由调用方、适配层或任务系统补齐。

## 页面与组件契约

- `content_type` 只能使用 `references/slide_types.md` 和 `templates/component_contracts.json` 中声明的英文 id。
- `layout_variant` 只能使用对应 `content_type` 明确支持的值；未知时留空，让生成器自适应。
- `content_plan.components` 是页面正文和视觉语义的主入口。组件只表达语义，不写坐标、字号、颜色、透明度、阴影、边距或卡片尺寸。
- 历史别名 `bar_chart`、`line_chart`、`pie_chart`、`doughnut_chart`、`table` 不能写入新的 `tasks.json`；趋势、占比和对比数据统一用 `chart_slide`，表格对比用 `comparison_table`。

## 内容规划原则

- 整套 PPT 优先采用“观点 → 论据 → 推论/行动”的叙事结构，而不是连续罗列要点。
- 每页只讲一个中心判断；多个判断并列时拆页或改成目录/框架页。
- 内容容量要匹配布局：卡片、列表、KPI、图表、长论述分别遵守 `component_contracts.json` 中的容量上限。
- 用户只给大纲时，也要扩写成可上屏内容；不要留下 `{要点1}`、`某公司`、`若干数据`、`核心观点` 这类占位词。
- 需要真实数据时，保留来源机构、时间和 URL；不要让生成器或渲染脚本承担事实补全。

## 图片与素材策略

本 skill 不捆绑 `assets/` 目录，也不维护 `assets/manifest.json`。通用 Agent 可以使用自身联网能力搜索并下载网络图片；skill 不规定下载工具、图片来源或保存目录，只要求渲染前把真实可读的本地路径写入 DeckSpec。

- 要让图片进入 PPT，传入 `visual_intent.local_path`、`image.local_path`、`image_path` 或等价字段。
- 图片保存在哪里由调用方或宿主 Agent 决定；skill 文档不规定任务目录结构。
- 可以保留 `asset_query`、`asset_subject` 等语义字段供外部素材流程使用，但 Python generators 不直接联网搜索，也不从离线 manifest 解析 `asset:` id。
- 不需要图片的页面应直接规划为文本、卡片、图表或浅色面板；一旦写入 `local_path`、`image_path`、`asset_path` 或旧 `asset:` id，就必须指向真实可读的本地文件，否则渲染应失败并暴露问题。

## 校验与渲染入口

渲染前先运行确定性预检：

```bash
python generators/validate_deck.py --work-dir <work-dir> --skills-dir <skills-dir>
```

整套生成优先使用 `generators/render_deck.py`：

```bash
python generators/render_deck.py --work-dir <work-dir> --skills-dir <skills-dir> --output deck.pptx
```

需要调试单页时使用 `generators/render_task.py`：

```bash
python generators/render_task.py --work-dir <work-dir> --skills-dir <skills-dir> --task-id <task-id>
```

`render_task.py` 会读取 `<work-dir>/tasks.json`，找到指定 `task_id` 对应页面，并输出该页 PPTX。若宿主系统没有 `task_id`，应在调用前由适配层按页码派生。

## 变更纪律

- 常规页面生成必须复用 `generators` 包、`validate_deck.py`、`render_deck.py` 或 `render_task.py`，不要让 Agent 手写底层 `python-pptx` 绘图代码。
- 修改页面类型、组件容量或字段契约时，同步 `templates/component_contracts.json`、`references/slide_types.md` 和相关 generator。
- 修改 generator 函数签名时，同步 `references/generators.md`、`generators/__init__.py`、`generators/render_task.py` 和对应测试。
