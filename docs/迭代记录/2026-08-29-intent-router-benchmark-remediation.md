# 2026-08-29 意图路由 Benchmark 修复与上线记录

- 事项：`PPT-ROUTER-001`
- 路线：`direct`
- 所属方向：生成编排与交付闭环、评估工程治理与运行可靠性
- 状态：已上线并完成冒烟

## 问题与修复

Router benchmark 首次复跑显示测试集 `4.00/5`（3/4 通过）；独立 validation 集显示 `2.875/5`（3/8 通过）。根因有两项：

1. 创建入口把模型标记的 `page_count`、`audience`、`style` 等可选生成参数当成必须澄清，阻断了主题已经明确的新建请求。
2. benchmark validation 保留 `create_deck` / `fix_existing` 的稳定评测词汇，而线上 HTTP 消息契约已简化为 `create` / `fix`；同时成功创建响应错误地复用了 `clarification_question` 字段承载可选缺项，导致 Judge 将正确路由判为澄清。

修复内容：

- 创建入口仅在模型明确请求澄清时才阻断；可选生成参数交由 Planner 补全。
- 成功创建结果不再携带 `clarification_question`。
- `pptbench` 的创建入口适配层显式输出稳定评测 intent、实际下游 Agent 和归一化请求，保持评测跨 HTTP 契约演进可比。
- 补足 `pptbench` 运行所需的可选图片搜索禁用配置，避免 Planner benchmark 因配置字段漂移无法编译。
- 忽略本地 benchmark 运行产物，避免模型 trace 进入提交。

## 本地验证

- `go test ./cmd/pptbench ./pkg/web ./pkg/agent/deck`：通过。
- `go build ./...`：通过。
- `go run ./cmd/pptbench -s router -p all`：测试集 `4.75/5`，4/4 通过。
- `go run ./cmd/pptbench --dataset validation -s router -p all`：`4.875/5`，8/8 通过。

## 上线与冒烟

- 部署目标：`remote-dev:/ppt/ppt-agent`。
- 交付物：Linux `amd64` 二进制，SHA-256 `340ab66bb4d47529ad19a15e1b5b1ffedd28b06307f20e708090678e61ff8406`。
- 备份：`/ppt/ppt-agent/ppt-agent-linux.bak.202608290416-intent-router-benchmark`。
- 新进程：PID `174218`，工作目录 `/ppt/ppt-agent/backend`，命令 `../ppt-agent-linux --mode web`，监听 `:8080`。
- 健康检查：`/api/health`、`/health/ready`、`/api/templates/layouts` 均返回 `200`；启动日志确认 MySQL 已连接且无立即失败。
- 低成本路由冒烟：临时登录后，`/api/messages` 对明确新建请求返回 `create/prepare_create/pptagent`；`/api/tasks` 对已有页改稿请求返回 `409` 和 `fix`，未创建或渲染任务；临时会话随后注销（`200`）。
