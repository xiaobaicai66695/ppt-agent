## Why

预设模板填充时，模型可能把 `content_plan.elements[].items` 返回为对象数组，而后端只接受字符串数组，导致整个创建流程解析失败并停在填充阶段。模板选择同时分散在首页、编排页和任务会话框，用户已经选过模板后仍需重复决策，也缺少根据主题自动完成视觉策略选择的可靠入口。

## What Changes

- 让 `content_plan` 输入兼容字符串、标量和常见结构化对象，并在进入 Agent/生成器前归一化为稳定字符串字段。
- 取消创建任务前独立的模板填充模型调用；主 Agent 在同一生成流程中接收空模板或已有文字的模板，补齐空字段并保留用户已有内容。
- 在提示词首页动态展示全部可用预设模板，并移除新任务会话框中的重复模板选择。
- 预设模板从首页提交后直接创建生成任务；自定义编排作为独立选项进入编排工作区，由用户自行调整页面结构。
- 新增“智能推荐”选项，根据用户主题从真实模板、配色和背景资源中选择一组可解释、可执行的生成策略。
- 将预设及推荐选择交给后端解析，避免前端复制模板 outline 后产生契约漂移。

## Capabilities

### New Capabilities
- `content-plan-input-compatibility`: 对模型和前端传入的 content plan 做宽容解析、确定性归一化和聚焦错误报告。
- `template-first-creation`: 在提示词首页一次性完成模板选择，预设直接生成，自定义编排单独进入编辑器。
- `topic-template-recommendation`: 根据用户主题推荐真实存在的模板、配色和背景使用策略，并将决策应用到任务 outline。

### Modified Capabilities

无。

## Impact

- 后端：`pkg/agent/deep` 的 outline JSON 契约和主 Agent prompt、`pkg/web` 的任务创建与模板 API、`pkg/templates` 的预设/主题/背景读取。
- 前端：`HomePage.vue`、`DashboardPage.vue`、`ConversationComposer.vue`、`ComposePage.vue`、`api.ts` 及相关单元测试。
- API：任务创建请求新增模板选择字段；现有 `query`/`outline` 请求保持兼容。
- 部署：完成后重新构建 Linux 后端与前端，并覆盖 `/ppt/ppt-agent` 服务目录。
