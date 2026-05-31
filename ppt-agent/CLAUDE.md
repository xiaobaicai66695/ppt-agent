# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在本代码库中工作时提供指导。

## 编译运行

```bash
# 后端 (Go 1.25.0)
cd backend
go mod tidy
go build ./...          # 验证编译
go build -o ppt-agent . # 编译二进制

# 前端 (React + TypeScript + Vite)
cd frontend
npm install
npm run dev             # 端口 3000，/api 代理到 localhost:8080

# 测试
cd backend
go test ./pkg/agent/utils/ -v       # 模型 fallback 单元测试
go test ./pkg/tools/qa/ -v          # QA 转换器集成测试
go test ./pkg/tools/search/ -v      # 搜索集成测试

# 独立搜索工具测试 (交互式)
cd backend
go run ./cmd/search_runner.go

# 运行时环境变量
export AGENT_MODE=deep              # "deep" (并行) 或空 (串行 plan-execute)
export INTERACTIVE=false           # 跳过人机交互确认
export DEEP_AGENT_CONCURRENCY=3    # 最大并行幻灯片生成数 (默认 5)
export COZELOOP_API_TOKEN=         # 可选: CozeLoop 可观测性
export COZELOOP_WORKSPACE_ID=      # 可选: CozeLoop 工作区
export LOG_FILE="./logs/app.log"   # 日志文件路径
export LOG_ANALYSIS_IDLE_INTERVAL="5m"  # 空闲分析间隔 (0 禁用)
export STREAM_TIMEOUT="3m"         # 单次 LLM 流式调用超时 (0 禁用)
export ENABLE_QA="true"            # 是否启用 QA 质检 (默认 true)
```

### 可观测性 (CozeLoop)

当设置了 `COZELOOP_API_TOKEN` 和 `COZELOOP_WORKSPACE_ID` 时，程序注册 CozeLoop 回调处理器进行链路追踪。本地 `LogHandler` (`pkg/callback/callback.go`) 始终注册——它记录结构化的 `[<ms>] [<agent>] →/← <component>` 行，用于工具调用、LLM 调用和 Agent 转换。流式输出通过事件循环中的 `MessageStream.Copy(2)` 处理（不走回调系统）。

## 架构

### 两种执行模式

1. **DeepAgent 模式** (`AGENT_MODE=deep`，默认)：使用 `eino prebuilt/deep`。主 Agent (`PPTTaskDeepAgent`) 将任务委托给三个子 Agent——`SlideExecutor`（通过 python-pptx 生成幻灯片）、`Reviewer`（多模态视觉 QA）、`Fixer`（修复 QA 发现的问题）。主 Agent 写入 `tasks.json` 并编排整个流水线。幻灯片可并行生成（由 `DEEP_AGENT_CONCURRENCY` 控制）。

2. **Plan-Execute-Replan 模式**（传统，串行）：使用 `eino prebuilt/planexecute`。三个 Agent 顺序运行——`Planner` → `Executor` → `Replanner`——循环直到所有幻灯片完成。每张幻灯片逐一生成，每张后接 QA。

两种模式共享相同的工具实现、Skill 注入和模型配置。Prompts 在两种模式间有约 70% 的重叠。

### QA 质检开关

QA 质检可通过 `ENABLE_QA` 环境变量控制（默认 `true`）：

- `ENABLE_QA=true`：启用 Reviewer 和 Fixer 子 Agent，进行完整的视觉质量审查
- `ENABLE_QA=false`：跳过 QA 步骤，直接将所有任务标记为 `done`

相关代码：
- `.env` 文件中的 `ENABLE_QA` 配置
- `pkg/agent/deep/types.go` 中的 `PPTTaskConfig.EnableQA` 字段
- `pkg/agent/deep/agent.go` 中的 `isQAEnabled()` 辅助函数和条件创建 Reviewer
- `pkg/prompts/deep/master_instruction.tmpl` 中的 `{{if .EnableQA}}` 条件渲染

### 模型 Fallback 链

`FallbackChatModel` (`pkg/agent/utils/model.go`) 包装多个 ARK 模型，从环境变量加载：
- `ARK_MODEL` → `ARK_MODEL_BACKUP1` → `ARK_MODEL_BACKUP2` → `ARK_MODEL_BACKUP3` → `ARK_MODEL_BACKUP4`
- 触发 429 限流时：失败模型暂停 30 秒，继续尝试下一个
- 所有模型失败 → 返回错误
- Fallback 链每个 Agent 独立创建（Planner、Executor、SlideExecutor、Fixer 各有一个实例）。QA Reviewer 通过 `QAModelFn` 获取独立模型。

### Token 追踪

`TokenTracker` (`pkg/agent/utils/token_tracker.go`) 通过原子计数器累计 LLM token 使用量。它附加在任务创建时的 context 上，由回调处理器（主模型调用）和压缩器（摘要调用）更新。压缩器的摘要 token 追踪使用粗略估算，因为摘要调用绕过了回调系统。

### 上下文压缩

`ChatModelCompressor` (`pkg/agent/utils/compressor.go`) 仅包装 DeepAgent 编排器的聊天模型（不包装子 Agent）。双触发条件：`MessageThreshold=12` 或 `TokenThreshold=30000`（估算字符数）。策略：`system + [压缩摘要] + [最近 4 轮对话]`。压缩通过专用 fallback 模型运行，`MaxTokens=4096`。

### 任务生命周期 (DeepAgent 模式)

主 Agent 将 `tasks.json` 写入工作目录。每个任务经历以下状态：

```
pending → generating → done → qa_done → fixed
```

