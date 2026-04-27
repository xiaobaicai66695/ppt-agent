# PPT Agent - 智能PPT制作助手

基于 eino ADK 框架的多Agent智能PPT制作系统，使用 Plan-Execute-Replan 范式，结合 Skill 系统实现模块化的设计规范生成和 PPT 文件写入。

## 项目结构

```
ppt-agent/
├── backend/                              # Go 后端服务
│   ├── main.go                          # 主入口
│   ├── go.mod                           # 依赖管理 (Go 1.24.7, eino v0.8)
│   └── pkg/
│       ├── agent/                        # Agent 模块
│       │   ├── skill.go                 # Skill 加载与格式化
│       │   ├── agents/
│       │   │   └── wrap_plan.go        # Plan 写入包装器
│       │   ├── command/
│       │   │   └── operator.go         # 命令行操作器
│       │   ├── planner/                 # 规划 Agent
│       │   │   └── planner.go
│       │   ├── executor/               # 执行 Agent
│       │   │   └── executor.go
│       │   ├── replanner/              # 重规划 Agent
│       │   │   └── replanner.go
│       │   └── utils/
│       │       ├── model.go            # 模型配置
│       │       └── utils.go            # 格式化工具
│       ├── tools/                       # 工具模块
│       │   ├── tools.go                 # 工具入口
│       │   ├── ppt/
│       │   │   └── ppt_tool.go        # PPT 生成工具
│       │   ├── qa/
│       │   │   └── qa_tool.go        # QA 视觉质量审查工具
│       │   ├── search/
│       │   │   └── search_tool.go      # 搜索工具
│       │   ├── bash_tool.go            # Shell 命令工具
│       │   ├── python_runner.go        # Python 脚本执行器
│       │   ├── edit_file.go           # 文件编辑工具
│       │   ├── read_file.go           # 文件读取工具
│       │   ├── submit_result.go       # 结果提交工具
│       │   ├── wrap.go                # 工具包装器
│       │   └── option.go              # 工具选项
│       ├── human/                       # 人机交互模块
│       │   ├── manager.go             # 交互管理器
│       │   └── prints/
│       │       └── prints.go          # 输出格式化
│       ├── generic/                     # 通用模块
│       │   ├── plan.go                # Plan 结构定义
│       │   └── time.go
│       ├── params/                      # 上下文参数
│       │   └── consts.go
│       └── utils/                       # 辅助工具
│           ├── model.go
│           ├── format.go
│           └── helper.go
├── skills/                              # Skill 系统
│   ├── visual_designer/                # 视觉设计 Skill
│   │   ├── SKILL.md                   # 设计规范（含配色、布局、NEVER清单）
│   │   └── scripts/
│   │       └── design_assistant.py
│   └── pptx_writer/                   # PPTX 写入 Skill
│       ├── SKILL.md                   # 操作规范（含MANDATORY指令、NEVER清单）
│       ├── scripts/
│       │   └── pptx_writer.py        # PPT 生成脚本
│       └── requirements.txt
├── frontend/                            # React 前端
│   ├── src/
│   │   ├── App.tsx
│   │   └── index.css
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── output/                             # 运行时输出目录（自动创建）
│   └── {task_id}/
│       └── {slide_index}_{slide_title}.pptx
└── README.md
```

