# 2026-08-27 PPT Agent 流式文本与视觉色板修复

## 背景

- 用户反馈主会话中英文流式文本单词被拼接，且模型可见输出应默认使用中文。
- 用户反馈 `image_text` 页面正文过短却占用大文本框，版式长期固定为左图右文，页面显得空洞。
- 背景图片上线后和固定主题色容易冲突，尤其是政府红等强色系会压过背景图和正文可读性。

## 行为变化

- 流式 assistant 内容合并时保留 ASCII 单词边界，兼容上游返回“增量 chunk”和“累计快照”两种形式。
- Planner、Reviewer、Fixer 的用户可见自然语言和页面正文默认要求中文，专有名词、英文缩写、URL 和代码字段除外。
- `image_text` 从三种变体扩展为四种：`image_left`、`image_right`、`image_top_band`、`image_bottom_band`；未显式指定时按页序自动轮换。
- `image_text` 规划质量门增加正文密度检查，少于 240 字时给出 `low_information_density` 警告，避免大面积空框。
- 背景图渲染改为确定性处理：降饱和、降对比、提高白色柔化层，并从背景位图提取弱化色系 token 覆盖固定主题色。
- 意图路由和用户画像不再向 Planner 注入历史主题色/固定推荐主题；顶层 theme 只作为无背景页和旧接口兜底。

## 本地验证

- `D:\anaconda\python.exe -m unittest tests.test_render_task_components`
- `python -m json.tool templates\component_contracts.json`
- `go test ./pkg/web ./pkg/task ./pkg/agent/deck ./pkg/prompts`
- `go build ./...`
- `npm test -- --run src/utils/workbench.test.ts`
- `npm run build`

## 上线记录

- 目标：`remote-dev:/ppt/ppt-agent`
- 时间：2026-08-27 00:22 Asia/Shanghai
- 新进程：PID `3603484`，命令 `../ppt-agent-linux -mode web -addr :8080`，cwd `/ppt/ppt-agent/backend`
- 备份：
  - 二进制：`/ppt/ppt-agent/ppt-agent-linux.bak.20260827002200-visualpalette`
  - 前端：`/ppt/ppt-agent/frontend/dist.bak.20260827002200-visualpalette`
- 启动确认：
  - `/api/health` 返回 HTTP 200，`{"status":"ok"}`
  - `/` 返回 HTTP 200，709 bytes
  - `/api/templates/layouts` 返回 HTTP 200，17522 bytes，包含 `image_bottom_band`
  - `/api/themes` 返回 HTTP 200，2614 bytes
  - `:8080` 正常监听，进程为 `3603484`
- 线上生成器 smoke：
  - 在 `/tmp/ppt-agent-visualpalette-smoke` 创建 4 个临时 `image_text` 任务，不调用模型。
  - 4 个任务均由 `/ppt/ppt-agent/skills/ppt-deck-planner/generators/render_task.py` 成功渲染，输出 `slide_1.pptx` 至 `slide_4.pptx`，单文件约 32K。
  - 临时 workdir、接口响应文件和传输包已清理。

## 遗留

- 本次没有发起完整 LLM 生成任务，避免额外消耗用户或系统上游 Key；已通过确定性渲染 smoke 覆盖生成器链路。
- 若后续仍出现背景局部复杂导致文字难读，可继续增加“文字安全区亮度/复杂度”检测，而不是依赖 Planner 视觉描述。
