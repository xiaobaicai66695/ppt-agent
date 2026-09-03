# 2026-09-03 闲聊 Agent 真实模型 benchmark 与 DeepSeek 发布

- **事项**：`PPT-CHAT-003`
- **路线**：OpenSpec `add-chat-agent-benchmark`
- **状态**：已上线

## 交付内容

- chat benchmark 的 test / validation 各 10 条，覆盖检索整合、无图片能力降级、来源链接清洗、图片归属、直接回答与多轮上下文。
- 实际模型 benchmark 通过生产 `deepseek/deepseek-chat` 调用每条用例；模型错误会写入结果并让 benchmark 失败，不再将 fallback 误报为模型通过。
- 闲聊 prompt 保留最近会话上下文，修复“厦门三天亲子行 → 再补充交通建议”等无检索 follow-up 丢失主题的问题。
- 默认与辅助文本模型统一走 provider-aware `MODEL_TEXT_*` 配置，线上默认 provider 为 DeepSeek。

## 验证

- 真实模型 chat benchmark：20/20 通过；模型错误、空回答和通用 fallback 均为 0。
- `go test ./...`、`go build ./...`（`ppt-agent/backend`）通过。
- `npm run build`、`npm test -- --run`（`ppt-agent/frontend`）通过。
- Python generator 编译和 28 项组件测试通过。

## 发布与冒烟

- 部署目标：`remote-dev:/ppt/ppt-agent`；最后新进程 PID `2037703`，工作目录 `/ppt/ppt-agent/backend`，监听 `:8080`。
- 备份：`deploy-backups/20260903-1749-deepseek-chat`、`20260903-1756-text-provider`、`20260903-1804-chat-context`。
- `/api/health`、`/health/ready` 均正常；访客隔离会话的上下文聊天冒烟通过，测试会话已删除。
- 线上日志确认主回复和路由辅助模型均初始化为 `deepseek/deepseek-chat(backup-0)`。
