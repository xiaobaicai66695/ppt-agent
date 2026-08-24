# 模板编排与 Agent UI 预研

日期：2026-08-24

## 核心结论

当前主要问题不是“前端样式过时”，而是前端仍按旧的手动模板编排理解系统：用户先挑页面布局、背景、标题和描述，再提交 outline；但后端主流程已经转向“用户意图 -> 推荐策略 -> Planner 组件级规划 -> 生成器确定性渲染”。因此前端需要从 Outline Editor 调整为 Plan Studio：先解释推荐策略，再允许用户调整章节、页数、视觉方向和关键页面，而不是逐页手填旧模板字段。

UI 配色也应该降噪。当前黑色侧栏 + 白色顶部栏 + 灰白画布让产品显得像后台管理系统，和“生成式创作工作台”不匹配。建议改为低饱和暖灰/青瓷底色、深青文字、青绿主操作、琥珀/珊瑚强调色，保留专业感但减少黑白硬对比。

## 当前代码断点

1. 前端自定义编排仍是旧心智

- `frontend/src/pages/ComposePage.vue` 仍维护 `layouts/backgrounds` 两类资源 tab，并用硬编码 `layoutCategories` 分组。
- 初始页面固定为 `title_slide/content_slide/summary_slide` 三页空页面。
- `buildOutline()` 只提交 `template/theme/title/slides`，虽然透传 `content_plan`，但页面上没有组件级编辑和推荐解释。
- `startGeneration()` 直接调用 `createTaskWithOutline()`，一旦传 outline 就跳过主 Agent 的规划优势。

2. 后端推荐策略可用，但没有变成产品接口

- `HomePage.vue` 的“智能推荐”直接 `createTask(query, { mode: 'recommended' })`，用户看不到推荐理由、候选模板、页数、背景策略和风险。
- `template_recommendation.go` 的 `recommendTemplateStrategyWithIntent()` 已能选择模板、页数、主题和背景，但只在创建任务内部消费。
- `recommendedOutlineFromTemplate()` 返回空 `Slides`，本质是“推荐风格 + 让 Planner 后续规划”，不是可编辑的页面大纲。
- `/api/templates`、`/api/templates/layouts` 仍偏资源目录，没有 `POST /api/templates/recommend` 这类“解释型推荐”接口。

3. 组件化契约还没有充分暴露给前端

- 生成器已围绕 `content_plan.components` 演进，但前端只展示 layout contract 的容量字段。
- `component_contracts.json` 尚未形成前端可消费的“组件族、容量、适用页面、例子、不可用组合”视图。
- 结果是用户只能挑页面模板，不能理解 Planner 为什么用图文、长论述、表格、流程或指标组件。

## 市面 Agent 平台参考

