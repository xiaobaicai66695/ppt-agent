# 007 多模型与账户凭据边界决策

## 当前决策

业务 Agent 不直接绑定单一模型上游。模型创建、能力差异、超时、fallback 和并发约束由 provider compatibility 层处理；用户自备 Key 按账户和 provider 归属，不能跨 provider 复用或写入可见事件。

## 执行准则

- 业务代码通过稳定的模型选择接口请求能力，不在 Planner、Reviewer、Fixer 或 Web 层散落上游特例。
- fallback 只能在兼容同一任务需求的 profile 中发生，并记录不含凭据的 provider/model 诊断信息。
- Key 仅用于请求初始化和受保护存储；日志、SSE、管理接口、测试输出和文档不得暴露原文、DSN 或可恢复指纹。
- 新 provider 接入先补 profile、能力/超时差异和失败降级测试，再允许在账户设置中选择。

## 取舍

兼容层增加了配置和测试维度，但避免账户凭据串用、单一供应商故障扩大，以及业务流程被上游 SDK 差异绑死。

## 来源事项

`PPT-MODEL-001`、`PPT-PREFERENCE-001/002`、`PPT-AUTH-001`。
