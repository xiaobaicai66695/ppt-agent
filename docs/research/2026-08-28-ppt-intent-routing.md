# PPT Agent 意图路由预研

## 背景

当前首页和工作台输入框与 PPT 创建流程耦合过紧。用户输入闲聊、概念解释、先规划或修复已有 PPT 时，容易走到新建演示入口，造成误创建、误渲染和高成本任务。

目标行为是让输入先经过后端统一意图路由，再由前端根据路由结果展示、切换模式或执行动作。

## 目标与非目标

目标：

- 建立后端统一意图契约：`chat | create | plan | fix`。
- 前端不再自行猜测是否创建任务，只消费后端路由结果。
- `chat` 不创建任务；`plan` 不渲染 PPT；`fix` 必须绑定已有任务；`create` 才允许进入创建链路。
- 闲聊回答可接入现有 web search 和图片搜索能力，能力未配置时明确降级。
- 保留现有 Compose、Dashboard、SSE、缩略图和下载链路。

非目标：

- 本轮不新增完整服务端 Draft 表或多人协同草稿管理。
- 本轮不重构 Planner/Reviewer/Fixer 主生成链路。
- 本轮不引入新的第三方搜索 provider。

## 现状

- 后端已有 `/api/tasks` 创建入口，并有 `request_router.go` 对创建请求做 `create_deck/fix_existing/clarify_topic/chat` 分类。
- 该分类仍以“创建入口防误用”为中心，缺少 `plan`，也没有统一 message API。
- 前端首页和 Dashboard 无选中任务时直接调用 `createTask`；选中任务时直接调用 `continueTask`，导致闲聊也可能被当成修复反馈。
- 现有工具中已有 `search` 和 `search_images`，可复用给 chat 回答，但需要允许配置缺失时降级。

## 推荐方案

第一版采用“模型分类 + 规则兜底 + 服务端最终裁判”：

```json
{
  "intent": "chat | create | plan | fix",
  "mode": "chat | pptagent",
  "confidence": 0.0,
  "needs_confirmation": false,
  "normalized_request": "制作一份面向研发负责人的产品评审 PPT",
  "task_id": "",
  "missing_fields": [],
  "action": "reply | prepare_create | save_plan | update_task | ask_clarification"
}
```

落地策略：

- 新增 `/api/messages`，前端所有自由输入先走该入口。
- `chat/reply`：直接返回普通回答，不创建任务；必要时尝试 web search / image search。
- `create/prepare_create`：前端切换到 PPT Agent；信息充分时再调用 `/api/tasks` 创建。
- `plan/save_plan`：前端进入 PPT Agent 规划状态，只展示/保存草稿，不调用生成接口。
- `fix/update_task`：已有选中任务时调用 `/api/tasks/:id/continue`；无任务时要求用户选择任务。
- `/api/tasks` 保留服务端二次校验，防止前端绕过路由层直接误创建。
- 创建入口增加短时间重复提交去重，避免同一消息重复创建任务。

## 风险与待确认问题

- 服务端已新增 `PlanDraftRecord` 和 `/api/plan-drafts` 草稿 API；后续仍需把草稿升级为完整 DeckSpec 编辑/创建链路，而不是只保存规划文本。
- `chat` 搜索能力依赖 `QIANFAN_API_KEY`、`UNSPLASH_ACCESS_KEY` 等环境配置，未配置时只能降级。
- 低置信度阈值已支持环境变量：`PPT_INTENT_LOW_CONFIDENCE_THRESHOLD`、`PPT_INTENT_CREATE_MISSING_FIELDS_AUTO_THRESHOLD`。
- 前端手动模式优先级需要继续细化：手动选择 PPT Agent 不等于绕过后端 create/fix/plan 裁判。
- 任务元数据已补充 `intent`、`conversation_id`、`source_message_id`、`parent_task_id` 字段；当前创建入口写入 `intent=create` 和 `conversation_id=task_id`，后续修复/草稿转任务时应继续写入来源链路。

## 结论

应把 PPT Agent 定义为“由意图进入的工作模式”，而不是把整个输入框默认绑定到创建演示。第一版先建立统一消息入口和四类意图边界，确保不会把闲聊、规划和修复误导到新建 PPT。

## 关联事项

- TODO: `docs/issues/todo.md#当前执行事项`
