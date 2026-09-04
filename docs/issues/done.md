# 已完成重要事项归档

本文件只保留仍能说明当前能力基线、架构决策或交付方式的完成事项。一次性 UI 微调、流式拼接修复、发布纠偏和同一主线的局部回归不再逐条长期归档；它们的代码证据保留在 Git 历史与对应迭代记录中。

未完成或尚未上线的工作不得写入本文件，应登记在 [`todo.md`](./todo.md)。长期决策见 [`docs/decisions/`](../decisions/README.md)。

## Agent 编排与运行可靠性

| 合并事项 | 覆盖历史 ID | 完成时间 | 沉淀 |
| --- | --- | --- | --- |
| Runtime harness、可回放执行轨迹与质量评估基线 | `PPT-HARNESS-001`、`PPT-EVAL-001`、`PPT-OBS-001`、`PPT-EVAL-002` | 2026-07-31 至 2026-09-02 | `RuntimeMeta`、持久化运行事件、Timeline、质量评估集与 test/validation 隔离；决策见 `001-agent-harness-observability.md`。 |
| Planner → Reviewer → Commit → 并发渲染主流程 | `PPT-FLOW-001/002`、`PPT-PLAN-001/002/006`、`PPT-COMPONENT-001` | 2026-08-04 至 2026-08-27 | `tasks.draft.json` 与正式 `tasks.json` 分离，确定性审查后原子提交，组件级 DeckSpec 驱动确定性生成器；架构入口见 `docs/architecture/ppt-agent-current-architecture-summary.md`。 |
| 任务交付与恢复闭环 | `PPT-DELIVERY-001`、`PPT-TRACE-002`、`PPT-DEPLOY-002` | 2026-08-03 至 2026-08-29 | 任务权限、SSE、渐进预览、文件对账、下载与会话恢复构成同一交付闭环；决策见 `003-generation-delivery-flow.md`。 |
| 服务运行与部署基线 | `PPT-OPS-001`、`PPT-RELEASE-001` | 2026-08-04、2026-08-29 | MySQL 与应用同机运行；运行变更完成本地验证、Linux 构建、部署、最小冒烟与证据回填；决策见 `005-ops-governance-reliability.md`。 |
| Web / Task / Model 运行时边界重构 | `PPT-RESILIENCE-001` | 2026-09-04 | 运行时代码归入 `pkg/runtime/{web,task,model}`，并按职责增加二级导航目录；HTTP、SSE、任务和模型契约保持不变。部署与验证证据见 [`2026-09-04-runtime-module-boundaries.md`](../迭代记录/2026-09-04-runtime-module-boundaries.md)。 |

## DeckSpec、视觉与素材能力

| 合并事项 | 覆盖历史 ID | 完成时间 | 沉淀 |
| --- | --- | --- | --- |
| 组件优先的 PPT Deck Planner 与视觉质量路线 | `PPT-QUALITY-001`、`PPT-QUALITY-011/012`、`PPT-COMPONENT-001`、`PPT-SKILL-001` | 2026-08-04 至 2026-09-02 | LLM 规划语义组件与容量，Python 生成器控制排版、对比度与溢出；组件契约是 Planner、后端和生成器的共同边界。 |
| 可追溯的图片搜索、物化与页面背景策略 | `PPT-ASSET-002`、`PPT-BG-001/002/004`、`PPT-SKILL-002/003/004/005/006/007` | 2026-08-21 至 2026-09-03 | 图片检索词在规划期确定，素材落盘至任务工作区，背景按页面类型复用并在渲染前校验；不再依赖旧 Visual Designer 本地背景主题。 |
| 独立 Deck Planner skill 的运行边界 | `PPT-SKILL-001/002/003` | 2026-08-28 至 2026-08-31 | Skill 负责平台无关的视觉契约、素材物化和生成器验收；后端 Planner 只写视觉意图，避免模型、下载器和生成器职责重叠。 |

## 工作台、会话与用户交付

| 合并事项 | 覆盖历史 ID | 完成时间 | 沉淀 |
| --- | --- | --- | --- |
| PPT 生产工作台重建 | `PPT-UI-001`、`PPT-UI-007`、`PPT-UI-005/006` | 2026-08-03 至 2026-09-03 | 首页、编排、任务、预览、认证与管理页统一为生产工作台；真实模板资源、明确状态和响应式操作路径优先。 |
| 统一会话、意图路由与聊天资料能力 | `PPT-UI-002`、`PPT-CHAT-001/003`、`PPT-SEARCH-001` | 2026-08-29 至 2026-09-04 | 同一用户任务承载对话、澄清和生成启动；聊天可使用受边界保护的网页/图片资料，来源和工具结果以可见、安全的形式呈现。 |
| 已完成 PPT 的反馈与运营可见性 | `PPT-DELIVERY-002`、`PPT-OPS-002` | 2026-09-02 至 2026-09-03 | owner-only 交付反馈、管理聚合指标和受保护管理 API；不输出 Key 原文或推断后端未提供的指标。 |

## 模型、账号与安全边界

| 合并事项 | 覆盖历史 ID | 完成时间 | 沉淀 |
| --- | --- | --- | --- |
| 用户偏好与模型供应商兼容层 | `PPT-PREFERENCE-001/002`、`PPT-MODEL-001` | 2026-08-05 至 2026-08-26 | 停用自动偏好学习，只使用显式画像与场景门控；模型通过 provider profile、账户级自备 Key 与 fallback 解耦。 |
| 认证与账号安全 | `PPT-AUTH-001` | 2026-09-03 | 注册与验证码登录分离；密码合规校验和 bcrypt 存储；未注册验证码登录不能隐式创建账户。 |

## 归档规则

- `small` 的单点回归、样式调整、SSE 分段/链接修复、单次发布纠偏和临时运维细节不再逐行保留；若它们改变了长期行为，则被合并到上表对应能力基线或决策文档。
- `medium`/`large` 事项仅在形成独立能力、跨模块边界或长期运行决策时保留；纯兼容或一次性迁移不单独保留。
- 2026-09-04 的 `PPT-RESILIENCE-001` 已完成本地验证、Linux 构建与线上健康检查；实现与证据见 [`2026-09-04-runtime-module-boundaries.md`](../迭代记录/2026-09-04-runtime-module-boundaries.md) 及 OpenSpec 变更目录。