## 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                  Plan-Execute-Replan Loop                   │
│                     (adk/prebuilt/planexecute)              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐           │
│  │ Planner  │───▶│ Executor │───▶│ Replanner │           │
│  │  Agent   │    │  Agent   │    │  Agent    │           │
│  │ (规划)   │    │ (执行)   │    │ (重规划)  │           │
│  └────┬─────┘    └────┬─────┘    └──────────┘           │
│       │                │                                    │
│       ▼                ▼                                    │
│  ┌──────────────────────────────────────┐                 │
│  │        Skill System (Prompt注入)       │                 │
│  ├──────────────────────────────────────┤                 │
│  │  visual_designer  │  pptx_writer      │                 │
│  │  - 设计哲学/NEVER  │  - MANDATORY加载 │                 │
│  │  - 配色/布局决策  │  - 精确操作步骤   │                 │
│  └──────────────────────────────────────┘                 │
│                                                              │
│  ┌──────────────────────────────────────┐                 │
│  │          Human-in-the-Loop             │                 │
│  ├──────────────────────────────────────┤                 │
│  │  工具审批 │ 图片搜索确认 │ 降级处理   │                 │
│  └──────────────────────────────────────┘                 │
│                                                              │
│  ┌──────────────────────────────────────┐                 │
│  │          Tools (via ToolsNode)          │                 │
│  ├──────────────────────────────────────┤                 │
│  │  Python │ Search │ PPT │ BatchQA    │                 │
│  └──────────────────────────────────────┘                 │
│                                                              │
│  ┌──────────────────────────────────────┐                 │
│  │          Vision QA (Multi-modal)        │                 │
│  ├──────────────────────────────────────┤                 │
│  │  PPTX→Images │ Visual Inspection │      │                 │
│  │  QA Report │ Auto-fix loop         │                 │
│  └──────────────────────────────────────┘                 │
└─────────────────────────────────────────────────────────────┘
```

## Skill 系统

### 设计理念

Skill 是**知识的外化机制**，而非"教 AI 做某事"。大模型已经知道绝大多数知识，Skill 的作用是：

- 传递**专家思维方式**（不教技术细节，教如何思考）
- 注入**反模式清单**（告诉 Agent 什么是绝对不能做的）
- 提供**精确操作规程**（对脆弱操作给出低自由度的精确指导）

```
通用 Agent + 优秀 Skill = 特定领域的专家 Agent
```

### Skill 评判标准

| 标准 | 说明 |
|------|------|
| **Token 效率** | 好 Skill = 专家独有的知识 - Claude 已有的知识 |
| **思维方式** | 注重思考框架，而非 Step 1/2/3 步骤清单 |
| **NEVER 清单** | 明确告诉 Agent 什么是"垃圾做法" |
| **触发机制** | Description 包含"做什么"和"何时用"的关键词 |
| **自由度匹配** | 创意任务高自由度，格式操作低自由度 |
| **加载触发** | 关键节点嵌入 MANDATORY 强制加载指令 |

### 1. visual_designer - 视觉设计 Skill

传递专业视觉设计师的思维方式，引导 Agent 在动手之前先回答三个问题：Purpose（解决什么问题）、Tone（审美方向）、Differentiation（差异化定位）。

**核心内容**：
- 设计哲学三问（Purpose / Tone / Differentiation）
- NEVER 清单（颜色、排版、布局、内容各4条禁止项）
- 4种精选配色系统（tech / professional / creative / minimal）
- 布局决策树（按要点数量和内容类型路由）
- 视觉元素使用原则（宁少勿滥，宁精勿滥）

**NEVER 清单示例**：
- 禁止紫色渐变 + 白底（最典型的"AI感"配色）
- 禁止使用 Inter/Roboto/Arial 作为主要英文字体
- 禁止空洞的要点（少于20字视为空洞）

### 2. pptx_writer - PPTX 写入 Skill

OOXML 格式操作规程，确保 Agent 能精确生成有效的 PPTX 文件。

**核心内容**：
- **MANDATORY 加载指令**（执行前必须完整阅读脚本）
- NEVER 清单（禁止跳过依赖检查、禁止硬编码坐标、禁止不验证文件）
- 决策路由表（创建 vs 编辑 vs 特殊操作 → 不同路径）
- 完整配色系统参考（与 visual_designer 保持一致）
- 错误处理规范（图片失败 → warnings 记录，不中断流程）

### Skill 加载机制

Skill 在 `main.go` 中统一加载，通过 prompt 注入到 Planner 和 Executor：

```go
// 1. 加载 SKILL.md 文件内容
loadedSkills, _ := agent.LoadSkillsFromDir(ctx, skillsDir)
skillsContent := agent.FormatSkillsForPrompt(loadedSkills)

