## Context

`/api/messages` 当前按单条消息路由，只有确认 `create` 后前端才调用 `/api/tasks` 创建任务。TaskRecord、ConversationMessage 和 TaskManager 已能保存任务级消息与历史，但没有“未渲染也可继续”的任务状态；PlanDraftRecord 仅保存一次规划文本，不能作为统一聊天上下文。

## Goals / Non-Goals

**Goals:**

- 每次工作台首次发送都创建可恢复 task，并将后续消息绑定到它。
- 路由器和 Planner 使用该 task 的有界历史，而不是孤立的当前消息。
- 任务在对话、澄清、规划、生成和完成之间可见地转换；仅生成阶段创建渲染 workflow。
- 保持现有已完成/运行中 PPT 的下载、预览、继续修改和权限模型。

**Non-Goals:**

- 不把通用聊天改造成跨任务的长期用户画像或记忆系统。
- 不在本次引入多用户协同、分支会话或新的消息队列。
- 不为纯 chat 强制调用 Planner、Reviewer 或渲染器。

## Decisions

### 1. Task 是统一会话主键，首次消息同步创建

`/api/messages` 接收可选 `task_id`。缺失时服务端生成 task ID、工作目录和持久化 TaskRecord，初始状态为 `conversation`，再写入第一条用户消息。后续请求必须回传该 ID，并通过现有所有权校验读取同一 task。

采用 task 而非独立 conversation 表：TaskRecord、ConversationMessage、冷加载、删除和任务列表已经以 task ID 为中心，避免两套生命周期和“草稿转任务”复制。替代方案是 conversation 表后续 promote 为 task，但会重新引入上下文丢失和双 ID 映射。

### 2. 阶段与运行状态分离

TaskInfo/TaskRecord 增加稳定的会话阶段（`conversation`、`planning`、`generating`、`completed`、`failed`），而现有运行状态继续表达可取消 workflow 的终态。纯 chat、澄清和 plan 不启动 Agent workflow，不应被启动时的 zombie 清理标为失败。

替代方案是复用 `running`：会让非渲染会话被误认为僵尸任务，也使任务列表无法说明“正在聊天”还是“正在渲染”。

### 3. 服务端构建有界上下文并返回 task_id

每次路由前，服务端读取该 task 最近若干完整 ConversationMessage，连同 task 的主题摘要/阶段组织为 router prompt context；每次响应都返回 `task_id`。创建请求进入 Planner 前，构建由完整会话摘要和当前请求组成的 Query，确保“你决定主题和风格”仍可解析为对前文的补充。

历史长度需要受限并在服务端裁剪，防止任意长 chat 占用路由模型上下文。原始消息仍持久化，以供任务详情恢复。

### 4. 前端先保存 task_id，再驱动意图专属 UI

Dashboard/Home 在第一次 `routeMessage` 响应后立即保存 task ID，并将用户消息和助手回复显示在该 task 会话里。后续所有发送都传 task ID；当路由为 `create` 时，前端调用“启动生成”而非再次 `createTask`。

为兼容直接创建 API，`/api/tasks` 保留为显式快速创建入口，但工作台统一输入不再使用它创建第二个任务。

## Risks / Trade-offs

- [每句闲聊增加 TaskRecord] → 使用轻量 `conversation` 阶段、无渲染工作目录，允许用户删除，并在任务列表明确显示。
- [旧运行状态被后台清理] → zombie 清理仅处理 `generating` 状态。
- [路由上下文成本增加] → 仅传递最近消息与确定性摘要，设置字符/轮次上限。
- [前后端发布顺序不一致] → 后端兼容缺失 task_id 的首次消息，并保留显式 `/api/tasks`；先发后端再发前端。

## Migration Plan

1. 迁移 TaskRecord 阶段字段，旧记录按已有 status 映射为完成/失败/生成阶段。
2. 部署后端：接受首次无 task_id 的消息并创建 conversation task。
3. 部署前端：持久保存并回传 `task_id`，用升级接口启动生成。
4. 冒烟验证跨轮主题委托、纯 chat、直接创建、已有 PPT 修复；清理临时任务。
5. 若出现严重回归，回滚二进制；新增字段可保留，旧记录和 `/api/tasks` 仍可工作。

## Open Questions

- 无。当前任务列表已被用户确认应承载所有工作台会话；默认只在任务标题与状态上做轻量区分。
