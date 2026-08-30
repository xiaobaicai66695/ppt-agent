# 2026-08-31 Planner/Reviewer 正文密度与组件字号修复

## 目标

- 将“页面内容太空”从主观观感转成 Planner、Reviewer 和 benchmark 共用的可验证约束。
- 修复 `image_text` 与 `card_grid` 的正文过小问题，避免在有大面积文字面板或四宫格卡片时显得空洞。

## 实现

- 确定性审查把以下 `low_information_density` 问题从 warning 升为 error，因而会阻断 Planner 首稿并进入 Reviewer 授权修复范围：
  - `image_text` 主正文少于 240 个中文字符；
  - `content_slide`、`card_grid`、`two_column`、`three_column`、`case_study`、`example_detail`、`deep_dive`、`summary_slide`、`comparison_table`、`swot_analysis` 的正文组件合计少于 220 个中文字符。
- Reviewer prompt 明确字数统计口径：仅计算组件 `body`、`text` 与 `items`，标题、图片 caption 与来源行不计入；完整论述型 `argument_block` 仍要求 440–840 字。
- Planner 的两条 benchmark case 添加 `minimum_content_characters` 预期；Reviewer 新增 `reviewer_repair_low_information_density_001`，只允许补齐第 2 页 `card_grid`，用于评价字数达标、最小修改和未授权页不变性。
- `image_text` 正文面板启用更强调的正文层级：短正文为 15pt，其余正文为 13.8–14pt；`card_grid` 正文为 13.2pt、标题为 17pt。其他页面维持原字号规则。

## 本地验证

- `go test ./pkg/agent/deck ./pkg/prompts ./cmd/pptbench` 通过。
- `go build ./...` 通过。
- `npm run build` 通过。
- bundled Python：全部 generator 编译通过，`python -m unittest discover -s skills/ppt-deck-planner/tests` 为 33/33 通过。
- 生成两页验收 PPT 并经 LibreOffice 转 PDF/PNG 检查：`image_text` 正文 15pt、`card_grid` 正文 13.2pt，未出现正文溢出或原有小字大空框。

## Benchmark 发布门禁

- `test` 只用于修复迭代；`validation` 是独立 holdout，只有验证集全量通过且无 `critical_failure` 才允许上线。该规则已写入 `ppt-agent/benchmark/fix-guide.md`。
- Planner test：`benchmark/runs/20260831-012500-test-planner-density-final`，2/2 通过，平均 5.00。
- Planner validation：`benchmark/runs/20260831-013300-validation-planner`，2/2 通过，平均 5.00。
- Reviewer 在 validation 首轮发现“显式 case issue 覆盖确定性质量门”的 benchmark/生产输入不一致；修复为合并两类 issue，并在 Reviewer 修正输入中强制同一次 patch 填写 `content_plan.slide_intent` 与正文密度修复。
- Reviewer test：`benchmark/runs/20260831-014200-test-reviewer-final`，4/4 通过，平均 5.00。
- Reviewer validation：`benchmark/runs/20260831-014300-validation-reviewer-release`，1/1 通过，平均 5.00，`critical_failures` 为空。

## 部署与线上冒烟

- 目标：`remote-dev:/ppt/ppt-agent`。后端 Linux 二进制、`frontend/dist` 和 `skills/ppt-deck-planner` 已替换；本次旧版本备份保留在 `/ppt/ppt-agent/.release-backups/20260831-density-release-014244`。
- 2026-08-31 启动新进程 PID `809752`，cwd `/ppt/ppt-agent/backend`，监听 `:8080`；日志确认 `deck_planner_skill_ready` 与 `mysql_connected`。
- `GET /api/health`、`/`、`/dashboard`、`/api/templates/layouts` 均返回 200，公网 `http://124.220.22.162:8080/` 同样通过。
- `/api/templates` 返回 404 是当前路由契约（旧接口已移除），不是本次发布故障；模板布局接口为 `/api/templates/layouts`。
