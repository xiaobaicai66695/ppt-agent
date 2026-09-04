# Web router / service / model 分层

## 事项

- ID：`PPT-RESILIENCE-003`
- 计划时间：2026-09-04
- 完成时间：2026-09-04
- 类型：direct，内部结构重构

## 产出

- 新增 `pkg/runtime/web/router`：集中注册认证、消息、任务、交付、凭据、模板和管理路由，并保持原 middleware 顺序。
- 新增 `pkg/runtime/web/model`：集中请求、响应、路由结果、健康检查和基准 DTO；凭据字段默认禁止 JSON 序列化。
- 新增 `pkg/runtime/web/service`：抽出任务创建/启动、大纲校验、重复请求抑制和模型凭据解析；保留 `web.NewServer` 兼容入口。
- 更新 Web/runtime 边界 README、架构索引和 struct 索引。

## 验证

```text
go test ./pkg/runtime/web/...
go test ./...
go build ./...
```

以上命令均通过；路由契约和凭据脱敏新增了边界回归测试。

本事项只调整内部 package 边界，没有执行服务器部署或线上冒烟；工作树中已有的日志分析下线、数据库锁和前端改动未纳入本事项。

代码提交：`547e62e`（router 默认未知路径为 404）及其父提交 `b209ef7`（主分层实现）。
