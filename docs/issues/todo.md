# 需求登记表

| ID | 事项 | Size | Route | Status | 产出/链接 |
| --- | --- | --- | --- | --- | --- |
| PPT-QUALITY-001 | 按顺序修复 PPT Agent 生成质量与前端展示的七个问题：前端 CSS token、自定义编排内容补齐、outline/prompt 契约、生成器文本适配、缩略图状态、输出状态可靠性、仓库产物治理 | large | direct | done | 代码：`ppt-agent/frontend/src/App.vue`、`ppt-agent/frontend/src/pages/ComposePage.vue`、`ppt-agent/frontend/src/components/SlidePreviewCard.vue`、`ppt-agent/backend/pkg/web/handler.go`、`ppt-agent/backend/pkg/agent/deep/types.go`、`ppt-agent/backend/pkg/agent/deep/run.go`、`ppt-agent/backend/pkg/task/manager.go`、`ppt-agent/backend/pkg/agent/utils/model.go`、`ppt-agent/backend/pkg/prompts/deep/*.tmpl`、`ppt-agent/skills/visual_designer/generators/base.py`、`.gitignore`；验证：`npm run build`、`go test ./pkg/web ./pkg/task ./pkg/agent/deep ./pkg/agent/utils`、`go build ./...`、`go vet ./pkg/web ./pkg/task ./pkg/agent/deep ./pkg/agent/utils`、`python -m py_compile skills/visual_designer/generators/base.py` |
