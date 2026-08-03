## Context

当前创作路径同时存在首页模板卡片、ComposePage 预设选择和 Dashboard 会话框模板下拉。首页只硬编码展示六个模板，真正创建任务时前端又复制 preset outline 并调用后端补全；模型一旦把 `items` 返回为对象，`json.Unmarshal` 会使整个请求失败。服务器已采用 Linux 单目录部署，本次变更需要保持旧 `query`/`outline` API 兼容并能直接滚动部署。

## Goals / Non-Goals

**Goals:**

- 模型返回常见结构变体时仍能稳定得到可消费的 `content_plan`。
- 用户只在提示词首页选择一次智能推荐、预设模板或自定义编排。
- 预设和智能推荐由后端基于实时模板目录创建 outline，避免前端契约漂移。
- 推荐模板、配色和背景策略可解释、低成本且只引用真实资源。

**Non-Goals:**

- 不恢复默认 QA/Reviewer。
- 不为模板推荐增加一次 LLM 调用。
- 不让推荐系统修改用户明确选择的预设或自定义 outline。
- 不在本轮建设在线模板市场或持久化自定义模板库。

## Decisions

### 1. 在类型边界归一化 content plan

为 `ContentElement` 增加自定义 JSON 解码：字符串直接保留；数字、布尔值转为文本；对象按 `title/label/name/key` 与 `text/description/value/content` 等常见字段组合；无已知字段时按稳定键序列化为简短 `key: value` 文本。归一化后的公开字段仍为 `[]string`，避免把结构变体扩散到 prompt、任务清单和 Python 生成器。

替代方案是把 `Items` 改成 `[]any`，但这会把类型判断推给所有消费者并扩大回归面，因此不采用。

### 2. 任务创建只准备模板结构，内容补齐归主 Agent

任务创建 API 增加可选 `template_selection`：`recommended` 或 `preset`。`outline` 仍拥有最高优先级以兼容 ComposePage 和旧客户端；`preset` 由服务端 loader 克隆结构；`recommended` 先计算推荐再构造结构；未提供选择时保持原有自由生成流程。

任务创建阶段只做 schema 解析、合法 `content_type/theme/background` 校验和空值归一化，绝不调用 `generateOutlineSlides`。模板 outline 带有 `content_mode`：服务端预设为 `template_scaffold`，表示模板标题/描述仅是结构提示，主 Agent 必须围绕用户 query 重写；自定义编排为 `user_outline`，表示非空文字是用户约束，主 Agent只补齐空字段。两种模式都在同一个主 Agent 运行中完成规划和生成，不新增独立“填充模板”步骤。

### 3. 推荐采用确定性元数据评分

推荐器对 query 与预设 `display_name/category/tags/description` 做归一化关键词评分，并用业务关键词映射补充类别权重；无显著命中时回退 `generic`。配色优先使用预设 `default_palette` 并经过主题目录校验。背景策略根据“正式/数据密集/极简”等抑制词和“国风/党政/活动/旅行/艺术/发布”等视觉词判断是否启用，再从真实背景场景元数据中评分选择；背景只应用于封面、章节和结束类页面。

推荐响应记录模板、主题、背景、是否启用及简短原因，便于测试和后续可观测展示。推荐只构造模板结构和视觉字段，页面内容仍由主 Agent 在生成流程中完成。

### 4. 首页成为唯一的新任务模板入口

HomePage 从 `/api/templates` 动态读取全部预设，在模板网格前放置“智能推荐”，末尾放置“自定义编排”。智能推荐和预设点击提交后直接创建任务并跳转 Dashboard；自定义编排进入 ComposePage，页面编辑器不再要求先选整套预设。Dashboard 的统一会话框继续负责新建自由任务及续聊，但移除模板复选框和下拉框，防止第二次选择。

## Risks / Trade-offs

- [对象内容无法无损映射为字符串] -> 优先提取语义字段并为未知对象提供稳定的键值文本，同时增加对象、标量和混合数组测试。
- [关键词推荐不如模型灵活] -> 返回推荐原因并保留用户直接选择所有模板的能力；通过模板元数据扩展评分而非增加在线模型成本。
- [模板较多导致首页过长] -> 使用响应式缩略图网格和分类信息，保持所有模板可见且不引入第二层选择弹窗。
- [旧前端仍发送 outline] -> 保留 outline 优先级和现有 API 字段，服务端新增字段完全可选。

## Migration Plan

1. 先部署兼容解析、主 Agent 双模式模板处理和扩展后的后端 API，旧前端仍可工作。
2. 部署动态模板首页和简化后的 Dashboard/ComposePage。
3. 通过聚焦 Go 测试、前端单测/构建和桌面/移动浏览器检查后覆盖 Linux 服务。
4. 若部署失败，恢复上一版二进制与前端 dist；数据库无需迁移。

## Open Questions

无。推荐规则保留为独立纯函数，后续可根据真实使用数据调整权重。
