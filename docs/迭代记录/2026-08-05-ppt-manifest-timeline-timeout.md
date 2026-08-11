# PPT Manifest Timeline 与生成超时修复

## 问题

任务 Timeline 在生成过程中持续出现大面积黄色 `manifest_validated · tasks.json` 事件。本次线上任务还在规划阶段因 `LLM 流式输出超时（3m0s）` 失败。

本次事项归入现有的 Agent Harness 与可观测性、生成编排与交付闭环、前端创作与交付工作台三个长期方向，登记为 `PPT-OBS-004`，按 direct 路线修复。

## 根因

- 后台交付监听器会周期性对账 `tasks.json`，但 `RecordManifestValidation()` 每次轮询都写入事件，没有判断 `done/total/pending/missing` 是否变化。
- 清单中仍有待生成页面属于正常运行态，却被记录为 `warning`；前端按 warning 给整行填充黄色背景，形成连续告警的错觉。
- 后端代码的默认流等待超时已经是 8 分钟，但服务器保留的 `.env` 仍设置为 3 分钟，覆盖了代码默认值。线上任务第二轮模型调用超过 3 分钟未返回后被主动取消。

## 实现

- `ppt-agent/backend/pkg/agent/utils/runtime_meta.go`
  - 缓存最近一次 manifest 校验快照，仅在完成数、总页数、待生成任务、缺失文件或状态变化时记录事件。
  - 待生成页面标记为 `running`；只有实际缺失文件才标记为 `warning`；全部交付后标记为 `ok`。
  - 事件摘要直接提供“已完成 X/Y 页，还有 N 页待生成”等可读信息。
- `ppt-agent/frontend/src/utils/workbench.ts`
  - 折叠历史任务中连续重复的 manifest 轮询事件，同时保留真实进度变化。
  - 将内部事件名转换为“交付进度核对”“PPT 页清单”“进行中”等用户可读文案。
- `ppt-agent/frontend/src/pages/DashboardPage.vue`
  - manifest 事件归入 Delivery 分类，不再归入黄色 Phase 分类。
  - Timeline 列表使用可读名称、状态和摘要，不再展示空泛的 `metadata`。
- 服务器 `STREAM_TIMEOUT` 从 3 分钟调整为 8 分钟。

## 验证

- `go test ./pkg/agent/utils ./pkg/task ./pkg/web ./pkg/agent/deep`
- `go test ./...`
- `go build ./...`
- Linux `amd64` 后端构建通过。
- `npm run test`：15 项通过。
- `npm run build`：通过。
- 模型 Chat Completions 最小请求：HTTP 200。

新增回归覆盖：

- 相同 manifest 快照重复轮询只记录一次。
- 0/N、1/N 等生成过程使用 `running`，N/N 使用 `ok`，缺失文件使用 `warning`。
- 前端折叠旧任务的连续重复事件，但保留不同进度摘要。
- Timeline 使用中文交付语义展示内部 manifest 事件。

## 上线

- 部署目录：`/ppt/ppt-agent`。
- 旧进程 PID：`459615`。
- 新进程 PID：`466458`。
- `http://124.220.22.162:8080/api/health`：HTTP 200。
- `http://124.220.22.162:8080/`：HTTP 200。
- 本地与远端后端二进制、前端入口文件 SHA-256 一致。
- 新进程日志确认 MySQL、技能和 Web 服务启动正常。
