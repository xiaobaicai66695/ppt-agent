# 2026-08-03 PPT 模板入口与智能推荐加固

## 迭代目标

修复模板填充时 `content_plan.elements[].items` 返回对象而导致 Go JSON 解析失败的问题，并把新任务的模板决策收敛到提示词首页。主 Agent 在一次正常运行中同时处理空白模板、示例文字模板和用户已填写模板，不再增加独立的模型填充阶段。

本次事项登记为 `PPT-UX-003`，按 large 需求进入 OpenSpec change：`openspec/changes/ppt-template-entry-recommendation-hardening/`。

## 已完成

- 新增 content plan 宽容解码边界，兼容字符串、数字、布尔值、对象和混合数组，归一化后继续向下游暴露稳定 `[]string`。
- 任务创建只校验和归一化 outline，不再调用模型补齐模板；旧 `/api/ai/generate-outline` 仅保留兼容用途。
- outline 增加 `template_scaffold` 与 `user_outline` 两种内容模式；主 Agent 重写预设示例文字，同时保留自定义 outline 的非空用户字段并补齐空字段。
- 任务创建 API 增加向后兼容的 `template_selection`，支持 `recommended` 和 `preset`，预设结构由后端实时 loader 解析。
- 智能推荐使用确定性元数据评分选择真实存在的模板、主题和背景策略，不增加 LLM 调用；背景只应用到封面、章节、引用、总结等视觉页。
- 首页动态展示智能推荐、19 套在线预设和自定义编排，预设或推荐提交后直接创建任务并进入 Dashboard。
- ComposePage 移除整套预设选择、批量填充、测试大纲、草稿中转和单页 AI 续写，默认提供三页可编辑空结构并直接创建 `user_outline` 任务。
- Dashboard 与 `ConversationComposer` 移除模板复选框、模板下拉和草稿状态，会话框只负责自由新建、排队反馈和续聊。
- Vite 开发代理支持 `VITE_API_PROXY_TARGET`，便于在本地使用远端 API 做真实目录验证，生产仍使用同源 `/api`。
- 部署时补充 MySQL DSN 默认连接/读写超时，并在迁移前拒绝系统 schema，避免数据库网络故障无限阻塞服务启动或误迁移 `information_schema`。

## 验证

- 聚焦后端测试通过：`pkg/agent/contentplan`、`pkg/agent/deck`、`pkg/generic`、`pkg/web`、`pkg/task`、`pkg/agent/utils`、`pkg/db`。
- `go build ./...` 通过；Linux `amd64`、`CGO_ENABLED=0` 二进制构建通过。
- `npm run test` 通过，2 个测试文件共 8 项；`npm run build` 通过。
- `openspec validate ppt-template-entry-recommendation-hardening --strict` 通过。
- Playwright + Edge 覆盖 375、768、1440：Home、Compose 和 Dashboard 均无页面级横向溢出；首页共展示 21 个选择项；Dashboard 中模板控件数量为 0。
- 请求拦截验证智能推荐提交 `template_selection.mode=recommended`；自定义编排提交 `content_mode=user_outline`、3 页结构，并保留空标题交给主 Agent。
- 远端覆盖部署到 `/ppt/ppt-agent`，Linux 服务 PID `4013875` 监听 `:8080`；`/api/health`、`/api/templates` 和 `/` 返回 200，远端前端入口文件与本地 SHA-256 一致。

## 已知环境问题

- `go test ./...` 的业务包均通过，但仓库原有 `test/intent` 纯规则评估准确率仍为 60%，低于 75% 基线；本次未修改意图分类器。
- 远端 MySQL 主机存在间歇性读取超时，当前服务会在有限等待后降级启动，因此公开页面和模板接口可用，但默认管理员认证探测返回 401。日志表现为网络超时而非账号拒绝；本次未改写用户提供的数据库凭据。需要在服务器网络或 MySQL 白名单侧继续排查。

## 关联内容

- `docs/issues/todo.md`
- `openspec/changes/ppt-template-entry-recommendation-hardening/`
- `ppt-agent/backend/pkg/agent/contentplan/`
- `ppt-agent/backend/pkg/web/template_recommendation.go`
- `ppt-agent/backend/pkg/agent/deck/agent.go`
- `ppt-agent/frontend/src/pages/HomePage.vue`
- `ppt-agent/frontend/src/pages/ComposePage.vue`
- `ppt-agent/frontend/src/components/ConversationComposer.vue`
