# 2026-08-27 PPT Agent 流式文本与视觉色板修复

## 背景

- 用户反馈主会话中英文流式文本单词被拼接，且模型可见输出应默认使用中文。
- 用户反馈 `image_text` 页面正文过短却占用大文本框，版式长期固定为左图右文，页面显得空洞。
- 背景图片上线后和固定主题色容易冲突，尤其是政府红等强色系会压过背景图和正文可读性。

## 行为变化

- 流式 assistant 内容合并时保留 ASCII 单词边界，兼容上游返回“增量 chunk”和“累计快照”两种形式。
- Planner、Reviewer、Fixer 的用户可见自然语言和页面正文默认要求中文，专有名词、英文缩写、URL 和代码字段除外。
- `image_text` 从三种变体扩展为四种：`image_left`、`image_right`、`image_top_band`、`image_bottom_band`；未显式指定时按页序自动轮换。
- `image_text` 规划质量门增加正文密度检查，少于 240 字时给出 `low_information_density` 警告，避免大面积空框。
- 背景图渲染改为确定性处理：降饱和、降对比、提高白色柔化层，并从背景位图提取弱化色系 token 覆盖固定主题色。
- 意图路由和用户画像不再向 Planner 注入历史主题色/固定推荐主题；顶层 theme 只作为无背景页和旧接口兜底。

## 本地验证

- `D:\anaconda\python.exe -m unittest tests.test_render_task_components`
- `python -m json.tool templates\component_contracts.json`
- `go test ./pkg/web ./pkg/task ./pkg/agent/deck ./pkg/prompts`
- `go build ./...`
- `npm test -- --run src/utils/workbench.test.ts`
- `npm run build`

## 上线记录

- 目标：`remote-dev:/ppt/ppt-agent`
- 时间：2026-08-27 00:22 Asia/Shanghai
- 新进程：PID `3603484`，命令 `../ppt-agent-linux -mode web -addr :8080`，cwd `/ppt/ppt-agent/backend`
- 备份：
  - 二进制：`/ppt/ppt-agent/ppt-agent-linux.bak.20260827002200-visualpalette`
  - 前端：`/ppt/ppt-agent/frontend/dist.bak.20260827002200-visualpalette`
- 启动确认：
  - `/api/health` 返回 HTTP 200，`{"status":"ok"}`
  - `/` 返回 HTTP 200，709 bytes
  - `/api/templates/layouts` 返回 HTTP 200，17522 bytes，包含 `image_bottom_band`
  - `/api/themes` 返回 HTTP 200，2614 bytes
  - `:8080` 正常监听，进程为 `3603484`
- 线上生成器 smoke：
  - 在 `/tmp/ppt-agent-visualpalette-smoke` 创建 4 个临时 `image_text` 任务，不调用模型。
  - 4 个任务均由 `/ppt/ppt-agent/skills/ppt-deck-planner/generators/render_task.py` 成功渲染，输出 `slide_1.pptx` 至 `slide_4.pptx`，单文件约 32K。
  - 临时 workdir、接口响应文件和传输包已清理。

## 遗留

- 本次没有发起完整 LLM 生成任务，避免额外消耗用户或系统上游 Key；已通过确定性渲染 smoke 覆盖生成器链路。
- 若后续仍出现背景局部复杂导致文字难读，可继续增加“文字安全区亮度/复杂度”检测，而不是依赖 Planner 视觉描述。

## 背景搜索关键词与同类型复用修正

- 问题：
  - 背景图原策略把整套 PPT 收敛为两张浅色背景并按页轮换，容易让同一页面类型出现不同背景节奏。
  - Planner 仍倾向生成 `wide landscape clean negative space` 一类长搜索词，Unsplash 可能被长尾描述带偏，搜到和主题不相关的图。
