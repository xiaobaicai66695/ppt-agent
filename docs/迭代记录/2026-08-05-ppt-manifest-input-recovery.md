# update_tasks_manifest 畸形输入恢复

**日期**：2026-08-05  
**方向**：生成编排与交付闭环  
**Route**：direct  
**状态**：in_progress

## 问题

线上任务偶发在首次规划阶段中断，错误为 `tasks string must contain a JSON array or task objects: invalid character '}' after array element`。失败发生在 `update_tasks_manifest` 入参解析阶段，原有 `tasks.json` 未被写入。

## 真实输入诊断

根据失败任务的 RuntimeEvent 完整 ToolCall 入参，单次调用同时出现：

- `tasks` 被序列化成 JSON 字符串，而不是 schema 定义的对象数组。
- 每个 task 对象后多出一个闭合花括号。
- 正文使用未转义的半角双引号，例如 `"一带一路"`。
- `content_plan.elements[].text/description` 偶发使用字符串数组。

已有恢复器只能处理单纯的额外分隔符；遇到裸引号或字段类型漂移后会放弃恢复，因此最终暴露原始数组解析错误。

## 实施方案

- 保持工具 schema 的结构化数组契约，并在主 Agent 规划说明中给出正向字段形态。
- 在 manifest 字符串兼容边界转义可判定的正文裸引号，再提取完整 task 对象。
- 对对象外字符采用白名单，只接受数组分隔符和多余闭合花括号。
- 任一 task 截断或无法解析时拒绝整批输入，继续保持 `tasks.json` 原子不变。
- 将 ContentElement 的 `text/description` 字符串数组规范化为换行文本，兼容模型常见类型漂移。

## 验证与上线

待完成本地聚焦测试、Linux 构建、服务器部署和 1-2 页线上冒烟后回填。
