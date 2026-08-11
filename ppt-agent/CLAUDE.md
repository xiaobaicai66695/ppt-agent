# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在本代码库中工作时提供指导。

## 编译运行

```bash
# 后端 (Go 1.25.0)
cd backend
go mod tidy
go build ./...          # 验证编译
go build -o ppt-agent . # 编译二进制

# 前端 (Vue 3 + TypeScript + Vite)
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
export AGENT_MODE=planner           # 默认 Planner + renderer workflow
export INTERACTIVE=false           # 跳过人机交互确认
export PLANNER_CONCURRENCY=3       # 最大并行幻灯片生成数 (默认 5)
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

### 执行模式

当前默认链路是 `AGENT_MODE=planner`：先做意图分类和模板/背景推荐，再由 `PPTPlanner` 生成强 schema 的 `tasks.json`，最后由 Eino workflow 的 renderer worker pool 按 `task_id` 并发调用 Python 生成器。系统不再使用 prebuilt 多子代理架构，也不再保留串行 plan-execute-replan 作为运行路径。

### QA 质检开关

QA 质检可通过 `ENABLE_QA` 环境变量控制（默认 `true`）：

- `ENABLE_QA=true`：启用 Reviewer 和 Fixer 子 Agent，进行完整的视觉质量审查
- `ENABLE_QA=false`：跳过 QA 步骤，直接将所有任务标记为 `done`

相关代码：
- `.env` 文件中的 `ENABLE_QA` 配置
- `pkg/agent/deck/types.go` 中的 `PPTTaskConfig.EnableQA` 字段
- `pkg/prompts/planner/master_instruction.tmpl` 中的 Planner 阶段约束

### 模型 Fallback 链

`FallbackChatModel` (`pkg/agent/utils/model.go`) 包装多个 ARK 模型，从环境变量加载：
- `ARK_MODEL` → `ARK_MODEL_BACKUP1` → `ARK_MODEL_BACKUP2` → `ARK_MODEL_BACKUP3` → `ARK_MODEL_BACKUP4`
- 触发 429 限流时：失败模型暂停 30 秒，继续尝试下一个
- 所有模型失败 → 返回错误
- Fallback 链按职责创建。当前主链路由 Planner 使用工具调用模型，QA 模型通过 `QAModelFn` 按需获取。

### Token 追踪

`TokenTracker` (`pkg/agent/utils/token_tracker.go`) 通过原子计数器累计 LLM token 使用量。它附加在任务创建时的 context 上，由回调处理器（主模型调用）和压缩器（摘要调用）更新。压缩器的摘要 token 追踪使用粗略估算，因为摘要调用绕过了回调系统。

### 上下文压缩

`ChatModelCompressor` (`pkg/agent/utils/compressor.go`) 包装 Planner 的聊天模型。当前阈值由 Planner 创建时配置，策略是保留系统指令、结构化摘要和最近对话。压缩通过专用 fallback 文本模型运行。

### 任务生命周期

Planner 将 `tasks.json` 写入工作目录。每个任务经历以下状态：

```
pending → generating → done → qa_done → fixed
```

渲染工作池在调用 `generate_slide` 前设置 `generating`，生成成功后设置 `done`。

`TasksManifest` 上的辅助方法（`NeedsFix()`、`PendingTasks()`、`DoneTasks()`）驱动编排循环。QA 结果存储在每个任务的 `qa_report` 字段中。每张幻灯片最多修复 2 次。

状态常量定义在 `pkg/agent/deck/types.go`。`WriteTasksManifest` 函数合并新任务与现有状态（写入不覆盖进行中的状态）。

### Visual QA 流水线

QA 默认不在低成本生成链路中执行；如后续重新启用，应复用底层流水线：
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

Planner prompt 位于 `pkg/prompts/planner/master_instruction.tmpl`。Prompt 要保持结构化、短路径、少歧义，避免把具体坐标、字号和底层绘制细节交给 LLM。

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
