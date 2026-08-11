# PPT 生成规划质量与压缩可观测性改造

## 背景

本轮集中处理六个相互关联的问题：背景图片规划覆盖不足、模型侧 metadata 过重、ToolCall 历史不完整、上下文压缩不可见且容易弱化用户要求、模板容量过紧，以及智能推荐机械继承固定 18 页模板。

事项归入“生成编排与交付闭环”“Agent Harness 与可观测性”“PPT 视觉质量与 Visual Designer 能力”三个长期方向，登记为 `PPT-FLOW-002`，采用 OpenSpec 路线实施：`openspec/changes/ppt-agent-generation-quality-and-observability/`。

## 设计取舍

- 智能推荐只提供模板视觉语言、配色、背景主题和建议页数，页面结构由主 Agent 围绕用户目标动态规划；显式选择 preset 时继续保留模板脚手架语义。
- 背景选择采用“主 Agent 主动规划 + manifest 确定性校验补齐”：优先视觉叙事页，兼顾普通内容页可读性，数据密集页保留清晰画布，整套目标覆盖率为 45%-65%。
- 模板容量区分正常写作的 `target_*` 区间和渲染安全边界 `max_*`，让内容完整性、版式容量和动态字号共同决定密度。
- 模型上下文只接收代码维护的 `generated_slides/total_slides`；工具、token、告警和压缩统计继续保留在运行时观测快照中。
- Prompt 以目标、决策顺序、合法候选和结构化输出为契约，通过覆盖完整场景做选择，不依赖不断叠加局部否定规则。

## 实现

### 推荐与背景规划

- 任务创建提前执行一次结构化 LLM 意图分类，并在 TaskManager 与推荐解析间复用结果。
- 意图结果补充推荐背景与背景使用策略；推荐页数采用有界模型建议，模型不可用时降级为通用风格和 12 页默认值。
- 用户明确写出“生成/制作/共 N 页”时，由代码提取合法总页数并覆盖 LLM 建议；“修改第 N 页”仍视为页码引用。
- 推荐模式创建 `recommended_style` outline，不复制 preset 的 `DefaultSlides`；显式 preset 保持 `template_scaffold`。
- 主 Agent 获得真实本地背景主题及图片数量，manifest 工具校验引用、按页面语义补齐低覆盖页面，并轮换同主题图片。

### 内容容量与 Prompt

- 放宽 `content_slide`、`summary_slide`、`two_column`、`three_column`、`card_grid`、`quote_slide` 六类信息页容量。
- 同步 `visual_designer/SKILL.md`、`references/generators.md` 与 Web 端模板说明。
- 重写主 Agent 规划指令，统一背景、容量、视觉意图和分批生成的正向决策顺序。

### RuntimeMeta、ToolCall 与压缩

- `StatusBar()`、`slide_progress` 和 `manifest_validated` 的进度 metadata 只暴露 `done/total`。
- callback 捕获的全部工具类型统一写入 RuntimeMeta event sink，并持久化到 MySQL；80 条 `RecentEvents` 只作为 SSE 热窗口，不再代表完整历史。
- 前端按事件 ID 合并持久化历史和 SSE 尾部，任务切换、重连和终态刷新不再覆盖早期 `read_file`、`update_tasks_manifest`、`search`、Python 等工具调用。
- 任务完成归档时的数据库与文件 I/O 移出 `TaskState.Mu`，会话消息的数据库写入也不再占用内存快照锁；相关数据库读写增加 5 秒上限，避免 MySQL 抖动拖住 `/conversation` 或 Agent callback。
- 压缩 handoff 新增 `user_intent_summary`、`preserved_requirements`、`progress_summary` 和 `conversation_summary`，优先锚定首个真实用户请求与后续明确约束。
- Compression 作为独立时间线类别展示压缩前后消息数、token、移除量、节省量、用户目标和保留要求，并兼容旧事件字段。

## 验证

- 后端聚焦测试：`go test ./pkg/agent/intent ./pkg/agent/deck ./pkg/agent/utils ./pkg/prompts ./pkg/templates ./pkg/web`
- 后端全量测试：`go test ./...`
- 并发回归：`go test -race ./pkg/task ./pkg/session ./pkg/web`
- 后端构建：`go build ./...`
- 前端单测：17 项通过。
- 前端生产构建：`npm run build` 通过。
- 24 个 single-page 模板 JSON 全部可解析。
- 容量 contract 回归测试覆盖六类放宽模板，验证目标最小值、目标最大值和渲染上限的顺序关系。
- OpenSpec：`openspec validate ppt-agent-generation-quality-and-observability --strict`。

## 上线

- 部署目录：`/ppt/ppt-agent`，保留服务器 `.env`、输出目录和运行日志。
- 运行进程：PID `507299`，Linux 后端二进制、本地源码和前端 dist 已同步。
- `GET /api/health`：HTTP 200；`GET /api/templates`：HTTP 200，19 个 preset；`GET /api/backgrounds`：HTTP 200，6 个背景主题。
- 受控智能推荐任务按用户明确要求生成 2/2 页，使用 `tech-intro + report_green`，1/2 页引用真实 `minimalist_blue` 背景，未继承固定 18 页模板。
- `/conversation` 返回 HTTP 200，耗时 3 ms，回放 83 条运行事件；包含 `read_file`、`update_tasks_manifest`、`task`、`python3` 等 ToolCall。
- `manifest_validated` 详情 metadata 字段严格为 `done,total`；部署前端包含独立“上下文压缩”类别和详情视图。
- 两条受控 smoke 任务在验证后已通过 API 删除，未保留在用户任务列表中。
