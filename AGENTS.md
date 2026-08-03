# AGENTS.md

本文件用于指导 Codex 在当前 PPT Agent 工作区中协作、阅读代码、实施变更和选择验证方式。

## 项目背景

- 工作区根目录：`D:\environment\codeGo\llm-examples\projects`。
- 主要项目：`ppt-agent`，一个基于 CloudWeGo Eino ADK 的多 Agent PPT 生成系统。
- 辅助项目："D:\environment\codeGo\llm-examples\eino-examples"，里面有eino框架的接口使用示例。
- harness工程的辅助设计文档"D:\桌面\个人信息\agent相关学习\深入理解-AI-Agent-李博杰-v1.2(1).pdf"
- 不要套用旧 ITSM 项目的目录、技术栈或验证命令；当前工作区以 PPT 生成、Go 后端、Vue 前端和 Python PPT 生成器为核心。

## 语言规则

- 用户使用中文时，默认用中文回复；代码、API 名称、命令和原始错误信息可保留英文。
- 代码注释、prompt、模板说明和用户可见文案中的既有中文术语应尽量保持一致。
- 最终说明应简洁说明改了什么、改在哪里、如何验证；无法验证时说明具体原因。

## 仓库边界

- 根目录包含：
  - `ppt-agent/backend`：Go 后端服务、Agent 编排、Web API、任务管理、模型 fallback、prompt 模板。
  - `ppt-agent/frontend`：Vue 3 + TypeScript + Vite 前端。
  - `ppt-agent/skills/visual_designer`：PPT 视觉设计 skill、Python `python-pptx` 生成器、模板、背景资源和参考文档。
  - `ppt-agent/doc`、`ppt-agent/test`、`ppt-agent/scripts`：项目文档、测试和辅助脚本。
  - `ppt-plugin/skills`：插件/skill 分发相关内容。
- 修改代码前，应在工作区根目录或相关子项目查看 `git status`。不要回滚与当前任务无关的用户改动。
- 除非任务明确涉及，不要编辑或清理：
  - `node_modules`
  - `__pycache__`
  - `dist`
  - `output`、`weboutput`
  - `examples_output`
  - `.exe`、`.zip`、生成的 `.pptx`、`.pdf`、`.jpg`
  - 日志、缓存、临时脚本、凭据和本地 `.env`
- 当前仓库可能已经跟踪了部分构建产物或依赖目录；不要在无关任务中扩大这些 diff。需要治理时单独处理。

## 后端开发约定

- 后端技术栈：Go、Gin、CloudWeGo Eino ADK、MySQL/本地任务状态、SSE 流式事件。
- 入口：`ppt-agent/backend/main.go`。
- 主要模块：
  - `pkg/agent/deep`：DeepAgent 主流程、SlideExecutor、Reviewer、Fixer、任务清单类型。
  - `pkg/prompts/deep`：主 Agent、单页执行器、审查器、修复器 prompt 模板。
  - `pkg/web`：HTTP API、任务创建、模板接口、文件下载、缩略图生成。
  - `pkg/task`：任务生命周期管理。
  - `pkg/templates`：模板和背景资源加载。
  - `pkg/agent/utils`：模型 fallback、上下文压缩、token 统计。
- 修改 Agent/prompt 逻辑时要特别注意：
  - `tasks.json` 是多 Agent 协作的核心契约，避免并发覆盖整个文件。
  - `content_type` 必须保持为生成器支持的稳定英文 id，不要写成中文显示名。
  - `description`、`content_plan`、`background`、`theme` 的含义要和前端、模板 loader、Python 生成器保持一致。
  - prompt 中不要加入互相冲突的指令，例如一处禁止某工具、一处又要求使用该工具验证。
  - 大纲模式如果跳过规划阶段，必须确保 outline 已经补齐并通过 schema 校验。
- 处理模型质量问题时，不要只调温度或换模型；优先检查结构化输入、prompt 契约、fallback 后能力差异、上下文压缩是否丢失关键信息。

## 前端开发约定

- 前端技术栈：Vue 3、Composition API、TypeScript、Vite。
- 主要文件：
  - `ppt-agent/frontend/src/api.ts`：API client 和数据类型。
  - `ppt-agent/frontend/src/pages/ComposePage.vue`：模板编排页。
  - `ppt-agent/frontend/src/pages/DashboardPage.vue`：任务列表、进度、结果展示和预览。
  - `ppt-agent/frontend/src/components/SlidePreviewCard.vue`：单页缩略图卡片。
  - `ppt-agent/frontend/src/App.vue`：全局样式和 design tokens。
- UI 应按工作台/生产工具设计：信息密度适中、状态清晰、操作路径明确，避免营销页式大卡片和纯装饰背景。
- 不要用 emoji 或纯渐变代替真实模板缩略图；模板接口中有 `thumbnail` 时优先使用真实资源。
- 预览状态必须区分“生成中”“缺少缩略图”“转换失败”“文件不存在”，不要把所有错误都显示为仍在生成。
- CSS token 避免自引用变量，例如不要写 `--border: var(--border)`。
- 主要操作文案应和业务一致，例如 PPT 生成按钮不要写成“开始翻译”。
- 大页面改动时优先拆分可复用组件、composable 或局部工具函数，避免继续膨胀单文件页面。

## PPT 生成器与视觉质量约定

- 核心目录：`ppt-agent/skills/visual_designer`。
- 修改 Python 生成器前，应先阅读：
  - `skills/visual_designer/SKILL.md`
  - `skills/visual_designer/references/generators.md`
  - 相关 `templates/single-page/*.json`
  - 相关 `generators/*_generator.py`
