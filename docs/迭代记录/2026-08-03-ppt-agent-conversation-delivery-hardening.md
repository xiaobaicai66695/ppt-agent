# 2026-08-03 PPT Agent 会话与交付加固

## 迭代目标

针对线上交付截图和实际使用反馈，集中修复完成页重叠、同页重复预览、任务后对话清空、Markdown 破坏、Agent 偏离不可见、长提示词撑开页面、图标缺失、Windows CLI 残留，以及生成入口和续聊入口割裂九项问题。

本次事项登记为 `PPT-UX-002`，按 large 需求进入 OpenSpec change：`openspec/changes/ppt-agent-conversation-delivery-hardening/`。默认 QA/Reviewer 仍保持关闭，不增加额外模型调用。

## 已完成

- 后端建立稳定输出文件身份，统一绝对路径、相对路径、Windows/POSIX 路径为 basename，并在任务完成、续聊、持久化和 API 输出前去重。
- continuation POST 改为立即返回 `202 accepted/queued` JSON；增量只走任务 SSE，并通过 `after_event_id` 避免续聊重放旧完成事件。
- assistant 回答按生命周期边界一次持久化，保留模型原始 Markdown 和换行，不再按标点拆消息；活跃、冷启动和旧任务均可恢复结构化会话。
- RuntimeMeta 增加用户意图锚点、冻结规划、当前执行页、契约对齐状态和结构化偏离告警，不调用额外模型。
- 新增唯一 `ConversationComposer`，统一创建、运行中排队和完成后续改；新任务可勾选预设模板或 ComposePage 返回的自定义大纲草稿。
- ComposePage 将版本化 outline 写入 `sessionStorage` 并返回 Dashboard，由用户在统一输入框中显式提交。
- 前端按逻辑页码和规范化文件名双重去重，长需求使用确定性 42 字摘要，原文通过折叠详情查看。
- 完成态主滚动区的一级业务区域全部禁止 flex shrink，预览、完成条、诊断和 Composer 依次参与文档流，不再互相覆盖。
- Markdown renderer 先转义 HTML，再支持标题、列表、段落、引用、代码块和表格；旧 `full_answer/conversation_content` 可恢复为可读消息。
- Visual Designer 增加 manifest 大小写、路径、图片可读性和部署态导入 smoke；未知图标显示实心主题色语义缩写，不再绘制空白方块。
- 项目 CLI 固定 `/bin/sh -c`，Python 默认 `/root/pptx_env/bin/python`，保留 `PYTHON_BIN` 覆盖；移除 Windows shell、Python 和 LibreOffice 健康检查分支。

## 关键决策

- 页面去重使用稳定身份而不是原始路径字符串；前后端同时兼容旧数据。
- 不新增 LLM 摘要成本，任务短标题由首个 Markdown 标题或首句确定性截断得到。
- 可观测性明确为结构契约对齐，不冒充语义评分；用户仍可同时查看原始需求、冻结规划和当前页。
- 会话只有一个输入入口，模板只影响新任务，不污染已有任务的续聊。
- 视觉资源必须完全位于 `ppt-agent/skills/visual_designer` 内，服务器无需工作区外部素材目录。

## 验证

- `go test ./pkg/agent/command ./pkg/tools/pythonutil ./pkg/agent/utils ./pkg/task ./pkg/session ./pkg/web` 通过。
- `go build ./...` 通过。
- `npm run test` 通过，4 个客户端契约测试全部成功；`npm run build` 通过。
- `python -m py_compile` 覆盖资源管理器、icon-grid 和验证脚本；`python scripts/validate_visual_assets.py` 通过。
- Playwright + Edge 覆盖 1440、768、375：16 页只渲染 16 张卡，横向溢出为 0，旧聊天框和 Sidebar 创建表单均不存在，完成条位于预览区之后，Composer 不覆盖最后一页，自定义草稿自动勾选。
- 浏览器滚动到底部展开诊断，确认意图、计划页轨道、当前执行、偏离期望/实际值和 Markdown 会话在桌面与手机均可读。
- `openspec validate ppt-agent-conversation-delivery-hardening --strict` 通过。

## 已知基线问题

`go test ./...` 的业务包均通过，但仓库原有 `test/intent` 离线规则评估仍因意图准确率 60% 低于 75% 基线而失败。本次未修改意图分类器；该结果不影响本次聚焦测试和后端构建。

## 关联内容

- `docs/issues/todo.md`
- `openspec/changes/ppt-agent-conversation-delivery-hardening/`
- `ppt-agent/backend/pkg/task/delivery.go`
- `ppt-agent/backend/pkg/agent/utils/runtime_meta.go`
- `ppt-agent/frontend/src/components/ConversationComposer.vue`
- `ppt-agent/frontend/src/utils/workbench.ts`
- `ppt-agent/scripts/validate_visual_assets.py`
