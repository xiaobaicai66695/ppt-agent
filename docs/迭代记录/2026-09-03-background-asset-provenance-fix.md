# 2026-09-03 背景素材溯源元数据修复

- **事项**：`PPT-ASSET-003`
- **路线**：direct
- **状态**：实现与本地验证完成，等待部署

## 修复内容

- 后端 `VisualIntent` 与 `PlanComponent` 正式保留 `provider`、`search_status`，避免 JSON 读写后丢失图片搜索溯源字段。
- Unsplash 背景和前景素材下载成功后写回 `provider="unsplash"`、`search_status="resolved"`；已有本地背景只有具备可信 provider 和成功状态时才会复用，否则由既有素材服务重新解析。
- 新增回归断言，覆盖背景/前景素材物化与组件 JSON 反序列化后的溯源字段。

## 本地验证

- `go test ./...`、`go build ./...`（`ppt-agent/backend`）通过。
- `npm run build`（`ppt-agent/frontend`）通过。
- `python -m unittest discover -s skills/ppt-deck-planner/tests -v`：31 项通过。
- Python generator 编译、独立 DeckSpec 预检与 1 页 PPTX 实际渲染通过；临时验证文件已清理。

## 部署状态

- 用户于 2026-09-03 明确要求当前执行结束后**暂不上线**；未复制文件、未重启服务、未创建线上冒烟任务。
- 后续部署需构建 Linux 交付物，确认 `remote-dev:/ppt/ppt-agent` 现有启动方式后，检查 `/api/health` 并执行 1–2 页低成本生成链路冒烟。