- 行为变化：
  - 背景物化从“两张全局背景槽”改为“按 `content_type` 分组”：同一页面类型复用同一张背景图，不同页面类型可以使用不同背景。
  - 背景搜索词优先使用 `asset_subject`，否则从 `asset_query` 中提取第一个有效关键词；去掉 `wide landscape`、`clean negative space`、`light/bright/airy` 等构图和明暗词。
  - Reviewer 增加两类 warning：同一 `content_type` 出现多个背景关键词、背景 `asset_query` 过长。
  - Planner / Reviewer prompt、Skill 和组件契约同步改为“背景 `asset_query` 尽量只写一个英文关键词”。
- 本地验证：
  - `python -m json.tool templates\component_contracts.json`
  - `go test ./pkg/agent/deck ./pkg/prompts`
  - `go test ./pkg/web ./pkg/task ./pkg/agent/deck ./pkg/prompts`
  - `go build ./...`
- 上线记录：
  - 目标：`remote-dev:/ppt/ppt-agent`
  - 时间：2026-08-27 01:38 Asia/Shanghai
  - 新进程：PID `3624298`，命令 `../ppt-agent-linux -mode web -addr :8080`
  - 部署基线：代码提交 `3dbc750`，同批发布分片规划、上下文压缩可见性和背景关键词复用修正。
  - 启动确认：
    - `/api/health` 返回 HTTP 200，`{"status":"ok"}`
    - `/` 返回 HTTP 200，709 bytes
    - `/api/templates/layouts` 返回 HTTP 200，17522 bytes，包含 `image_bottom_band`
    - `/api/themes` 返回 HTTP 200，2614 bytes
    - `:8080` 正常监听，进程为 `3624298`
    - `ppt-agent-linux` SHA256 为 `f00ee08e341c4c118548b7d77bc82c7281f4198c01c6f17129339885ba7d7d71`
  - 说明：本次没有再次发起完整 LLM 生成任务，避免额外消耗上游 Key；通过接口 smoke、layout contract 标记和二进制探针确认运行包已包含新逻辑。

## 分片规划与上下文压缩可见性

- 问题：
  - 12 页以上 DeckSpec 由单个 Planner 一次性填完整 `tasks.json`，后半段容易退化为抽象套话，且不利于后续单页/单节修复。
  - 上下文压缩阈值过高，实际规划调用接近窗口上限时不容易触发；触发后也缺少前端明确回显。
- OpenSpec：
  - `openspec/changes/archive/2026-08-26-chunked-deck-planning-and-context-compression/`
- 行为变化：
  - 新增 chunked planning 路径：BlueprintPlanner 先锁定页码、章节、标题、页面类型和 `content_bank`，SectionPlanner 按小节补全 2-4 页内容，后端 Merger 确定性合并后仍只走一次 Task Reviewer。
  - `tasks.json` 页面增加可选 `section_id`、`section_title`、`page_intent`、`evidence_refs`，`content_plan` 增加可选 `evidence_refs`，为后续定点修复提供粒度。
  - Merger 负责填充 `task_id`、`output_file`、`status`、`created_at`、`capacity_hint.component_count`，并按 `content_type` 收敛背景关键词，减少 LLM 填工程字段的负担。
  - 上下文压缩 prompt 显式包含历史消息、上次压缩交接和当前用户最新问题；压缩开始时记录 `compressing_context` 阶段并通过 SSE progress/runtime_meta 回显。
  - 默认规划压缩阈值从 200000 估算 token 调整为 24000，可通过 `PLANNER_COMPRESSOR_TOKEN_THRESHOLD` 覆盖。
- 本地验证：
  - `go test ./pkg/agent/deck ./pkg/agent/utils ./pkg/task`
  - `go build ./...`
  - `npm test -- --run src/utils/workbench.test.ts`
  - `npm run build`
  - `go test ./...` 已执行，但 `test/plan_benchmark` 的既有 gold DeckSpec 因背景关键词/旧金标组件容量规则失败，需单独更新 gold 或调整评测基线。