主 Agent 在将任务分发给 SlideExecutor 时设置 `generating`，在执行器报告成功后设置 `done`。`generating` 状态在 Agent prompt 指令中设置（Go 类型中不强制）。

`TasksManifest` 上的辅助方法（`NeedsFix()`、`PendingTasks()`、`DoneTasks()`）驱动编排循环。QA 结果存储在每个任务的 `qa_report` 字段中。每张幻灯片最多修复 2 次。

状态常量定义在 `pkg/agent/deep/types.go`。`WriteTasksManifest` 函数合并新任务与现有状态（写入不覆盖进行中的状态）。

### Visual QA 流水线

根据执行模式有两种 QA 模式：

- **DeepAgent 模式**：Reviewer 子 Agent 使用 `single_qa_review` (`pkg/tools/qa/qa_tool.go`)——每次一页，由主 Agent 在每个任务完成后调用。
- **Plan-Execute-Replan 模式**：所有幻灯片生成后使用批量 QA。

两者共享相同的底层流水线：
1. 运行 `pptx_qa_converter.py`，调用 LibreOffice（PPTX→PDF）然后 pdftoppm（PDF→JPEG，150 DPI）
2. 查找与请求的 PPTX 文件名匹配的图像
3. 将图像 + system prompt 发送给多模态 LLM（通过 `modelFn`）
4. 解析响应中的 `high`/`medium`/`low` 严重级别问题
5. 合并结果到 `.qa_result.json`；在 `.qa_attempts.json` 中追踪尝试次数（每张幻灯片最多 2 次）

QA 结果通过 `||` 做累加合并——一旦 `HasHighIssue` 变为 true，就永远不会变回 false。

### 人机交互搜索审批

搜索工具通过 `InvokableSearchApprovalTool` (`pkg/tools/human_in_the_loop.go`) 包装。第一次调用时，工具使用 `SearchApprovalInfo` 调用 `StatefulInterrupt`。`human.Manager` 循环捕获中断，提示用户（1=跳过，2=确认，3=编辑查询），然后使用 `ResumeWithParams` 恢复。在非交互模式下，所有搜索默认为跳过。

### 关键硬编码路径

- **Python 二进制**：`/root/pptx_env/bin/python`——硬编码在 `python_runner.go:83` 和 `qa_tool.go:170`。项目仅在 Linux 上运行，python-pptx 安装在此确切路径。
- **转换器搜索**：`qa_tool.go` 使用相对路径向上搜索 8 级父目录查找 `pptx_qa_converter.py`，回退到 `PROJECT_ROOT` 环境变量。

### Skill 注入模式

Skills 从 `skills/` 目录（`SKILL.md` 文件）通过 `LoadSkillsFromDir` → `FormatSkillsForPrompt` 加载，然后作为 `{skills}` 模板变量注入到 Planner/Executor prompts 中。Eino 的 skill middleware 从 `eino/adk/middlewares/skill` 初始化但从未实际使用——skills 作为文本手动注入。

### Prompt 字符串模式

`slide_executor.go` 和 `planexecute/executor.go` 使用包级常量 `bt`（`"\`"`) 在 Go 原始字符串字面量中嵌入反引号。当向 prompt 字符串添加 markdown 代码格式（反引号）时，使用 `+ bt +` 连接——永远不要在原始字符串字面量中嵌套反引号。

### 后台日志分析

后台服务 (`pkg/log_analysis/service.go`) 监控系统日志和任务失败：

- **空闲分析**：当没有任务运行时（可通过 `LOG_ANALYSIS_IDLE_INTERVAL` 配置间隔），从 `LOG_FILE` 读取最后 300 行并发送给 LLM 分析。
- **失败分析**：任务失败时立即读取最后 300 行并触发 LLM 分析。

结果存储在 `task_error_analyses` DB 表中（`pkg/db/db.go`），包含字段：`analysis`（LLM 摘要）、`root_cause`、`suggestion`。通过以下接口查询：
```
GET /api/log-analyses              # 最近 50 条分析
GET /api/log-analyses/task/:task_id  # 特定任务的分析
```

LLM 分析器使用 `read_file` 工具动态读取相关 prompt 模板和 Python 生成器源码，工具路径由 `prompts/log_analysis/analyzer_instruction.tmpl` 中的 `{{ .SkillsDir }}` 占位符指定。LLM 通过 ReAct 循环自主决定何时调用 `read_file` 工具获取额外上下文。

必需的环境变量（配置在 `.env` 中）：
- `LOG_FILE`：要读取的日志文件路径（例如 `./logs/app.log`）
- `LOG_ANALYSIS_IDLE_INTERVAL`：空闲分析间隔（例如 `5m`，`0` 禁用）
- `STREAM_TIMEOUT`：单次 LLM 流式调用可阻塞的最长时间（例如 `3m`，`0` 禁用）。超时时任务退出并返回超时错误，以便取消和恢复。

### 用户偏好学习

`pkg/agent/learning/` 模块自动分析用户历史任务并提取风格偏好：
- 偏好收集器 (`collector.go`) 从完成的任务中提取配色、布局、语言风格等信息
- 偏好分析器 (`analyzer.go`) 调用 LLM 总结偏好
- 偏好更新器 (`updater.go`) 将分析结果存储到 `user_style_profiles` 表
- 学习引擎 (`engine.go`) 协调整个学习流程
- 偏好数据用于在 `PPTTaskConfig.StyleContext` 中注入个性化上下文

### 路由引擎

`pkg/agent/router/engine.go` 根据任务特征（查询内容、用户偏好、当前负载）决定将任务分配给哪个执行模式。