- 所有单页 PPT 生成应复用 `generators` 包和 `base.py` helper，避免在 Agent 生成代码中手写底层 `python-pptx` 绘制逻辑。
- 生成器质量优先级：
  - 文字不溢出、不重叠、不被背景压低对比度。
  - 字号、行距、留白、卡片尺寸和图表区域有稳定规则。
  - `description` 和 `content_plan` 的内容密度要匹配布局容量。
  - 背景图只在适合的页面使用；不要强制所有页面都使用图片背景。
  - 背景主题和 palette 要协调，必要时通过 `get_palette_for_background` 自动匹配。
  - 每页需要事实或数据时，应保留来源信息并避免匿名实体。
- 对 `content_slide`、`card_grid`、`two_column`、`three_column`、`kpi_dashboard`、`chart_slide` 这类信息页，优先做结构化参数和容量控制，不要直接塞 300-400 字长段落。
- 增加或调整模板时，同步检查：
  - `templates/single-page/*.json`
  - `templates/full-decks/*.json`
  - `generators/__init__.py`
  - `references/generators.md`
  - 前端 `contentTypeLabels` 或模板展示逻辑

## Prompt 与内容规划约定

- 前端传入 outline 时，后端应在创建任务前保证：
  - 所有页面 `content_type` 合法。
  - 空 `description` 已补齐或明确交给后续阶段补齐。
  - `content_plan` 不为空时结构可被 SlideExecutor 稳定消费。
  - `background` 来自实际可用背景列表，而不是硬编码但不存在的主题。
- Prompt 示例中的 JSON 字段值应使用真实合法 id，避免写“布局名”这类占位值。
- 内容质量要求要和具体布局容量一致：
  - bullet 页控制条数和单条长度。
  - 图文页可以使用较长 paragraph。
  - 图表页必须给结构化 labels/datasets。
  - KPI 页必须给 value/label/delta/baseline 等必要字段。
- 需要真实数据时，尽量在规划阶段明确数据项、来源和格式，减少 SlideExecutor 同时承担搜索、抽取、排版、校验的压力。

## 验证方式

选择能覆盖风险的最小有效验证，不要无差别运行全量命令。

后端在 `ppt-agent/backend` 中运行：

```powershell
go test ./...
go test ./pkg/web ./pkg/task ./pkg/agent/deep
go build ./...
```

前端在 `ppt-agent/frontend` 中运行：

```powershell
npm run build
npm run dev
```

Python 生成器在 `ppt-agent` 或 `ppt-agent/skills/visual_designer` 相关目录运行，按改动选择：

```powershell
python -m py_compile skills/visual_designer/generators/*.py
python skills/visual_designer/generators/generator.py
```


不要在没有必要时启动长期服务、数据库迁移、全量生成所有模板或跑高成本 LLM 任务。如果验证因为依赖、数据库、模型凭据、外部服务或沙箱限制失败，应说明执行的命令和阻塞原因。

## 人机协作工作流

本项目采用**事项登记 + 按需分流**的迭代模式，所有 Agent 都必须遵守：

- **小需求可直接闭环**：若范围清晰、可在当前会话内完成，可直接实现；若过程中发现无法当场闭环，必须补登记到 [][`docs/issues/todo.md`][docs/issues/todo.md](./docs/issues/todo.md)
- **中大型需求必须先登记**：凡是跨多文件/多层级、需要设计判断、需要跨会话追踪的事项，必须先记录到[docs/issues/todo.md](./docs/issues/todo.md)
- **TODO 按时间只追加**：`计划时间` 使用首次登记日期；新增事项按计划时间顺序追加到表尾，不得插入或重排历史条目；事项完成时在原行回填 `完成时间`、状态和产出链接
- **预研文档按需放在 `docs/research/`**：当边界不清、依赖外部系统、存在多方案权衡时，先写预研，再决定实施路线
- TODO 事项有两条正式路线
  - `direct`：直接进入实现
  - `opsx`：进入 `openspec/changes/<change>/` 工作流
- **Agent 必须回填状态与链接**：无论走 `direct` 还是 `opsx`，都要把 research、spec/change、最终产出回填到 `todo.md`

详见：[docs/workflows/human-agent-iteration.md](docs/workflows/human-agent-iteration.md)

## 安全与敏感信息

- 不要在回复、日志或提交说明中暴露 API key、token、cookie、数据库 DSN、邮箱授权码、模型服务凭据或私有证书路径。
- `.env` 只用于本地运行，不要把新增凭据写入文档或示例。
- 搜索、模型调用和外部转换工具失败时，报告错误类型即可，不要泄露完整敏感上下文。

## 变更纪律

- 改动范围贴合用户请求和所属模块边界。
- 优先复用现有 loader、generator、prompt 模板和 API 类型，不轻易引入新框架。
- 质量问题要从“数据契约 -> 内容规划 -> 生成器容量 -> 渲染验证 -> 前端展示”整条链路定位，不要只做表面样式微调。
- 行为变更应补充聚焦测试或最小可复现样例，尤其是任务状态、outline 生成、缩略图、QA 修复和文件下载。
- 最终回复应列出修改文件和验证结果；若未运行验证，说明原因。

## 文档索引

- [需求登记表](./docs/issues/todo.md)

- [人机协作迭代工作流](./docs/workflows/human-agent-iteration.md)

- [预研文档规范](./docs/research/README.md)
