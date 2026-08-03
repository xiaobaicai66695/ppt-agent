## Context

当前 Web 端由 Vue 3 单页应用和 Gin API 组成。任务生成已经支持按页输出、SSE、3 秒轮询和 RuntimeMeta，但任务资源路由只验证登录，没有验证所有权；部分处理器读取了未设置的 Gin key；缩略图在浏览器请求时同步启动 Python/LibreOffice；前端还会提前下载完整 PPTX；Dashboard 和 Compose 主要依赖固定桌面布局。默认在线 QA 已停用，但用户文案和状态仍有残留。

本变更必须保持 Go/Vue/Python 现有技术栈，不引入新的前端组件框架，不改变生成器内容契约，不依赖网络素材，并保留 RuntimeMeta 作为开发诊断能力。

## Goals / Non-Goals

**Goals:**

- 所有任务资源操作都验证当前用户归属，并统一 API 身份与错误语义。
- SSE 在断线重连后增量回放，终态可由轮询可靠恢复，系统步骤能够被用户界面消费。
- 单页 PPTX 落盘后后台准备缩略图，浏览器先展示稳定占位，再在预览就绪时刷新图片。
- 移除无效 PPTX 预缓存，原文件只在明确下载时传输。
- 使用本地生成器产出的真实图片作为预设模板缩略图。
- Dashboard 和 Compose 在桌面、平板和手机视口均有可操作布局。
- 用户阶段进度优先，RuntimeMeta 与 Timeline 收纳为可折叠诊断视图。

**Non-Goals:**

- 不恢复默认 Reviewer/QA 模型调用。
- 不更换 Vue、Gin、SSE 或 LibreOffice 转换方案。
- 不实现浏览器内 PPTX 编辑或完整幻灯片播放引擎。
- 不引入外部图片服务、CDN 或新数据库迁移。

## Decisions

### Decision 1: Route-level task ownership middleware

在认证中间件之后为所有 `/:id` 任务资源路由增加统一所有权中间件。中间件从 Request Context 获取用户，加载任务摘要，并仅允许任务所有者或管理员继续；不存在或无权访问统一返回 404，避免任务 ID 枚举。处理器继续使用现有 TaskManager，不复制授权判断。

选择统一中间件而不是在每个 handler 内散落比较逻辑，因为任务流、文件、缩略图、取消、删除、继续和会话必须保持相同规则，也更容易编写路由级测试。

### Decision 2: Normalize API response handling and profile identity

后端处理器统一使用 `userIDGin`，不再读取未设置的 `c.Get("userID")`。前端所有变更型和数据型请求通过共享 `checkResponse` 处理非 2xx。偏好编辑器使用字符串草稿，保存时拆分为数组，避免把输入框字符串发送给 `[]string` 字段。

### Decision 3: Incremental SSE replay with event IDs

`SSERichEvent` 增加单调递增 ID，服务端输出 SSE `id:` 行，并根据 `Last-Event-ID` 只回放缺失事件。前端保留原生 EventSource 自动重连，不在网络错误时立即关闭连接；轮询继续作为状态终态兜底。`system_step`、`thumbnail_ready` 和 `thumbnail_error` 成为显式前端事件类型。

该方案复用 SSE 标准能力，比自定义重连查询参数简单，并能减少重复日志。事件环仍保留上限，过旧 ID 只能从当前环首继续回放。

### Decision 4: Background thumbnail preparation

TaskManager 在发现新的 PPTX 文件后调用可选的文件就绪回调。Web Server 使用该回调在后台调用现有缩略图转换器，并向 TaskState 广播 `thumbnail_ready` 或 `thumbnail_error`。缩略图转换继续按 workDir 串行，但 `GenerateThumbnail` 在获得锁后再次检查磁盘缓存，避免浏览器请求与后台任务重复转换。

前端收到 `file_ready` 后立即显示 16:9 骨架；收到 `thumbnail_ready` 后刷新图片 URL。完整 PPTX 不再写入 Cache API。相比引入常驻 UNO 服务，这一方案改动较小，也能显著降低首次可见预览的等待和重复转换。

### Decision 5: Real local preset thumbnails

新增可重复运行的本地维护脚本，使用现有标题页生成器和 LibreOffice/转换工具为每个 full-deck preset 生成一张真实 JPEG/PNG，输出到 `frontend/public/templates/thumbs` 并提交生成结果。模板接口继续返回现有 URL，不新增网络依赖。

### Decision 6: Responsive workbench, not a second mobile product

Dashboard 在窄屏把固定 Sidebar 变为抽屉，双栏任务区变为单栏，聊天条占满可用宽度；Compose 在中等视口使用双栏加编辑抽屉，在手机视口改为纵向模板库与画布，编辑器作为全屏层。现有组件和信息架构保持不变，只重排，不新建另一套页面。

所有点击型 `div` 在本次触及范围内改为按钮或补键盘语义，主要图标按钮至少提供可访问名称，触控目标尽量达到 44px。

### Decision 7: User progress and developer diagnostics are separate layers

主进度组件在总页数未知时也显示当前阶段和不确定进度；计时器使用响应式 interval；轮询结果同步回任务列表状态。RuntimeMeta/Timeline 放入默认折叠的 `details` 区域，警告存在时自动保持可见提示，但不抢占预览主区域。

## Risks / Trade-offs

- [Risk] 后台缩略图仍依赖 LibreOffice，机器未安装时会失败。 -> 保留明确的转换失败状态、重试按钮和 PPTX 下载能力。
- [Risk] SSE 事件环截断后无法从任意旧 ID 完整恢复。 -> 轮询恢复任务摘要和文件列表，历史文本从 conversation 接口恢复。
- [Risk] 响应式 CSS 改动可能影响桌面密度。 -> 使用 820px/1100px 分层断点，并保留大屏现有网格比例。
- [Risk] 提交真实缩略图会增加仓库体积。 -> 每个 preset 仅保存一张压缩的 16:9 WebP/JPEG，并控制分辨率。
- [Risk] 所有权中间件可能暴露旧数据中 `UserID=0` 的任务。 -> 仅管理员可访问无明确所有者的历史任务，普通用户返回 404。

## Migration Plan

1. 先部署任务授权、身份读取和 API 错误处理，验证现有登录任务链路。
2. 部署 SSE ID 与轮询终态同步；旧前端可忽略新增事件字段。
3. 启用后台缩略图回调和前端预览状态；失败时仍保留原按需端点。
4. 发布真实模板缩略图和响应式 UI。
5. 回滚时可分别撤销缩略图回调和响应式样式；任务数据与数据库结构无需回滚。

## Open Questions

- 常驻 LibreOffice/UNO 转换服务可进一步降低延迟，但需要独立生命周期管理，本变更暂不引入，待真实耗时数据决定。