// 2. 同时注入到两个 Agent 的 prompt 中
planAgent, _ := planner.NewPlanner(ctx, operator, skillsContent)
executeAgent, _ := executor.NewExecutor(ctx, operator, skillsContent)
```

加载后，Planner 会参考设计规范做规划决策，Executor 会参考操作规范生成 PPT。

### Skill 与工具的区别

| 概念 | 本质 | 作用 | 示例 |
|------|------|------|------|
| Tool | 模型能做什么 | 执行动作 | bash、read_file、write_file |
| Skill | 模型知道做什么 | 指导决策 | 设计规范、审查指南、格式操作规程 |

工具是能力的边界，没有 bash 工具模型就无法执行命令。Skill 则是技巧的注入，没有视觉设计 Skill，模型写出的 PPT 将千篇一律。

## Agent 模块

### Plan-Execute-Replan 循环

```
用户需求
    │
    ▼
┌──────────┐    制定计划    ┌──────────┐    评估进度    ┌──────────┐
│ Planner  │──────────────▶│ Executor │──────────────▶│ Replanner│
│ (规划)   │               │ (执行)   │               │ (重规划) │
└──────────┘               └──────────┘               └──────────┘
    │                           │                           │
    ▼                           ▼                           ▼
  生成幻灯片计划           执行当前步骤              判断是否完成
  (slides[])           生成 PPT 文件           或需要调整计划
                              │                           │
                              ▼                           ▼
                        提交步骤完成              提交最终结果
                              │                      或重新规划
                              └──────────────────────────────┘
```

**Planner**：分析用户需求，制定包含幻灯片列表的完整计划，参考 `visual_designer` 的设计哲学选择配色和布局。

**Executor**：根据当前步骤生成 PPT 文件，参考 `pptx_writer` 的操作规范确保文件正确生成，支持工具调用（代码代理、搜索、PPT生成）。

**Replanner**：评估已执行步骤的正确性，判断计划是否仍然适用，异常时调用 `create_ppt_plan` 重新规划。

## QA 视觉质量审查系统

基于多模态 LLM 的自动化视觉 QA，在所有幻灯片生成完毕后执行批量审查，确保输出质量。

### 工作流程

```
所有幻灯片生成完毕
        │
        ▼
┌──────────────────┐
│ batch_qa_review  │ ──── 自动触发
│ 工具调用          │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ PPTX → 图片转换  │ 150 DPI 分页输出
│ (pptx_qa_converter.py) │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ 多模态 LLM 审查  │ 视觉 AI 模型
│ (Vision QA)      │
└────────┬─────────┘
         │
         ▼
    QA 报告输出
         │
         ├─── 有问题 ──▶ Replanner 生成修复步骤
         │
         └─── 无问题 ──▶ 提交最终结果
```

### 检查问题类型

| 类型 | 说明 | 严重程度 |
|------|------|---------|
| overlap | 文字与形状/图片重叠、线条穿过文字 | high |
| overflow | 文字超出文本框边界被截断 | high |
| contrast | 浅色文字在浅色背景上 | medium |
| spacing | 元素间距不一致、过于靠近 | medium |
| alignment | 同一列元素未对齐、视觉重心不稳 | medium |
| placeholder | 占位符文本残留 (xxxx/lorem) | high |
| ai_style | AI 感特征（装饰线、紫色渐变） | low |

### 审查规则

- 假设有错，不要因为"看起来还行"就跳过
- 每一页都要单独审查，不能跳页
- 高严重程度的问题必须报告
- 发现问题一定要报告，不要遗漏

## 工具模块

| 工具 | 文件 | 说明 |
|------|------|------|
| Search | `search/search_tool.go` | 互联网内容搜索 |
| ImageSearch | `search/search_tool.go` | 图片素材搜索（含审批机制） |
| PPT | `ppt/ppt_tool.go` | 调用 pptx_writer 脚本生成 PPT 文件 |
| BatchQA | `qa/qa_tool.go` | 批量视觉质量审查 |
| Bash | `bash_tool.go` | 执行 Shell 命令 |
| EditFile | `edit_file.go` | 编辑文件内容 |
| ReadFile | `read_file.go` | 读取文件内容 |
| SubmitResult | `submit_result.go` | 提交最终结果 |
| PythonRunner | `python_runner.go` | Python 脚本执行器 |

## 人机交互 (Human-in-the-Loop)

基于 eino ADK 的中断机制实现人机交互审批流程。

### 审批工作流程

```
Agent 调用工具
       │
       ▼
┌──────────────┐
│ 工具包装器    │ ──── 中断等待审批
│ (wrapper)   │
└──────────────┘
       │
       ▼
