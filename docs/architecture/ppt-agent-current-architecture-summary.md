# PPT Agent 当前架构设计总结

更新时间：2026-08-27

本文档记录当前 PPT Agent 的架构基线，用于后续新 feature 开始前快速对齐系统边界、主流程和高风险契约。

## 1. 核心设计原则

PPT Agent 当前架构的核心判断是：**LLM 负责规划和内容判断，代码负责确定性执行和交付闭环**。

```text
LLM 负责：
- 理解用户需求、受众、场景和交付目标
- 规划整套 PPT 的叙事结构、章节节奏和页面角色
- 选择合法 content_type，填写 description 与 content_plan.components
- 在规划阶段审查和润色 DeckSpec

代码负责：
- schema、容量、content_type、background、output_file 校验
- tasks.draft.json / tasks.json 原子写入和最终提交
- 按 task_id 并发调度 Python generator
- 单页 PPTX、合并文件、缩略图、下载和 SSE 状态
- runtime events、conversation、错误和交付元数据持久化
```

## 2. 主流程

当前目标主流程是“重规划、轻渲染”：

```text
用户输入 / outline
  ↓
创建入口意图分类：新建 PPT / 修改已有任务 / 主题澄清 / 闲聊
  ↓
PPTPlanner 生成 DeckSpec 草稿
  ↓
TaskPlanReviewer 根据 Go 审查报告检查并修正需求偏差、叙事、容量和组件 schema
  ↓
Go 最多执行 3 轮校验/修正循环
  ↓
最终 commit 为 tasks.json
  ↓
DeckRenderWorkflow 按页并发渲染
  ↓
输出文件对账、合并、缩略图、下载交付
  ↓
前端展示 assistant 显式输出、执行观察和交付状态
```

规划阶段允许使用 `tasks.draft.json` 承载中间态；只有最终 `commit` 才写正式 `tasks.json` 并触发完整硬校验。这样避免规划/润色过程中的半成品 JSON 直接打断主流程。

## 3. 分层职责

| 层级 | 路径 | 当前职责 |
| --- | --- | --- |
| Web/API | `ppt-agent/backend/pkg/runtime/web` | 任务创建、鉴权、SSE、下载、缩略图、conversation、runtime event 查询 |
| 任务状态 | `ppt-agent/backend/pkg/runtime/task` | 任务生命周期、工作目录、DB 持久化、SSE 缓存、交付终态校验 |
| 规划编排 | `ppt-agent/backend/pkg/agent/deck` | Planner、manifest 工具、草稿/提交、恢复、并发渲染 workflow |
| Prompt | `ppt-agent/backend/pkg/prompts/{planner,reviewer,fixer}` | 首轮规划、规划质量修正和生成后定点修复的独立职责提示词 |
| 模板加载 | `ppt-agent/backend/pkg/templates` | 读取 `component_contracts.json` 页面类型契约和 theme 元数据；不再维护固定整套 preset |
| 前端工作台 | `ppt-agent/frontend/src` | 生成入口、任务列表、预览下载、会话、执行观察和运行状态 |
| PPT Deck Planner 与生成器 | `ppt-agent/skills/ppt-deck-planner` | Skill 契约、组件 schema、Python generator、图片落盘和渲染测试 |

## 4. DeckSpec / tasks.json 契约

`tasks.json` 是跨 Planner、Renderer、TaskManager、前端展示和下载交付的核心契约。

关键字段：

```json
{
  "title": "整套 PPT 标题",
  "theme": "ocean_soft",
  "template": "dynamic",
  "tasks": [
    {
      "task_id": "3",
      "page_index": 3,
      "title": "页面标题",
      "content_type": "content_slide",
      "description": "页面内容语义说明",
      "content_plan": {
        "slide_intent": "本页在整套 PPT 中承担的角色",
        "components": [
          {"type": "argument_block", "title": "核心判断", "body": "完整论述"},
          {"type": "list", "title": "证据", "items": ["事实 1", "事实 2"]}
        ],
        "capacity_hint": {"estimated_density": "normal", "overflow_risk": "low"},
        "reviewer_status": {"locked": true, "issues": []}
      },
      "output_file": "3_页面标题.pptx",
      "status": "pending"
    }
  ]
}
```

约束：

- `content_type` 必须是稳定英文 id，不允许中文显示名或未实现组件类型冒充页面类型。
- `content_plan.components` 是生成器消费的主数据源；旧 `summary/elements` 不再作为可维护契约，`description` 只保留页面意图摘要。
- 组件只表达语义内容、关系和优先级，不写坐标、字号、颜色、透明度、边距等视觉参数。
- 背景图和实景图通过 `visual_intent.local_path` 或 `image.local_path` 引用任务工作区文件；`background` 仅作为历史字段处理。
- `output_file` 必须唯一、有序、可由后端安全拼接。
- 固定整套 preset、模板推荐 API 和本地背景目录已经移出主链路；`template` 字段仅保留为兼容/来源标记，不再驱动页面结构。

## 5. 组件优先生成器

当前 Python 生成器已经从“每种页面硬编码消化旧参数”迁移到“组件优先”：

```text
render_task.py
  ↓
build_params(content_type, task, manifest)
  ↓
generate_xxx(..., components=content_plan.components)
  ↓
component_layout.py
  ↓
base.py / python-pptx
```

设计边界：

- `components` 非空时只消费显式组件，避免旧参数和组件重复渲染。
- `comparison_matrix/table` 有独立表格渲染，不再退化为卡片。
- `argument_block` 支持 220-420 字完整论述，适合深度说明页。
- `list/numbered_list/bullet_list/evidence_list` 是列表语义组件，不应拆成泛化卡片。
- `agenda` 应消费多条 `toc_item`，不得把整套目录拼成一个长卡片。

## 6. 风格输入策略

固定风格、称谓、组织背景、配色或表达偏好由用户在当前任务提示词中显式说明，Planner 只在本次任务内使用。

## 7. 前端呈现边界

前端主会话展示的是模型显式输出内容，不展示模型隐藏推理，也不把系统预设话术伪装成思维链。

当前约定：

- Planner、Reviewer 和 Fixer 的普通 assistant content 直接作为用户可见 Markdown 展示，不由系统重新总结或伪造隐藏思维链。
- 工具调用参数、工具结果、token 细节保留在 runtime events / 执行观察中。
- `/conversation` 只把 `assistant_output` 作为普通 AI 消息返回，前端用 Markdown 渲染。

## 8. 后续 feature 起点

下一阶段建议围绕主流程继续推进：

1. 为 PPTPlanner、TaskPlanReviewer、PPTFixer 的交接增加更多 fixture 与低成本线上回归样例。
2. 把 `templates/component_contracts.json` 继续作为后端校验、Planner 审查和生成器能力暴露的单一契约入口。
3. 继续扩展组件族：长论述、列表、表格、图表、流程、案例、架构图分别走独立渲染族，减少“一切变卡片”。
4. 建立低成本线上 smoke fixture，覆盖 agenda、argument_block、comparison_table、kpi_dashboard 和缩略图交付。
5. 将 done/todo/decisions 继续作为架构基线记录，不再把每个小 bug 写成长期历史。
