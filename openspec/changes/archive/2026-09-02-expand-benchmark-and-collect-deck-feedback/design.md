## Context

`pptbench` 已按 router、planner、reviewer、fixer 分 suite，并将 test 与 validation 分目录保存；但目前每个 suite 的数量不一致且偏少。Dashboard 已在任务完成后展示下载与预览，`continue` 请求会把自由文本写入日志，却不是可查询、不可歧义的交付评价。

## Goals / Non-Goals

**Goals:**

- 让每个 suite 在 test 和 validation 数据集各稳定拥有 10 条具名 case，并用离线测试守住这一基线。
- 让已完成任务的 owner 能在交付页左下角提交 1–5 分和可选建议，重载后仍可见，并能修正自己的评价。
- 将反馈与继续修改分离：反馈不能启动 Agent、改写 Deck 或污染运行状态。

**Non-Goals:**

- 不运行新增 benchmark 的真实 LLM 全量评测；本次只校验数据结构和加载契约。
- 不新增管理员反馈运营后台、跨用户聚合报表、页面级评分或公开评论。
- 不将未完成、失败或非 owner 任务暴露为可提交反馈。

## Decisions

### 1. Test 与 validation 均按 suite 补足至十条

每个数据集的 `router`、`planner`、`reviewer`、`fixer` 均以 10 条作为下限，validation 与 test 保持题材、数值、措辞和修复目标独立。选择双集齐平而非只扩 test，是为了避免 test 的扩容掩盖 holdout 过小的问题。

### 2. 以 loader 契约测试守护样本基线

在 `cmd/pptbench` 增加无需模型 Key 的测试，使用生产 loader 验证每个 dataset/suite 的数量、非空 ID 和全局无重复 ID。相比 README 人工约定，这能在 CI 和本地修改时立即发现删漏或错误合并。

### 3. 反馈独立持久化，按任务和 owner 唯一

新增 `TaskFeedback` 记录，唯一键为 `(task_id, user_id)`，包含 `rating`、`suggestion` 和更新时间。任务 owner 使用幂等 upsert 更新自己的记录；反馈在读取 task 时一并返回。选择独立表而不是复用会话文本，保证评分可查询、验证且不触发继续工作流。

### 4. 仅在已交付状态显示紧凑左下角组件

Dashboard 将评分组件固定在工作台左下角，仅对 `completed`、`waiting_confirmation` 或存在最终文件的任务显示。评分使用五个明确的可访问按钮，建议为可选文本框；提交成功后切为“已记录，可修改”，提交失败保留用户输入。位置贴近交付结果且不遮挡预览/下载主操作。

## Risks / Trade-offs

- [数据集扩容引入 schema 不合法 case] → 为每类新增 fixture 测试并复用现有 loader。
- [用户把建议当修改请求] → 文案明确“仅用于改进，不会自动修改本次 PPT”，并保持与继续编辑输入分离。
- [反馈表升级影响已有部署] → GORM AutoMigrate 创建新表；旧任务没有反馈时返回空值。
- [移动端左下角遮挡内容] → 使用窄屏媒体规则改为正常文档流底部条。

## Migration Plan

1. 增加 benchmark fixture 与离线数量测试。
2. 增加反馈模型、迁移、owner-only API 和后端测试。
3. 增加 Dashboard 反馈组件和 API 类型，构建前端。
4. 构建 Linux 后端与前端交付物，部署后以一个低成本完成任务验证读取和提交，再清理冒烟数据。
5. 回滚时可下线前端入口和路由；已写反馈记录不会影响任务或交付文件。

## Open Questions

- 无。本期采用任务级单评分；后续如需页面级诊断，可在不改变此表语义的前提下新增明细表。