┌──────────────┐
│ 用户审批     │ ──── Y: 执行 │ N: 拒绝 │ E: 编辑参数
└──────────────┘
       │
       ▼
  ResumeWithParams 恢复执行
```

### 图片搜索审批流程

```
需要搜索图片
       │
       ▼
┌──────────────────────────────┐
│      图片搜索审批对话框       │
├──────────────────────────────┤
│ 搜索词: "AI大模型架构图"     │
│ 用途: PPT第3页配图           │
│                              │
│ Y: 执行搜索                  │
│ N: 使用默认占位图  ◀── 降级   │
│ E: 编辑搜索词                │
└──────────────────────────────┘
       │
       ▼ (用户选择N)
┌──────────────────────────────┐
│ 返回降级信息:                │
│ {                           │
│   "status": "fallback",     │
│   "message": "使用默认图片"  │
│ }                           │
└──────────────────────────────┘
```

### 交互模式

**交互模式（默认）**：
```bash
export INTERACTIVE=true  # 默认值
./ppt-agent.exe
```
- 用户手动审批工具调用
- 图片搜索可选择使用默认图

**自动模式**：
```bash
export INTERACTIVE=false
./ppt-agent.exe
```
- 所有工具自动批准
- 图片搜索自动降级为默认图片

### 审批选项

| 选项 | 说明 | 适用场景 |
|------|------|---------|
| Y / YES | 批准执行 | 确认操作 |
| N / NO | 拒绝执行 | 取消操作 |
| E / EDIT | 编辑后执行 | 修改参数 |
| Q / QUIT | 退出程序 | 终止任务 |

## eino ADK 核心用法

### 1. Skill 加载与注入

```go
import (
    "github.com/cloudwego/eino-ext/adk/backend/local"
    "github.com/cloudwego/eino/adk/middlewares/skill"
    "github.com/cloudwego/ppt-agent/pkg/agent"
)

// 创建 filesystem backend（参考 eino-examples 方式）
be, _ := local.NewBackend(ctx, &local.Config{})

skillBackend, _ := skill.NewBackendFromFilesystem(ctx, &skill.BackendFromFilesystemConfig{
    Backend: be,
    BaseDir: skillsDir,
})

// 加载 SKILL.md 内容，注入到 prompt
loadedSkills, _ := agent.LoadSkillsFromDir(ctx, skillsDir)
skillsContent := agent.FormatSkillsForPrompt(loadedSkills)
```

### 2. Plan-Execute-Replan 编排

```go
entryAgent, _ := planexecute.New(ctx, &planexecute.Config{
    Planner:       planAgent,
    Executor:      executeAgent,
    Replanner:     replanAgent,
    MaxIterations: 20,
})
```

### 3. 带审批的 Tool

```go
import "github.com/cloudwego/ppt-agent/pkg/tools/wrapper"

searchTool := tools.NewSearchTool()

approvableTool := &wrapper.SearchApprovableTool{
    InvokableTool: searchTool,
    UsageScenario:  "PPT内容搜索",
    FallbackOption: "使用已有内容",
}

agent, _ := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    ToolsConfig: adk.ToolsConfig{
        ToolsNodeConfig: compose.ToolsNodeConfig{
            Tools: []tool.BaseTool{approvableTool},
        },
    },
})
```

### 4. 人机交互循环

```go
hm := human.NewManager(interactive) // true: 交互模式, false: 自动模式
iter := runner.Query(ctx, query, adk.WithCheckPointID("task-1"))

event, err := hm.RunWithApproval(ctx, runner, "task-1", iter)
```

## 编译运行

### 后端编译

```bash
cd backend
go mod tidy
go build -o ppt-agent.exe .
./ppt-agent.exe
```

### 环境变量

```bash
# API 配置
export ARK_API_KEY=your_api_key
export ARK_MODEL=your_model_name

# 交互模式（可选，默认 true）
export INTERACTIVE=true
```

### 前端运行

```bash
cd frontend
npm install
npm run dev
```

## 版本信息

| 依赖 | 版本 |
|------|------|
| Go | 1.25.0 |
| eino | v0.8.8 |
| eino-ext/adk/backend/local | v0.2.1 |
| python-pptx | >= 0.6.21 |
| Pillow | >= 9.0.0 |
