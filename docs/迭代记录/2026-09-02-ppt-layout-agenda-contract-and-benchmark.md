# 2026-09-02 PPT 版式、目录副标题契约与 Planner Benchmark

## 目标

针对图文页正文偏小、章节分隔页保留旧蓝色编号侧栏、目录卡片只有章节名且显得空的问题，统一调整独立 `ppt-deck-planner` skill 的渲染和规划契约；同时把目录副标题纳入 `tasks.json` 与 Planner benchmark 的内容质量检查。

## 交付

- `image_text` 不再把正文截断后重复写到标题区；改用独立 `lede`，正文最小字号提高到 `15.5pt`，常规图文叙事使用 `16.5pt`，紧凑版使用 `18pt`。
- `section_divider` 统一为背景、主标题和副标题；旧 `number_sidebar` 只保留输入兼容别名，渲染器不再显示蓝色侧栏或大编号。
- `agenda` 使用 2×2 导航卡片承载四项目录；每个 `toc_item` 用 `title` 保存章节名、`body` 保存 18–44 字的非重复阅读引导。
- `component_contracts.json` 将上述 `agenda` 字段、长度和不可重复规则写成结构化规划约束；`SKILL.md`、生成器参考和 Planner prompt 同步说明。
- `pptbench` 的 `content_quality` 新增 `agenda_toc_items`、`agenda_toc_subtitles`、`agenda_subtitle_issues`。缺标题、缺副标题和副标题重复标题分别输出 `missing_toc_title`、`missing_toc_subtitle`、`repeated_toc_title`。
- 新增 test case `planner_agenda_navigation_001`，并新增独立 validation holdout `planner_validation_agenda_navigation_001`，均要求目录副标题及从问题到决策的叙事链。

## 验证

- `go test ./cmd/pptbench ./pkg/agent/deck` 通过。
- `python -m unittest discover -s skills/ppt-deck-planner/tests -v`：37 项通过；生成器逐文件 `py_compile` 通过。
- `validate_deck.py` 对 6 页目录/章节/图文回归 fixture 通过（仅 7 条现有素材 attribution 警告，无错误）。
- 真实 Planner benchmark：`planner_agenda_navigation_001` 运行 `model + judge`，总分 `5.00/5`；首稿质量报告为 `agenda_toc_items=4`、`agenda_toc_subtitles=4`、无 `agenda_subtitle_issues`。
- 本地渲染验收：目录为带副标题的双列导航，章节页无蓝色侧栏，图文正文放大且未重复副标题。

## 上线与冒烟

- 运行时 skill 已同步至 `remote-dev:/ppt/ppt-agent/skills/ppt-deck-planner`；共享路径 `/ppt/skills/ppt-deck-planner` 与项目路径内容一致。
- 本轮微调随 2026-09-02 功能发布一并部署。服务从 `/ppt/ppt-agent/backend` 以 `ppt-agent-linux -mode web -addr :8080` 重启，最终 PID `1688027`，监听 `:8080`。
- 线上内网与公网 `http://124.220.22.162:8080/api/health` 均返回 `200`；`/health/ready` 的 MySQL、LibreOffice、Python 组件均为 `ok`。
- 线上 skill fixture 预检通过，6 页 `render_deck.py` 冒烟成功；临时 smoke 目录和误启动 CLI 所产生的空任务输出已清理。完整发布证据见 `docs/迭代记录/2026-09-02-benchmark-feedback-release.md`。

## 遗留

- 本次 Planner prompt 的文字同步已在源码完成；生产运行时主要读取已上线 skill 契约。后续图文微调已纳入同一 Linux 发布包并完成线上复测。

## 后续图文阅读微调

- 图文正文面板内边距调整为 `0.50in × 0.46in`；短中篇正文在阅读区内轻微垂直居中，长文继续顶对齐以保留容量。
- 图文正文增加 `0.25pt` 字间距并将行距由 `0.98` 放松至 `1.06`；只作用于 `image_text` 正文，不影响标题、目录和普通内容页。
- 验证：组件回归 38 项通过、6 页本地 PPTX→PDF→PNG 目检无溢出；线上 skill 6 页渲染 smoke 成功。上线后 Web 进程 PID `1688027`，内网与公网 `/api/health` 均为 200，临时 smoke 文件已清理。
