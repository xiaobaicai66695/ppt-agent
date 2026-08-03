## Context

当前任务生成、任务续聊、模板编排分别由 Sidebar、Dashboard 悬浮聊天框和 ComposePage 提交，状态和输入契约割裂。任务文件同时来自 `tasks.json` progress 事件和目录扫描 `file_ready` 事件：前者常为相对文件名，后者或最终任务记录可能为绝对路径，前端按原字符串合并会产生同页双卡。会话方面，POST continue 实际返回 SSE 流，但前端按 JSON 读取；打开聊天又先清空历史，冷任务只有 `conversation_content/full_answer` 时因此显示空白。

Visual Designer 素材本已放在 skill 内，但图标解析依赖 manifest 中的 PNG 文件存在；部署遗漏、大小写或路径异常时只绘制空的浅色占位。命令执行器仍保留 Windows shell 和 Python 路径分支，与服务器 Linux-only 目标不一致。

## Goals / Non-Goals

**Goals:**

- 让同一逻辑页在任务、SSE、历史恢复和完成态中只有一个稳定身份。
- 让初始生成和后续修改共享同一消息输入、同一历史、同一 Markdown 渲染和同一事件流。
- 将自定义模板草稿从 ComposePage 带回任务工作台，由统一输入框选择并提交。
- 展示用户意图、规划快照、当前执行项和偏离告警，帮助定位 Agent 从哪一步开始跑偏。
- 保证 `ppt-agent` 目录包含运行时所需视觉资源，并只支持 Linux CLI。

**Non-Goals:**

- 不恢复默认在线 QA/Reviewer 流程。
- 不引入新的 LLM 摘要调用；短标题使用确定性文本摘要，避免增加成本和延迟。
- 不改造为通用聊天平台，也不在本次引入 WebSocket。
- 不依赖服务器目录之外的共享素材库。

## Decisions

### 1. 采用稳定的页面身份并在两端去重

后端输出文件列表在发布前按 `page_index/task_id/basename` 归一化和去重，前端仍对旧数据做兼容去重。页面以合法 `page_index` 为首选 key，缺失时使用规范化 basename，再回退 `task_id`。相比直接按绝对路径或只按标题合并，该方案既能消除路径差异，也不会把不同页的同名标题误合并。

### 2. Continue 改为受理式 POST，所有增量走任务 SSE

`POST /api/tasks/:id/continue` 对已完成任务返回 `202 accepted` 并异步启动处理；运行中仍返回 `202 queued`。回答、工具、进度、完成事件统一由现有 `GET /stream` 发布。后端为每轮 continuation 建立 answer buffer，在 `continue_complete` 时作为一条 assistant message 持久化，保留模型原始 Markdown 和换行，不按标点重分句。相比维持 POST 流式响应，这与浏览器 `EventSource`、重连和事件回放契约一致。

### 3. 统一会话数据模型，兼容历史冷数据

Dashboard 以结构化 `messages` 为主；若旧任务没有 messages，则从 `full_answer` 或 `conversation_content` 合成一条只读 assistant 历史。打开面板时保留现有内容并显示加载态，只有任务切换时才替换。Markdown 通过经过 HTML 转义的本地 renderer 输出，支持标题、列表、段落、代码和表格的基本可读样式，不执行原始 HTML。

### 4. 用一个 Composer 覆盖新建、模板生成和续聊

移除 Sidebar 的新建任务表单及 Dashboard 的独立悬浮聊天入口，在任务主区底部使用统一 `ConversationComposer`。未选任务时提交创建任务；选中已完成任务时提交续聊；选中运行任务时提交排队反馈。Composer 可选 preset 或 ComposePage 保存的自定义 outline 草稿，模板只影响下一次新任务。ComposePage 的“开始生成”改为“保存并返回生成”，将版本化草稿写入 sessionStorage 后跳转 Dashboard，避免把大纲塞进 URL 或长期写入 localStorage。

### 5. 确定性展示标题与意图对齐轨迹

任务保留完整 `query`，新增纯前端 `summarizeTaskTitle`：提取首个非空 Markdown 标题/首句，去除标记并限制 42 个可视字符，完整内容放在可展开的“查看原始需求”区域。RuntimeMeta 新增 intent anchor、plan snapshot、current slide 和 alignment warnings；TaskManager 在首次得到 manifest 时冻结规划快照，后续对页数、task id、标题、content_type 和缺失文件做确定性比较。该机制不声称语义评分，只呈现可解释的结构偏离。

### 6. 视觉资源自包含并提供有意义降级

资源根目录始终从 `asset_manager.py` 自身位置解析，启动/测试时校验 manifest 的文件存在性和大小写。图标缺失时使用带语义缩写和主题色的实心 fallback，而不是空白浅色框。生成器 smoke test 从一个临时工作目录导入 skill，模拟仅复制 `ppt-agent` 的部署方式。

### 7. CLI 明确为 Linux-only

`LocalOperator` 固定通过 `/bin/sh -c` 执行，Python 默认值固定为可配置的 Linux venv 路径；删除 `runtime.GOOS`、`cmd.exe` 和 Windows 路径分支。Linux 默认值仍可由 `PYTHON_BIN` 覆盖，便于不同服务器部署。

## Risks / Trade-offs

- [旧任务没有结构化 assistant 消息] -> 前端用 `full_answer/conversation_content` 兼容恢复，新任务从本次变更开始按轮次持久化。
- [页码被模型错误复用] -> 归一化时记录 duplicate warning，并优先保留与 manifest task 对应且文件存在的条目。
- [自定义草稿只放 sessionStorage] -> 刷新 ComposePage 前草稿仍在当前标签页；不跨设备同步，避免本次引入数据库迁移。
- [结构对齐不能判断深层语义偏离] -> UI 明确标注为“契约对齐”，保留原始需求、规划和当前页供人工判断。
- [移除 Windows 分支后本机不能完整运行后端 CLI] -> Go 单测使用可注入 shell/纯函数测试，实际运行验收放在 Linux 服务器；前端和 Python 生成器仍可在当前环境验证。

## Migration Plan

1. 先部署兼容读取逻辑、文件去重和新 RuntimeMeta 可选字段，旧前端可忽略新增字段。
2. 再部署新前端统一 Composer 和 Markdown 历史恢复。
3. 在 Linux 环境校验 `/bin/sh`、`PYTHON_BIN`、LibreOffice/Poppler 与 skill asset manifest。
4. 回滚时可整体回退前后端；新增可选字段和既有数据库消息不阻碍旧版本读取。

## Open Questions

- 无阻塞问题。自定义模板跨设备持久化和真正的模型语义对齐评分作为后续独立需求评估。
