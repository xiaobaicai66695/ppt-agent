# Model runtime

模型运行时代码按职责导航。当前文件保持在同一 Go package 以保留内部 helper 与压缩器依赖；子目录作为稳定的迁移索引。

- `config/`：选项、环境变量与 provider 解析
- `fallback/`：模型工厂、降级链、重试与限流
- `stream/`：流式工具调用清洗与 chunk 处理
- `observability/`：请求元数据、脱敏与 runtime 状态
