## Why

当前 benchmark 的分类和评分链路已具备基础，但各 suite 的样本数仅 2–8 条，覆盖面不足，难以发现回归。同时，线上 PPT 交付后缺少低摩擦的结构化反馈入口，无法把用户对最终成品的满意度沉淀为可分析信号。

## What Changes

- 将 `router`、`planner`、`reviewer`、`fixer` 的 test 与独立 validation 数据集分别补足为每类 10 条有效样本，并保持 validation 不被用于逐例调 prompt。
- 为 benchmark 增加样本数量与唯一 ID 的离线契约校验，阻止未来删减或重复样本悄然降低覆盖度。
- 新增任务级交付反馈：完成的 PPT 在 Dashboard 左下角显示 1–5 分评分，用户可选填建议并独立提交；不触发重新生成或修改工作流。
- 持久化反馈并提供任务读取接口，使刷新、重连和跨设备访问仍能恢复已提交反馈；同一用户对同一任务可更新自己的反馈。

## Capabilities

### New Capabilities

- `benchmark-suite-sample-coverage`: 定义四个 benchmark suite 在 test/validation 数据集中的最低样本量、唯一性和隔离要求。
- `ppt-delivery-feedback`: 定义已交付 PPT 的任务级评分、可选建议、持久化、权限与幂等更新行为。

### Modified Capabilities

- `ppt-agent-developer-status-ui`: Dashboard 新增已完成 PPT 的紧凑反馈入口，并在重载时恢复反馈状态。

## Impact

- Benchmark：`ppt-agent/benchmark/{cases,validation_cases}`、`cmd/pptbench` 数据集校验和操作文档。
- 后端：数据库反馈模型、任务反馈读写 API、TaskInfo/会话数据契约与授权校验。
- 前端：API 类型、Dashboard 左下角交付反馈组件及其提交/错误/已提交状态。
