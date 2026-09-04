# PPT Agent 命名与兼容约定

## 目标

代码优先表达业务动作，避免将内部框架术语扩散到工具、函数和用户可见运行信息。

| 旧术语 | 新代码优先术语 | 说明 |
| --- | --- | --- |
| deck | ppt | 表示整份演示文稿时使用 `ppt`。 |
| DeckSpec / manifest | plan | 表示可渲染的 PPT 页面计划时使用 `plan`。 |
| TaskItem | page | 表示单页时使用 `page`。 |
| materialize / hydrate | download / prepare | 表示下载并写入图片元数据时使用 `download` 或 `prepare`。 |
| reconcile | sync | 表示文件与页面状态对齐时使用 `sync`。 |
| patch | update | 表示模型对计划的受限修改时使用 `update`。 |

## 命名规则

- 获取数据用 `get` 或 `read`；保存用 `write`；更新用 `update`；删除用 `delete`。
- 函数注释说明输入、输出、会修改的文件、可恢复错误与恢复边界。
- 新增 Agent/Tool 名称优先使用 `PPTPlanner`、`PlanReviewer`、`PPTPageEditor` 和 `update_ppt_plan` 等直白动词短语。
- 不为一次性流程引入抽象层；固定顺序流程优先写成清晰的普通函数调用。

## 兼容边界

- `tasks.json`、`tasks.draft.json`、现有 JSON 字段、HTTP/SSE 字段和 Python 生成器契约不得因命名整理直接变更。
- 新文件名或新内部 API 必须提供旧任务读取兼容，直到历史任务不再需要恢复。
- 对外 SDK、已有路由和已存数据库值的改名需要单独迁移方案，不能作为简单格式化处理。
