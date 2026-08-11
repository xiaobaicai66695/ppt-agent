## Why

PPT 生成目前同时存在规划质量和过程可见性两类断点：智能推荐机械继承固定模板页数与示例内容，背景和正文密度不足；运行轨迹又会丢失早期工具调用，压缩发生时用户无法判断保留了什么。需要把推荐、规划、压缩与前端事件展示收敛为可验证的结构化契约。

## What Changes

- 将“智能推荐”改为基于 LLM 意图结果选择模板风格、配色、背景主题和建议页数，由主 Agent 围绕用户主题动态规划页面，不再复制 preset 的固定页清单。
- 为背景规划增加页面语义、信息密度、可读性和整套视觉节奏目标，使主 Agent 主动为适合的页面选择真实可用背景，并由代码校验背景引用。
- 将模板容量从过紧的逐项硬字数限制调整为按版式定义的目标密度区间和可渲染上限，充分利用生成器已有的动态字号与溢出保护。
- 精简模型侧运行 metadata 与 manifest 验证事件，只传递总页数和已生成页数；丰富诊断信息保留在运行时可观测快照中，不占用模型上下文。
- 完整保留并合并 SSE 与持久化 ToolCall 事件，确保 `read_file`、`edit_file`、manifest 工具、搜索与 Python 等调用都可回看。
- 将上下文压缩作为独立事件呈现，提供消息数、token 数、节省量和保留要求的前后差异；压缩摘要始终以原始用户请求及后续明确约束为锚点。

## Capabilities

### New Capabilities
- `ppt-agent-generation-planning`: 定义智能推荐、动态页数、背景覆盖策略、内容密度和规划契约。

### Modified Capabilities
- `ppt-agent-runtime-harness`: 收敛模型侧进度 metadata，记录以用户意图为锚点的压缩结果与结构化压缩差异，并保证完整工具事件可持久化回放。
- `ppt-agent-developer-status-ui`: 合并实时与持久化事件历史，展示完整 ToolCall 类型和独立压缩事件详情。

## Impact

- 后端：DeepAgent 配置与 prompt、推荐模式、意图分类复用、RuntimeMeta、上下文压缩、SSE/会话事件接口。
- 前端：Dashboard 时间线事件归并、事件分类和压缩详情组件。
- Visual Designer：单页模板容量 contract、生成器参考文档及相关校验。
- 接口兼容性：沿用现有任务创建和 SSE 事件类型；新增结构化压缩 metadata 字段，客户端应容忍缺失字段。