- ChatGPT/Codex 类产品更强调会话主线和工作区入口：用户先给目标，再由系统选择工具和产物形态；不要求用户先逐项配置底层模板。参考：[ChatGPT pricing/overview 页面展示其多工具工作形态](https://chatgpt.com/pricing/)。
- Claude Artifacts 的关键启发是“对话和产物分离”：重大产物进入独立窗口，方便继续修改和引用。Anthropic 官方说明 Artifacts 会在对话旁的独立窗口展示内容，适合 PPT 这种可预览产物。参考：[Claude Artifacts Support](https://support.claude.com/en/articles/9487310-what-are-artifacts-and-how-do-i-use-them)、[Anthropic Projects](https://www.anthropic.com/news/projects)。
- Dify 的启发是“可观察工作流”：工作流不是黑盒，而是可视化节点、输入输出和执行状态。PPT Agent 可借鉴为 Planner/Reviewer/Renderer 的阶段卡片。参考：[Dify 官网](https://dify.ai/)、[Dify Workflow](https://dify.ai/blog/dify-ai-workflow)。
- LangGraph Agent Inbox 的启发是“待处理事项/中断队列”：Agent 需要用户确认时，用 inbox/ticket 模式承载未决动作，而不是把所有事件堆在聊天末尾。参考：[Agent Inbox GitHub](https://github.com/langchain-ai/agent-inbox)、[LangChain ambient agents](https://www.langchain.com/blog/introducing-ambient-agents)。

## 建议的产品改造

### 1. 新增推荐预览接口

新增 `POST /api/templates/recommend`：

- 输入：`query`、可选 `page_count`、`mode`、`audience`、`style_preference`。
- 输出：`strategy`、`ranked_templates`、`theme`、`page_count`、`visual_policy`、`background_policy`、`reason`、`risks`。
- 不直接创建任务，只做低成本意图识别和模板推荐。
- 创建任务时继续用 `template_selection.mode=recommended`，但前端提交用户确认后的策略覆盖项。

### 2. Compose 改成 Plan Studio

页面结构建议：

- 顶部：需求输入 + 推荐按钮 + 开始生成。
- 左侧：推荐策略摘要，展示模板、页数、语气、背景/图片策略、配色。
- 中间：章节级大纲，而不是单页布局堆叠；支持增删章节、调整页数范围、标记重点页。
- 右侧：视觉预览和组件契约，包括适合的组件族、容量边界、风险提示。
- 只有进入“高级编辑”时，才展示原来的逐页布局/背景编辑。

### 3. 后端模板推荐补强

- 不再只默认 `generic`；按 intent 的 domain、audience、deliverable 维度打分。
- 推荐结果保留 2-3 个候选模板，给出“为什么推荐/为什么不选”的短说明。
- 当 Unsplash 已配置时，背景策略改为“图片搜索策略”，不再推荐本地背景主题作为主路径。
- 保留本地背景作为无图/失败兜底，但 UI 文案不要叫“背景模板优先”。

### 4. 组件契约前端化

新增 `GET /api/templates/components`：

- 返回组件类型、适用页面、容量、必填字段、示例数据、不可组合规则。
- 前端在推荐预览中显示“本次将重点使用：图文混排、长论述、表格、流程、指标”等组件族。
- 高级编辑里按组件编辑，而不是只编辑整页 description。

## UI 配色方向

建议从当前黑白灰体系改为“轻量创作工作台”：

- Canvas：`#F5F7F4` 或 `#F4F6F2`
- Surface：`#FFFFFF`
- Subtle Surface：`#EEF3F0`
- Primary Text：`#1B2523`
- Secondary Text：`#5D6864`
- Primary Action：`#0F766E`
- Primary Hover：`#115E59`
- Accent：`#D97706` 或 `#D9654A`
- Info：`#2563EB`
- Border：`#D9E2DE`
- Focus Ring：`#2DD4BF`

侧栏不建议继续用近黑色，可改为深青灰 `#17312E` 或直接浅色侧栏；如果保留深色侧栏，也要减少纯黑感，并让选中态使用青绿细条 + 柔和背景，而不是白字黑底硬对比。

## 分阶段落地

P0：把推荐从“直接创建任务”改为“先返回可解释推荐”

- 新增推荐接口。
- 首页智能推荐点击后展示策略确认面板。
- 创建任务时带上确认后的策略。

P1：改造 Compose 为 Plan Studio

- 默认进入推荐/章节编辑，不再默认三张空页。
- 旧逐页编辑折叠到高级模式。
- 展示组件契约和推荐组件族。

P2：统一设计系统

- 将 `App.vue` token 调整为低饱和青绿/暖灰体系。
- 减少暗黑侧栏占比；保留高对比但不使用纯黑白。
- 对 Home、Dashboard、Compose 统一按钮、卡片、状态色和空态。

P3：清理历史包袱

- 标记旧 full-deck preset 的定位：仅作为推荐候选/章节 scaffold，不再作为最终逐页模板真相源。
- 删除或隐藏不再参与组件化主流程的前端入口和后端兼容字段。
- 文档中明确：`content_plan.components` 是生成器主契约，页面 layout 是渲染能力边界，不是用户必须手工填写的设计稿。

## 风险

- 推荐接口如果直接调用大模型，首页会变慢；需要 20-30 秒超时和失败兜底。
- 如果一次性重写 Compose，容易影响现有自定义用户；建议先做推荐确认面板，再重构高级编辑。
- 配色调整不能只改全局 token，部分页面 scoped CSS 写死了颜色，需要逐页清理。
