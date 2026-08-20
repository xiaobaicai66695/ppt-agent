# PPT Agent 当前架构设计总结

更新时间：2026-08-20

本文档记录当前 PPT Agent 的架构基线，用于后续新 feature 开始前快速对齐系统边界、主流程和高风险契约。

## 1. 核心设计原则

PPT Agent 当前架构的核心判断是：**LLM 负责规划和内容判断，代码负责确定性执行和交付闭环**。

```text
LLM 负责：
- 理解用户需求、受众、场景和交付目标
- 规划整套 PPT 的叙事结构、章节节奏和页面角色
- 选择合法 content_type，填写 description 与 content_plan.components
- 对用户画像做场景门控，避免历史风格跨场景过拟合
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
意图识别、模板/配色/背景推荐、用户画像门控
  ↓
Planner 生成 DeckSpec 草稿
  ↓
Plan Reviewer 检查意图、叙事、容量、组件 schema、画像过拟合
  ↓
Plan Refiner 最多 3 轮修订 DeckSpec
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
| Web/API | `ppt-agent/backend/pkg/web` | 任务创建、鉴权、SSE、下载、缩略图、conversation、runtime event 查询 |
| 任务状态 | `ppt-agent/backend/pkg/task` | 任务生命周期、工作目录、DB 持久化、SSE 缓存、交付终态校验 |
| 规划编排 | `ppt-agent/backend/pkg/agent/deck` | Planner、manifest 工具、草稿/提交、恢复、并发渲染 workflow |
| Prompt | `ppt-agent/backend/pkg/prompts/planner` | DeckSpec 规划契约、ReAct 风格可见日志、组件级规划规范 |
| 用户画像 | `ppt-agent/backend/pkg/style`、`pkg/agent/learning` | 确定性用户事实直接注入，历史风格偏好按领域相似度门控 |
| 模板加载 | `ppt-agent/backend/pkg/templates` | full-deck、single-page、theme、background 元数据 |
| 前端工作台 | `ppt-agent/frontend/src` | 生成入口、任务列表、预览下载、会话、执行观察和运行状态 |
| 视觉生成器 | `ppt-agent/skills/visual_designer` | Skill 契约、组件 schema、Python generator、背景和本地素材 |

## 4. DeckSpec / tasks.json 契约

`tasks.json` 是跨 Planner、Renderer、TaskManager、前端展示和下载交付的核心契约。

关键字段：

```json
{
  "title": "整套 PPT 标题",
  "theme": "ocean_soft",
  "template": "business-report",
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
      "background": "minimalist_blue/images/1.jpg",
      "output_file": "3_页面标题.pptx",
      "status": "pending"
    }
  ]
}
```

约束：

- `content_type` 必须是稳定英文 id，不允许中文显示名或未实现组件类型冒充页面类型。
- `content_plan.components` 是生成器优先消费的数据源；旧 `summary/elements/description` 只作为兼容兜底。
- 组件只表达语义内容、关系和优先级，不写坐标、字号、颜色、透明度、边距等视觉参数。
- `background` 只能来自可用背景主题或具体主题图片引用。
- `output_file` 必须唯一、有序、可由后端安全拼接。

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

## 6. 用户画像策略

用户画像分两类处理：

- 确定性事实：姓名、单位、岗位、常用组织名称等，可以直接注入当前提示词。
- 场景敏感偏好：主题、配色、模板、布局、语言风格、成功模式、领域计数等，必须经过当前领域/场景门控。

原则：

- 当前用户显式要求优先级最高。
- 当前任务主题和背景推荐优先于历史风格偏好。
- 没有同领域可靠历史时，只给出弱提示或完全不注入场景偏好。
- Reviewer 必须检查 `profile_overfit`，避免把党建、政务、企业汇报等旧风格套到不相干主题。

## 7. 前端呈现边界

前端主会话展示的是模型显式输出内容，不展示模型隐藏推理，也不把系统预设话术伪装成思维链。

当前约定：

- Planner prompt 可要求模型输出 ReAct 风格的 `Thought/Action/Observation` 文本。
- 这类 `Thought` 是 assistant content 中的可见文本，不是模型原生隐藏思维链。
- 工具调用参数、工具结果、token 细节保留在 runtime events / 执行观察中。
- `/conversation` 只把 `assistant_output` 作为普通 AI 消息返回，前端用 Markdown 渲染。

## 8. 后续 feature 起点

下一阶段建议围绕主流程继续推进：

1. 将 Plan Reviewer / Refiner 从 prompt 规范进一步固化为代码级质量门和 fixture 测试。
2. 把 `templates/component_contracts.json` 作为后端校验、Planner 审查和生成器能力暴露的单一契约入口。
3. 继续扩展组件族：长论述、列表、表格、图表、流程、案例、架构图分别走独立渲染族，减少“一切变卡片”。
4. 建立低成本线上 smoke fixture，覆盖 agenda、argument_block、comparison_table、kpi_dashboard 和缩略图交付。
5. 将 done/todo/decisions 继续作为架构基线记录，不再把每个小 bug 写成长期历史。
