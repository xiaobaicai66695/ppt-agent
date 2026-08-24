# PPT Agent 智能演示文稿生成系统

PPT Agent 是一个面向中文演示文稿生产的多智能体系统。后端使用 Go 与 CloudWeGo Eino ADK，前端使用 Vue 3、TypeScript 与 Vite，PPT 生成器使用 Python 和 `python-pptx`。系统围绕意图识别、DeckSpec 规划、规划审查润色、按页渲染、素材管理、缩略图预览、会话记录和运行观测构建，目标是稳定生成结构清晰、风格统一、可交付的 PPT。

本文档尽量使用中文描述。代码标识、环境变量、命令、文件名、接口路径和第三方库名称保留原文，避免影响复制执行和模型识别。

## 项目结构

```text
ppt-agent/
├── backend/                         # Go 后端服务
│   ├── main.go                      # 服务入口
│   ├── go.mod                       # Go 依赖
│   └── pkg/
│       ├── agent/                   # 智能体编排与任务规划
│       │   ├── deck/                # DeckSpec 规划、任务清单和并发渲染编排
│       │   ├── intent/              # 用户意图识别与模板/背景推荐
│       │   ├── learning/            # 用户偏好学习
│       │   ├── router/              # 路由决策
│       │   └── utils/               # 模型回退、上下文压缩、运行元数据
│       ├── auth/                    # 登录、验证码、令牌校验
│       ├── callback/                # 工具调用和模型调用观测
│       ├── db/                      # 数据库访问
│       ├── log_analysis/            # 后台日志分析
│       ├── prompts/                 # 内嵌提示词模板
│       ├── session/                 # 会话消息管理
│       ├── style/                   # 风格画像与提取
│       ├── task/                    # 任务生命周期管理
│       ├── templates/               # 模板、配色和背景目录加载
│       └── web/                     # HTTP、SSE、缩略图、静态前端服务
├── frontend/                        # Vue 3 前端工作台
│   ├── src/
│   │   ├── api.ts                   # 接口客户端
│   │   ├── App.vue                  # 全局布局与样式
│   │   ├── pages/                   # 首页、编排页、任务页、管理页
│   │   └── components/              # 侧边栏、进度、预览和运行事件组件
│   └── public/                      # 前端静态资源
├── skills/
│   └── ppt-deck-planner/             # PPT Deck Planner：组件规划契约与生成器
│       ├── SKILL.md                 # 规划约束和生成契约
│       ├── generators/              # Python 单页生成器
│       ├── templates/               # component_contracts.json
│       └── references/              # 生成器、页面类型和图片搜索参考
├── scripts/                         # 评估、缩略图和样例生成脚本
├── doc/                             # 项目文档
├── test/                            # 测试辅助材料
└── README.md
```

## 运行链路

用户通过前端创建任务后，后端会先识别意图并选择叙事模板、配色和页数。随后 Planner 生成 `tasks.json`。目标主流程会在渲染前加入 Plan Reviewer / Refiner：先审查整套 PPT 的叙事、信息密度、用户画像使用、组件容量、图片检索意图和组件计划，修订到通过质量门后，再由渲染工作池调用 Python 生成器生成单页 PPTX。任务管理器负责合并结果、记录进度、生成缩略图并通过 SSE 推送状态。

```text
用户请求
  ↓
意图识别与模板推荐
  ↓
Planner 规划 DeckSpec / tasks.json
  ↓
Plan Reviewer / Refiner 审查并润色结构化计划
  ↓
DeckRenderWorkflow 按页并发生成 PPTX
  ↓
任务管理器校验产物与缩略图
  ↓
前端展示进度、预览和下载入口
```

## 核心模块

### 后端服务

后端入口是 `backend/main.go`，主要职责包括：

- 提供任务创建、查询、下载、缩略图、会话和管理接口。
- 管理任务目录、任务状态、输出文件和数据库记录。
- 调用 Planner 完成页面规划；目标态通过 Plan Reviewer / Refiner 做渲染前质量门，再由渲染工作池完成分页生成。
- 维护运行元数据，记录模型调用、工具调用、错误和阶段耗时。
- 加载内置模板推荐、配色和组件契约。

### 前端工作台

前端位于 `frontend/`，用于承载真实生产流程，而不是营销页。主要页面包括：

