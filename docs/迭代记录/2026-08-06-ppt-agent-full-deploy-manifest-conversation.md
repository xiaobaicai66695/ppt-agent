# PPT Agent 全量打包部署与线上冒烟

## 背景

用户要求将当前 PPT Agent 工作区全部打包部署上线。本轮按交付闭环执行：本地验证、Linux 产物构建、服务器备份与部署、重启确认、线上低成本冒烟、测试数据清理和证据回填。

## 本轮修复

- 增强 `update_tasks_manifest` 的畸形 `tasks` 参数恢复能力：
  - 支持 `tasks` 为 JSON 字符串。
  - 支持多余闭合括号、顶层字段残片、整段参数被错误嵌套。
  - patch 模式缺少 `task_id` 时，支持按 `page_index` 或列表顺序映射到既有任务。
- 修复终态任务 `/conversation` 卡死：
  - `running` 任务继续走内存实时会话。
  - `completed/failed/cancelled` 任务即使还留在内存 map，也走持久化快照和轻量 runtime summary，避免被收尾 goroutine 的 `TaskState.Mu` 阻塞。

## 本地验证

- `go test ./pkg/agent/deep -run ManifestTool -count=1`
- `go test ./pkg/web -run Conversation -count=1`
- `go test ./...`
- `go build ./...`
- Linux 构建：`GOOS=linux GOARCH=amd64 go build -o ppt-agent-linux .`
- 前端单测：`npm run test`，17 项通过。
- 前端生产构建：`npm run build`
- Python 生成器编译：`python -m py_compile skills/visual_designer/generators/*.py`
- 单页模板 JSON 解析：24 个通过。
- OpenSpec：`openspec validate ppt-agent-generation-quality-and-observability --strict`

## 上线

- 部署目标：`remote-dev:/ppt/ppt-agent`
- 监听端口：`:8080`
- 最终进程：PID `730224`，命令 `./ppt-agent-linux -mode web -addr :8080`
- 最终二进制：`/ppt/ppt-agent/backend/ppt-agent-linux`，大小 `60742574`
- 服务器保留项：`backend/.env`、日志、`output`、`weboutput` 用户产物目录。
- 最终备份：
  - 后端：`backend/ppt-agent-linux.bak.20260806111428`
  - 前端：`frontend/dist.bak.20260806111428`

## 线上冒烟

- `GET /api/health`：200，`{"status":"ok"}`
- `GET /`：200，返回新前端构建 `index-DSBs7C6p.js`
- `GET /api/templates`：200，19 个 preset
- `GET /api/backgrounds`：200，6 个背景主题
- 创建 2 页低成本任务：`68e4a39d-afa1-43de-b891-d2e378f75ea0`
  - 最终状态：`completed`
  - 进度：2/2
  - 产物：`1_番茄钟方法.pptx`、`2_三步开始专注.pptx`
  - Runtime：`alignment_status=aligned`，`tool_calls` 包含 `update_tasks_manifest`、`task`、`python3`
- `GET /api/tasks/:id/conversation`：200，20397 bytes，0.003s
- 文件下载：
  - `1_番茄钟方法.pptx`：200，495139 bytes
  - `2_三步开始专注.pptx`：200，29293 bytes
- 缩略图：
  - 第 1 页：200，192954 bytes
  - 第 2 页：200，105365 bytes

## 清理

- 已通过 API 删除 smoke 任务：
  - `a49ef3ea-884c-4d37-ae8d-94af88516667`
  - `2c6a3b51-779d-4403-81fb-e7daadc2646e`
  - `68e4a39d-afa1-43de-b891-d2e378f75ea0`
- 已手动删除对应 smoke 输出目录：
  - `/ppt/ppt-agent/weboutput/1-a49ef3ea-884c-4d37-ae8d-94af88516667`
  - `/ppt/ppt-agent/weboutput/1-2c6a3b51-779d-4403-81fb-e7daadc2646e`
  - `/ppt/ppt-agent/weboutput/1-68e4a39d-afa1-43de-b891-d2e378f75ea0`
- 已删除本轮服务器 `/tmp` 部署包和 smoke 脚本。

## 遗留风险

- 日志分析后台偶发模型 400/429，不影响主生成链路；建议后续单独治理日志分析模型的输入裁剪和限流。
- 本地策略阻止删除未跟踪的 `ppt-agent/backend/ppt-agent-linux` 构建产物，需后续由人工或允许的清理流程移除。
