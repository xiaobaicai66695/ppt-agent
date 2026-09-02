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
- 独立 Agent 需要使用 Unsplash 图片 CLI 时读取 `references/unsplash-cli.md`。
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
- 内容容量要匹配布局：`templates/component_contracts.json` 的 `planning_rules.component_text_density`、`planning_rules.page_content_density` 与每个 `content_type.capacity` 是**唯一的数字来源**。规划前读取它们，按组件类型和页面类型同时核算字数、列表项和组件数；空间不够时拆页、合并次要组件或切换页面类型，不能依赖渲染器截断文字。
- 通用 Agent 必须把这些容量规则用于首次 DeckSpec，而不只在渲染报错后补救：卡片/段落/列表/完整论述按组件目标字数写，信息页、图文页和双栏页再满足页面总信息量与结构要求。标题、图片元数据、来源行和装饰组件不计入正文容量。
- `agenda` 的每个 `toc_item` 都要同时写章节标题和 18-44 字的 `body` 副标题，副标题说明该章要回答的问题、观察角度或听众收获，不能只重复标题。目录超过 3 项时会以双列章节卡片呈现，不要把整页摘要塞进某一个目录项。
- `section_divider` 固定使用 `background_title`：背景图上只呈现主标题和副标题；`section_marker` 如需保留仅作顺序元数据，不能规划为左侧色块或大号可见编号。
- `image_text` 的完整论述只放在正文面板；可选 `lede` 只写一句独立引导，不能复制正文开头。正文已按图文阅读距离采用更高字号，若内容超出契约容量应拆页而不是降低字号。
- 用户只给大纲时，也要扩写成可上屏内容；不要留下 `{要点1}`、`某公司`、`若干数据`、`核心观点` 这类占位词。
- 需要真实数据时，保留来源机构、时间和 URL；不要让生成器或渲染脚本承担事实补全。

## 图片与素材策略

视觉策略由 DeckSpec 顶层 `visual_policy` 显式声明，是 Planner、素材解析器和渲染器共享的唯一契约。除非用户明确要求纯文字/无图片，新 deck 必须设置 `mode="required"` 和 `required_roles`，并为**每一页**规划并物化背景或前景图片；`min_image_pages` 是覆盖率一致性字段，必须不小于非豁免页数，不能被写成“只给少数关键页配图”的目标。只有在页面确实应保持纯文字且已明确说明原因时，才可在该页写 `content_plan.visual_intent = {"role":"clean_text_only","search_status":"skipped","skip_reason":"..."}` 作为豁免。纯文字 deck 必须设置 `mode="none"` 并说明 `reason`。不要用缺失字段、低 `min_image_pages` 或未落地的 query 暗示“无图”。

Planner 只负责可执行的图片语义：背景图写入 `content_plan.visual_intent`；页内实景或证据图写成 `content_plan.components` 中的 `type="image"`。两种声明都应包含 `asset_purpose`、英文 `asset_query`、`asset_subject`、`composition` 和 `orientation`。一旦规划图片，则 `local_path`、`image_path`、`asset_path` 必须在渲染前指向真实可读的本地文件，禁止虚构路径或旧 `asset:` id。

`scripts/` 下的 Unsplash CLI 只供 **ppt-agent 项目外的独立 Agent** 使用。项目内 Agent 必须使用后端既有素材搜索与下载链路，不能读取 skill 根目录的 `auth.txt` 或调用该 CLI。

独立 Agent 首次使用时，在 skill 根目录以 Node.js 22+ 执行一次 `npm link`，即可注册本机 `unsplash` 命令。解析器会处理背景 `visual_intent` 和前景 `image` 组件；先完成完整 DeckSpec，再运行：

```bash
npm link
unsplash auth
unsplash fetch --work-dir <work-dir>
```

`unsplash auth` 会显示 `accessToken（输入后不会回显）:`；粘贴或输入 Access Key 后按 Enter，输入字符不可见是正常的保护行为。认证信息保存到 skill 根目录、已被 Git 忽略的 `auth.txt`。Access Key 在 Unsplash Developers 控制台创建应用后可获得；不要写入 `tasks.json`、日志、prompt、命令参数或仓库文件。下载成功后脚本会回写 `local_path`、`source_url`、`attribution` 和图片元数据。

## 校验与渲染入口

渲染前运行确定性预检：

```bash
python generators/validate_deck.py --work-dir <work-dir> --skills-dir <skills-dir>
```

整套生成优先使用 `generators/render_deck.py`：

```bash
python generators/render_deck.py --work-dir <work-dir> --skills-dir <skills-dir> --output deck.pptx
```

独立 Agent 的固定顺序为：`视觉规划（含逐页 visual_intent） → unsplash auth（首次） → unsplash fetch → validate_deck.py → render_deck.py → PDF/PNG 视觉验收`。`required` 策略下，遗漏下载、只写 query、少配图片或仅用低 `min_image_pages` 掩盖缺图都会被预检和整套渲染入口拒绝；`none` 策略才可跳过素材解析。

需要调试单页时使用 `generators/render_task.py`：

```bash
python generators/render_task.py --work-dir <work-dir> --skills-dir <skills-dir> --task-id <task-id>
```

`render_task.py` 会读取 `<work-dir>/tasks.json`，找到指定 `task_id` 对应页面，并输出该页 PPTX。若宿主系统没有 `task_id`，应在调用前由适配层按页码派生。

## 变更纪律

- 常规页面生成必须复用 `generators` 包、`validate_deck.py`、`render_deck.py` 或 `render_task.py`，不要让 Agent 手写底层 `python-pptx` 绘图代码。
- 修改页面类型、组件容量或字段契约时，同步 `templates/component_contracts.json`、`references/slide_types.md` 和相关 generator。
- 修改 generator 函数签名时，同步 `references/generators.md`、`generators/__init__.py`、`generators/render_task.py` 和对应测试。
