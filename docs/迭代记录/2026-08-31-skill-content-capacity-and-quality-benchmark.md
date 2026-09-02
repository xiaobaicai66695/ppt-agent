# 2026-08-31 Skill 内容容量合同与内容质量 Benchmark

## 目标

- 将 Planner 首稿的正文容量规则从项目专属 prompt 迁入独立 `ppt-deck-planner` skill，使通用 Agent 与项目内 Planner 读取同一份合同。
- 把“主题是否聚焦、页面是否推进论证”纳入真实 Planner benchmark，避免只以合法 JSON、字数和页面类型判断质量。

## 实现

- `skills/ppt-deck-planner/templates/component_contracts.json` 新增 `planning_rules.page_content_density`，集中声明信息页、图文页、完整论述、列表项与双栏对照的容量规则；`SKILL.md` 和生成器参考将其标为唯一数字来源。
- `plan_review_tool.go` 运行时读取 skill 合同，不再保存平行的内容密度数字常量；缺少合同本身会阻断审查。
- Planner 首稿工具只允许一次初稿和一次质量门重提；第三次 `initialize` 被拒绝，防止模型在失败反馈中循环。短于完整论述门槛的正文保留为 `paragraph`，不再错误标为 `argument_block`；缺失的组件 ID 由服务端稳定补齐。
- Planner prompt 增加“用户要求/素材事实 → 页码 → title/content_type → 组件”的内部可追溯检查，明确页面角色、内容类型、数据事实与决策请求不得仅被隐含提及。
- `cmd/pptbench` 新增 `content_quality` 证据：deck 主张、逐页主张、缺失主张页、完全重复主张和连续同类叙事布局；Planner Judge rubric 增加必填维度“主题聚焦与跨页叙事连贯性”。两个 test Planner case 增加中心结论、论证链和连续布局期望。

## 验证

- `go test ./cmd/pptbench ./pkg/prompts ./pkg/agent/deck ./pkg/web`：通过。
- `go test ./test/plan_benchmark -run TestGoldDeckSpecsPassReviewer -count 1`：5/5 通过。
- 真实 test Planner benchmark：业务 KPI 与技术架构 case 均为 5/5；结果分别保存在 `ppt-agent/benchmark/runs/20260831-115400-test-planner-content-quality-evidence/` 与 `ppt-agent/benchmark/runs/20260831-141000-test-planner-content-quality-technical-semantic-component/`。
- 独立 validation Planner benchmark：留存专项、城市公交均为 5/5，平均 5.00；结果在 `ppt-agent/benchmark/runs/20260831-142000-validation-planner-content-quality-id-normalization/`。

## 部署与冒烟

- 2026-08-31 14:12（Asia/Shanghai）部署到 `remote-dev:/ppt/ppt-agent`；新进程 PID `985264`，cwd `/ppt/ppt-agent/backend`，监听 `:8080`。
- 二进制 SHA-256：`ad41e1969a89bd14a831d7495be5a607a978eafb6a360a60fa9552a9663b1967`；前序二进制与两个 skill 根目录文件均按时间戳备份。
- `/api/health`、`/health/ready`、`/api/templates/layouts`、`/`、`/dashboard` 的内网/公网检查均为 HTTP 200；ready 检查确认 MySQL、Python、LibreOffice 正常。
- 已同步 `/ppt/ppt-agent/skills/ppt-deck-planner` 与 `/ppt/skills/ppt-deck-planner`；两个 `component_contracts.json` SHA-256 一致。远端上传暂存目录已删除。

## 遗留风险

- 完整线上生成冒烟未执行：`/api/tasks` 必须登录，未携带会话 token 返回 401；本轮不读取、伪造或绕过用户认证，故不能安全创建后再删除临时任务。实现已发布且基础服务冒烟通过，但需在有专用测试账号/授权会话时补做 1–2 页生成、文件/缩略图核对和任务清理，才可声明生成链路的完整线上闭环。
