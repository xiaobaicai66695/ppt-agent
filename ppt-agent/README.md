# PPT Agent - 智能 PPT 制作助手

基于 CloudWeGo [Eino](https://github.com/cloudwego/eino) ADK 框架的多 Agent PPT 生成系统，支持两种执行模式，结合 Skill 系统实现模块化的设计规范与自动化质量审查。

## 项目结构

```
ppt-agent/
├── backend/                              # Go 后端服务
│   ├── main.go                          # 主入口
│   ├── go.mod                           # 依赖管理 (Go 1.25.0, eino v0.8)
│   └── pkg/
│       ├── agent/                       # Agent 核心模块
│       │   ├── skill.go                 # Skill 加载与格式化
│       │   ├── command/operator.go       # 命令行操作器
│       │   ├── deep/                     # DeepAgent 模式 (并行)
│       │   │   ├── agent.go             # 主 Agent (PPTTaskDeepAgent)
│       │   │   ├── run.go               # 运行逻辑
│       │   │   ├── slide_executor.go   # 幻灯片生成执行器
│       │   │   ├── reviewer.go           # 视觉 QA 审查器
│       │   │   ├── fixer.go              # 修复执行器
│       │   │   └── types.go             # 任务清单与状态类型
│       │   ├── planexecute/             # Plan-Execute-Replan 模式 (串行)
│       │   │   ├── planner.go           # 规划 Agent
│       │   │   ├── executor.go          # 执行 Agent
│       │   │   └── replanner.go         # 重规划 Agent
│       │   ├── agents/wrap_plan.go      # Plan 写入包装器
│       │   └── utils/
│       │       ├── model.go             # 模型配置与 fallback chain
│       │       ├── compressor.go        # 上下文压缩
│       │       ├── token_tracker.go     # Token 计数
│       │       └── utils.go            # 格式化工具
│       ├── tools/                       # 工具模块
│       │   ├── tools.go                 # 工具注册入口
│       │   ├── wrap.go                 # 工具包装器 (审批/HITL)
│       │   ├── human_in_the_loop.go     # 人机交互审批工具
│       │   ├── python_runner.go         # Python 脚本执行器
│       │   ├── ppt/ppt_tool.go         # PPT 生成工具
│       │   ├── qa/qa_tool.go           # 视觉 QA 审查工具
│       │   ├── search/search_tool.go    # 搜索工具
│       │   ├── plan_tool.go            # 计划管理工具
│       │   ├── checkpoint_tool.go       # 检查点工具
│       │   ├── submit_result.go         # 结果提交工具
│       │   ├── edit_file.go            # 文件编辑工具
│       │   ├── read_file.go            # 文件读取工具
│       │   ├── bash_tool.go            # Shell 命令工具
│       │   └── option.go              # 工具选项
│       ├── web/                         # HTTP 服务
│       │   ├── server.go              # 主服务器
│       │   ├── handler.go             # 请求处理器
│       │   ├── streamer.go            # SSE 流式响应
│       │   ├── middleware.go          # 中间件
│       │   ├── thumbnail.go           # 缩略图生成
│       │   └── health.go              # 健康检查
│       ├── task/manager.go            # 任务生命周期管理
│       ├── human/                      # 人机交互模块
│       │   ├── manager.go             # 交互管理器
│       │   └── prints/prints.go       # 输出格式化
│       ├── auth/                       # 认证模块
│       │   ├── auth.go                # 认证逻辑
│       │   └── email.go               # 邮件验证
│       ├── db/db.go                   # 数据库操作 (SQLite)
│       ├── store/store.go             # 内存存储
│       ├── logger/logger.go           # 日志模块
│       ├── metrics/metrics.go         # 指标收集
│       ├── prompts/                    # Prompt 模板 (go:embed)
│       │   ├── prompts.go             # 模板加载器
│       │   └── executor_user_prompt.tmpl
│       ├── params/consts.go           # 上下文常量
│       └── generic/                   # 通用类型
│           ├── plan.go               # Plan 结构
│           └── time.go
├── frontend/                            # React 前端
│   ├── src/
│   │   ├── App.tsx
│   │   ├── pages/DashboardPage.vue    # 仪表盘
│   │   └── components/
│   │       ├── SlidePreviewCard.vue   # 幻灯片预览卡片
│   │       └── ProgressBar.vue        # 进度条
│   └── ...
├── skills/                              # Skill 系统
│   └── visual_designer/
│       ├── SKILL.md                   # 设计规范
│       ├── scripts/design_assistant.py
│       └── templates/                  # 模板资产
│           ├── single-page/          # 12种单页布局
│           └── full-decks/           # 6个完整模板
└── README.md
```

## 核心架构

### 两种执行模式

系统支持两种 Agent 执行模式，由环境变量 `AGENT_MODE` 控制：

#### 模式一：DeepAgent (并行，默认)

```
AGENT_MODE=deep  # 或不设置
```

使用 Eino prebuilt `deep` 编排模式，Master Agent 并行调度三个子 Agent：

```
PPTTaskDeepAgent (Master)
├── SlideExecutor ──── 生成幻灯片 (python-pptx)
├── Reviewer     ──── 视觉 QA 审查 (多模态 LLM)
└── Fixer        ──── 修复 QA 发现的问题
```

幻灯片生成可并行执行（默认 5 个并发，受 `DEEP_AGENT_CONCURRENCY` 控制）。

#### 模式二：Plan-Execute-Replan (串行)

```
AGENT_MODE=planexecute  # 显式设置
```

传统串行模式，三个 Agent 顺序执行：

```
Planner ──▶ Executor ──▶ Replanner ──▶ (循环)
```

### 模型 Fallback 链

所有 LLM 调用使用多级 fallback 链，自动处理 429 限流：

```
ARK_MODEL → ARK_MODEL_BACKUP1 → ARK_MODEL_BACKUP2 → ... → ARK_MODEL_BACKUP4
```

触发 429 时失败模型暂停 30 秒，继续尝试下一个；全部失败才报错。

### 上下文压缩

DeepAgent 模式使用 `ChatModelCompressor` 压缩对话历史，触发条件（二选一）：

- 消息数超过 12 条
- Token 估计超过 30,000

压缩策略：`system + [压缩摘要] + [最近 4 轮对话]`

### Visual QA 流水线

```
PPTX 文件 ──▶ pptx_qa_converter.py ──▶ PDF ──▶ pdftoppm ──▶ JPEG (150 DPI)
                                                              │
                                                              ▼
                                                    多模态 LLM 视觉审查
                                                              │
                                                              ▼
                                                   .qa_result.json
```

- **DeepAgent 模式**：每完成一页立即审查（单页 QA）
- **Plan-Execute 模式**：全部生成后批量审查

QA 严重等级：`high` → `medium` → `low`。每页最多修复 2 次，仍有问题则跳过。

## 工具列表

| 工具 | 文件 | 说明 |
|------|------|------|
| Search | `search/search_tool.go` | 互联网内容搜索 |
| PPT | `ppt/ppt_tool.go` | 调用 python-pptx 生成 PPT |
| single_qa_review | `qa/qa_tool.go` | 单页视觉质量审查 |
| batch_qa_review | `qa/qa_tool.go` | 批量视觉审查 |
| update_progress | `checkpoint_tool.go` | 更新任务进度 |
| create_ppt_plan | `plan_tool.go` | 创建 PPT 计划 |
| submit_result | `submit_result.go` | 提交最终结果 |
| PythonRunner | `python_runner.go` | Python 脚本执行器 |
| EditFile | `edit_file.go` | 编辑文件 |
| ReadFile | `read_file.go` | 读取文件 |
| Bash | `bash_tool.go` | 执行 Shell 命令 |

所有工具均通过 `wrap.go` / `human_in_the_loop.go` 支持人机审批。

## Skill 系统

Skill 负责将专家知识注入 Agent prompt，引导决策而非替代执行。

### visual_designer

视觉设计规范，包含 12 种单页布局模板 + 6 个完整模板，6 种配色系统，NEVER 清单。

### pptx_writer

PPTX 操作规程，包含MANDATORY 加载指令和精确操作步骤。

## 编译运行

### 后端

```bash
cd backend
go mod tidy
go build -o ppt-agent.exe .
./ppt-agent.exe
```

### 前端

```bash
cd frontend
npm install
npm run dev   # 端口 3000，/api 代理到 localhost:8080
```

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `ARK_API_KEY` | API 密钥 | 必填 |
| `ARK_MODEL` | 主模型 | 必填 |
| `ARK_MODEL_BACKUP1~4` | 备用模型 | 可选 |
| `AGENT_MODE` | 执行模式：`deep` 或 `planexecute` | `deep` |
| `INTERACTIVE` | 交互模式：`true`/`false` | `true` |
| `DEEP_AGENT_CONCURRENCY` | DeepAgent 并发数 | `5` |
| `COZELOOP_API_TOKEN` | CozeLoop 可观测 | 可选 |
| `COZELOOP_WORKSPACE_ID` | CozeLoop 工作区 | 可选 |

## 版本信息

| 依赖 | 版本 |
|------|------|
| Go | 1.25.0 |
| eino | v0.8.8 |
| python-pptx | >= 0.6.21 |
| LibreOffice | 用于 PPTX→PDF 转换 |
| poppler-utils | 用于 PDF→JPEG 转换 |
