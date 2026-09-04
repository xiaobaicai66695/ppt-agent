# Model runtime

模型运行时代码按职责导航。所有实现文件都在当前目录并属于同一 Go package，以保留内部 helper 与压缩器依赖。

- `model.go`、`utils.go`：选项、环境变量与 provider 解析
- `model.go`、`model_runtime.go`：模型工厂、降级链、重试与限流
- `model.go`：流式工具调用清洗与 chunk 处理
- `runtime_meta.go`、`token_tracker.go`：请求元数据、脱敏与 runtime 状态
