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
- 需要使用 Unsplash 图片 CLI 或部署认证时读取 `references/unsplash-cli.md`。
- 需要可执行示例时读取 `examples/README.md`，再按需查看对应示例目录。
- 只有在接入本仓库 `ppt-agent` Go 后端时，才读取 `references/ppt-agent-integration.md`。

## 执行优先级与素材预检

- 在本 skill 目录或 `ppt-agent` 项目中处理 PPT 请求时，本 skill 负责 DeckSpec、素材路线、校验和渲染；先读取本文件并完成素材预检，再调用通用 PPT 或图像生成能力。
- **默认先搜索图片，且不得因页面可由文本、图表或卡片表达而跳过。** 除用户明确要求纯文字/无图片、页面被明确标记为 `clean_text_only`，或图片服务实际返回可报告的配置、鉴权、网络或无结果错误外，不得无图降级。
- 默认素材路线按宿主区分：**通用 Agent**（Codex、Claude Code、OpenCode、Pi 等）使用本 skill 的 Unsplash CLI；**`ppt-agent` 项目 Agent** 中，闲聊 Agent 使用 `search_images` Tool 向用户展示候选图，PPTPlanner 不注册图片 Tool，只提交视觉意图，由后端确定性素材物化层在审查后搜索、下载并回填。成功物化后应将真实 `local_path`、`source_url`、`attribution`、`provider` 与 `search_status` 写入 `visual_intent` 或 `image` 组件。
- **背景复用规则：同一任务内，所有相同 `content_type` 的页面必须使用同一张已物化背景图。** 规划时先为该类型确定一个共享的背景 `asset_query`；下载后这些页面的背景 `local_path`、来源和署名必须一致。不同 `content_type` 才可使用不同背景，以形成有节制的版式节奏。
- 仅在上述豁免或实际失败发生后，文本、图表、流程和浅色面板才可作为无图页的交付方案；必须在任务记录或最终说明中写明豁免理由或失败类型，禁止静默 fallback。
- 不要为了绕过搜图认证、下载失败或素材缺失而自动改用通用 AI 图像生成。只有用户明确要求 AI 生成配图，或明确同意切换该路线时，才可以使用它。
- 若用户要求“配图不可用即停止”，任一选定素材路线失败后都必须在渲染前停止，并说明失败的工具或服务、失败类型，以及可用的替代路线；不得静默降级或切换路线。

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

图片默认是必需的视觉素材，而不是可选增强。只有“执行优先级与素材预检”中列明的豁免或实际素材失败，才可使用无图页；一旦规划图片，则 `local_path`、`image_path`、`asset_path` 必须指向真实可读的本地文件，禁止虚构路径或旧 `asset:` id。经图片服务解析的背景还必须保留实际 `provider` 与 `search_status:"resolved"`/`"downloaded"`，以便渲染前校验其来源链路。具体素材路线遵循上面的“执行优先级与素材预检”。

背景不是按页轮换的装饰素材：每个 `content_type` 在一套 deck 内只选定一张背景。若同类型页已有不同背景路径，必须在渲染前统一为该类型的一个已解析背景；`validate_deck.py` 会拒绝这种不一致的任务清单。

需要由 Agent 搜图时，按以下边界执行：

- **通用 Agent**：必须使用本 skill `scripts/` 下的 Unsplash CLI，例如 Codex、Claude Code、OpenCode、Pi 等。CLI 负责搜索并下载可用于本次 DeckSpec 的本地素材。
- **`ppt-agent` 闲聊 Agent**：使用项目内 `search_images` Tool 搜索 Unsplash 候选，并将预览、来源页和摄影师署名回显给用户；不要要求闲聊 Agent 调用 skill CLI，也不要下载到 PPT 任务目录。
- **`ppt-agent` PPTPlanner / Reviewer / Fixer**：不调用 `search_images` Tool，也不调用 CLI；它们只规划或修正 `visual_intent`。后端确定性素材物化层负责后续下载和字段回填。

独立 Agent 首次使用时，在 skill 根目录以 Node.js 22+ 执行一次 `npm link`，即可注册本机 `unsplash` 命令。需要图片时，在 `content_plan.visual_intent` 中规划 `asset_purpose`、`asset_query`、`asset_subject`、`composition` 与 `orientation`：

```bash
npm link
unsplash auth
unsplash fetch --work-dir <work-dir>
```

`unsplash auth` 会显示 `accessToken（输入后不会回显）:`；粘贴或输入 Access Key 后按 Enter，输入字符不可见是正常的保护行为。服务器部署可在已加载环境变量后执行 `unsplash auth --from-env`，它读取 `UNSPLASH_ACCESS_KEY`（兼容 `UNSPLASH_ACCESS_TOKEN`）且不会回显密钥。认证信息保存到 skill 根目录、已被 Git 忽略的 `auth.txt`。Access Key 在 Unsplash Developers 控制台创建应用后可获得；不要写入 `tasks.json`、日志、prompt、命令参数或仓库文件。下载成功后脚本会回写 `local_path`、`source_url`、`attribution` 和图片元数据。

## 校验与渲染入口

渲染前运行确定性预检：

```bash
python generators/validate_deck.py --work-dir <work-dir> --skills-dir <skills-dir>
```

整套生成优先使用 `generators/render_deck.py`：

```bash
python generators/render_deck.py --work-dir <work-dir> --skills-dir <skills-dir> --output deck.pptx
```

默认执行顺序为：`视觉规划 → 图片搜索与下载 → validate_deck.py → render_deck.py → PDF/PNG 视觉验收`。无图页只能在已记录豁免或搜索失败后执行预检和渲染。

需要调试单页时使用 `generators/render_task.py`：

```bash
python generators/render_task.py --work-dir <work-dir> --skills-dir <skills-dir> --task-id <task-id>
```

`render_task.py` 会读取 `<work-dir>/tasks.json`，找到指定 `task_id` 对应页面，并输出该页 PPTX。若宿主系统没有 `task_id`，应在调用前由适配层按页码派生。

## 变更纪律

- 常规页面生成必须复用 `generators` 包、`validate_deck.py`、`render_deck.py` 或 `render_task.py`，不要让 Agent 手写底层 `python-pptx` 绘图代码。
- 修改页面类型、组件容量或字段契约时，同步 `templates/component_contracts.json`、`references/slide_types.md` 和相关 generator。
- 修改 generator 函数签名时，同步 `references/generators.md`、`generators/__init__.py`、`generators/render_task.py` 和对应测试。