- 创作入口：输入主题、选择模板、选择智能推荐或自定义编排。
- 编排工作台：查看模板结构、页面类型和组件容量。
- 任务页：展示生成进度、缩略图、文件下载、会话和运行事件。
- 管理页：查看用户、任务和日志分析信息。

### PPT Deck Planner

`skills/ppt-deck-planner` 是 PPT 质量的核心契约层。它主要约束 Planner 如何填充 DeckSpec / `tasks.json`，并规定生成器如何消费结构化计划。字号、坐标、颜色、边距等底层绘制细节由 Python generator 和模板 contract 控制，不应交给 LLM 在 prompt 中直接决定。

它规定：

- 合法页面类型与字段结构。
- 每种页面的内容容量和适用场景。
- 图片搜索意图、配色、组件和布局的使用方式。
- `render_task.py` 如何调用 Python 生成器。
- 生成器如何保证文本不溢出、元素不重叠、背景可读。

常见页面类型包括：

```text
title_slide、agenda、section_divider、content_slide、card_grid、
two_column、three_column、timeline、process_flow、stat_slide、
kpi_dashboard、chart_slide、image_text、quote_slide、summary_slide
```

## 图片与风格规则

本地背景库和离线素材目录已经下线。Planner 需要图片时，应先把页面主题转换成可搜索的视觉主体，调用图片搜索工具下载到当前任务工作区，再把 `local_path`、来源链接和署名写入 `visual_intent` 或 `image` 组件。顶层 `background` 字段仅作为历史兼容字段，新规划保持为空。

## 任务清单契约

`tasks.json` 是 Planner、规划审查/润色和渲染工作池之间的核心契约。Planner 只通过 `update_tasks_manifest` 初始化或更新任务清单，避免并发覆盖整个文件。目标态会把它升级为更强的 DeckSpec：页级计划之外，还包含组件级内容计划和容量提示。

单页任务的关键字段：

```json
{
  "task_id": "1",
  "page_index": 1,
  "title": "封面",
  "content_type": "title_slide",
  "layout_variant": "photo_full_bleed_center",
  "description": "页面内容描述",
  "background": "",
  "output_file": "1_cover.pptx",
  "status": "pending",
  "content_plan": {
    "summary": "页面核心信息",
    "visual_intent": {
      "role": "cards",
      "asset_purpose": "scene",
      "asset_query": "delivery drone flying above urban neighborhood, wide landscape",
      "local_path": "images/unsplash-example.jpg"
    },
    "components": [
      {
        "component_id": "headline-1",
        "type": "headline",
        "text": "页面主论点"
      },
      {
        "component_id": "card-1",
        "type": "feature_card",
        "title": "组件标题",
        "body": "组件承载的事实、观点或案例",
        "emphasis": "primary"
      }
    ],
    "capacity_hint": {
      "density": "normal",
      "overflow_risk": "low"
    }
  }
}
```

重要约束：

- `content_type` 必须是合法英文标识，不写中文布局名。
- `theme` 是整套 PPT 的配色锚点。
- `background` 是历史本地背景字段，新规划保持为空。
- `description` 和 `content_plan` 描述内容语义，不写坐标、字号、颜色、透明度等实现细节。
- `components` 描述页内组件的类型、内容、关系和强调级别；底层布局仍由 generator 决定。
- 数据页必须包含可结构化的数据和来源信息。

## 规划质量门

目标主流程中，Planner 输出不会直接进入渲染。Plan Reviewer 先检查 DeckSpec，Plan Refiner 根据问题结构化修订，直到通过质量门或达到循环上限。

质量门重点检查：

- 用户本次意图和受众是否被正确表达。
- 用户画像是否发生跨场景过拟合；确定性信息可直接注入，风格偏好必须通过场景门控。
- 整套 PPT 是否有清晰叙事和章节节奏。
- 每页是否有足够信息密度，是否重复、空洞或超载。
- `content_type`、`layout_variant`、组件计划和模板容量是否匹配。
- 图表、KPI、案例页是否缺必要数据、事实或来源。

## 运行模式

默认使用 Planner 模式：

```bash
AGENT_MODE=planner
```