- 上线记录：
  - 目标：`remote-dev:/ppt/ppt-agent`
  - 时间：2026-08-27 01:38 Asia/Shanghai
  - 新进程：PID `3624298`，命令 `../ppt-agent-linux -mode web -addr :8080`
  - 部署基线：代码提交 `3dbc750`
  - 启动确认：
    - `/api/health` 返回 HTTP 200，`{"status":"ok"}`
    - `/` 返回 HTTP 200，709 bytes
    - `/api/templates/layouts` 返回 HTTP 200，17522 bytes，包含 `image_bottom_band`
    - `/api/themes` 返回 HTTP 200，2614 bytes
    - `:8080` 正常监听，进程为 `3624298`
    - `ppt-agent-linux` 二进制探针包含 `compressing_context`
  - 清理：远端 `/tmp/ppt-home-smoke.html`、`/tmp/ppt-layouts-smoke.json`、`/tmp/ppt-themes-smoke.json`、`/tmp/ppt-agent-linux-deploy`、`/tmp/ppt-agent-frontend-dist-3dbc750.tar.gz`、`/tmp/ppt-agent-skill-3dbc750.tar.gz` 已删除。
  - 备份：二进制备份为 `ppt-agent-linux.bak.20260827013702-chunked`；前端可回退备份沿用上一轮 `frontend/dist.bak.20260827004200-bgvisible`。
  - 遗留：未发起完整 LLM 生成任务，避免额外消耗上游 Key；分片 Planner 的端到端模型质量仍建议后续用 1 个低页数任务或专门评测集补测。

## 背景取色文本可读性修正

- 问题：
  - 背景图偏黄时，生成器把背景提取色直接写入 `primary` / `secondary` / `accent` 文本 token，导致 kicker、副标题、图片说明等文字变成浅黄，和浅色背景/玻璃面板低对比。
  - 同时，矩形、面板和分割线仍主要使用固定主题色，出现“文字跟背景走，色块不跟背景走”的反向效果。
- 行为变化：
  - `background_image_palette()` 将文本 token 和形状 token 分离：正文、标题、小标题、图片说明保持深色/主题可读文本色；背景派生色只通过 `primary_fill`、`secondary_fill`、`accent_fill` 供面板、色块、分割线和强调装饰使用。
  - 组件渲染在带背景图时临时创建 runtime shape palette，让形状使用背景派生色，同时所有 `add_text(..., colors=colors)` 仍读取可读文本 token。
  - Skill、generator 参考文档和组件契约同步改为“背景取色用于容器和装饰，不用于浅色文字”。
- 本地验证：
  - `D:\anaconda\python.exe -m unittest tests.test_render_task_components`
  - `Get-ChildItem -LiteralPath generators -Filter *.py | ForEach-Object { D:\anaconda\python.exe -m py_compile $_.FullName }`
  - `D:\anaconda\python.exe -m json.tool templates\component_contracts.json`
  - 本地构造黄底 `image_text` smoke，经 `render_task.py` 生成 PPTX，LibreOffice 转 PDF，Poppler 渲染 PNG；检查到 XML 同时包含可读文本色 `17202A` / `51616D` 和背景派生 shape 色 `B8CC6C`。
- 上线记录：
  - 目标：`remote-dev:/ppt/ppt-agent/skills/ppt-deck-planner`
  - 时间：2026-08-27 10:39 Asia/Shanghai
  - 部署基线：代码提交 `44135bf`
  - 后端进程：PID `3624298` 保持运行，未重启；`/api/health` 返回 HTTP 200，`{"status":"ok"}`
  - 备份：`/ppt/ppt-agent/skills.bak.20260827103809-palette-text`
  - 线上生成器 smoke：
    - 在 `/tmp/ppt-agent-palette-text-smoke` 创建 1 页黄底 `image_text` 临时任务，不调用模型。
    - `/ppt/ppt-agent/skills/ppt-deck-planner/generators/render_task.py` 成功生成 `slide_01.pptx`，83,300 bytes。
    - PPTX XML 包含可读文本色 `17202A`、`51616D` 和背景派生 shape 色 `B8CC6C`，临时 workdir 和传输包已清理。