Planner 负责整体规划，目标态的 Reviewer / Refiner 负责渲染前计划质量，渲染工作池负责按页并发调用生成器。当前生成闭环主要依赖代码维护的任务状态和输出文件校验；生成后视觉 QA 默认不作为低成本生成链路的必要步骤。

## 模型与回退

模型配置通过环境变量提供。主模型不可用或遇到限流时，会按备用模型顺序回退：

```text
ARK_MODEL → ARK_MODEL_BACKUP1 → ARK_MODEL_BACKUP2 → ARK_MODEL_BACKUP3 → ARK_MODEL_BACKUP4
```

轻量意图识别可使用 `ARK_TEXT_MODEL`。视觉质量审查可使用 `ARK_QA_MODEL`。

## 上下文压缩与观测

系统会在上下文变长时压缩历史信息，保留系统指令、结构化摘要和最近对话。运行过程中会记录：

- 当前阶段、耗时、任务进度。
- 模型调用、工具调用、重试和错误。
- Token 统计与压缩前后信息。
- 输出文件、缩略图和任务完成状态。

这些信息会写入本地运行元数据和数据库，前端任务页可以按需查看。

## 本地运行

### 后端

```bash
cd backend
go mod tidy
go build -o ppt-agent .
./ppt-agent -mode web -addr :8080
```

### 前端

```bash
cd frontend
npm install
npm run dev
```

开发环境中，前端通过代理访问后端接口。生产部署时，后端会托管前端构建产物。

### Python 生成器

```bash
python -m py_compile skills/ppt-deck-planner/generators/*.py
python skills/ppt-deck-planner/generators/generator.py
```

生成器依赖 `python-pptx`。缩略图和 QA 渲染链路依赖 LibreOffice 与 Poppler。

## 常用验证

后端聚焦验证：

```bash
cd backend
go test ./pkg/web ./pkg/task ./pkg/agent/deck
go build ./...
```

前端验证：

```bash
cd frontend
npm run build
```

## 环境变量

| 变量 | 说明 | 默认值 |
| --- | --- | --- |
| `ARK_API_KEY` | 模型服务密钥 | 必填 |
| `ARK_BASE_URL` | 模型服务地址 | 必填 |
| `ARK_MODEL` | 主模型 | 必填 |
| `ARK_MODEL_BACKUP1` 到 `ARK_MODEL_BACKUP4` | 备用模型 | 可选 |
| `ARK_TEXT_MODEL` | 轻量文本模型 | 可选 |
| `ARK_QA_MODEL` | 视觉审查模型 | 可选 |
| `AGENT_MODE` | 执行模式 | `planner` |
| `PLANNER_CONCURRENCY` | 分页生成并发数 | `5` |
| `ENABLE_QA` | 是否启用 QA | `false` |
| `STREAM_TIMEOUT` | 单次流式调用超时 | `3m` |
| `PYTHON_BIN` | Python 可执行文件 | `/root/pptx_env/bin/python` |
| `MYSQL_DSN` | MySQL 数据源 | 可选 |
| `LOG_FILE` | 日志文件路径 | 可选 |
| `LOG_ANALYSIS_IDLE_INTERVAL` | 空闲日志分析间隔 | `0` |
| `COZELOOP_API_TOKEN` | CozeLoop 令牌 | 可选 |
| `COZELOOP_WORKSPACE_ID` | CozeLoop 工作区 | 可选 |

## 部署与冒烟

服务器默认目录：

```text
/ppt/ppt-agent
```

默认服务命令：

```bash
cd /ppt/ppt-agent/backend
./ppt-agent-linux -mode web -addr :8080
```

上线后至少检查：

- 新进程存活。
- `:8080` 正常监听。
- 启动日志出现数据库连接成功且无立即失败。
- `/api/health` 返回 200。
- `/api/templates` 返回可用模板和组件布局数据。
- 涉及生成链路时，创建 1 个低成本 1 到 2 页任务并清理测试数据。

## 版本与外部依赖

| 依赖 | 用途 |
| --- | --- |
| Go | 后端服务与智能体编排 |
| CloudWeGo Eino ADK | 多智能体运行框架 |
| Vue 3、TypeScript、Vite | 前端工作台 |
| Python、python-pptx | PPTX 生成 |
| LibreOffice | PPTX 转 PDF |
| Poppler | PDF 转图片 |
| MySQL | 用户、任务、会话和运行事件存储 |
