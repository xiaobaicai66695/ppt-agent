# 003 生成编排与交付闭环决策

## 当前决策

首次生成流程以 LLM 直接路由和主 Agent 批量规划为准，`tasks.json` 是唯一页面协作契约；页面交付完成与否以代码维护的元数据和文件对账为准，而不是让主 Agent 反复读写文件或凭自然语言判断。

## 执行准则

- 主 Agent 只通过 `update_tasks_manifest` 初始化或批量 patch `tasks.json`，不得用脚本或普通文件编辑绕过 manifest tool。
- 规划阶段尽量一次补齐 `title`、`description`、`content_plan`、`background`、`layout_variant`，减少 SlideExecutor 同时承担搜索、抽取、规划和修复的压力。
- SlideExecutor 只负责单页生成，不修改任务完成状态；后端根据输出文件和交付元数据收敛 N/N。
- QA/Fix 默认关闭，除非显式配置启用；常规闭环依赖 generator 稳定性、manifest 校验和交付对账。
- 流式等待、慢模型和畸形 manifest 输入要有恢复策略，不能因为半包 JSON、超时或输入包裹成字符串导致整套任务卡死。
- 会话续写和终态 conversation 不能阻塞任务完成；历史会话恢复必须有锁边界和轻量返回。

## 取舍

这个决策牺牲了部分在线自修复能力，但显著降低 LLM 循环、重复读写 `tasks.json` 和高成本 QA 带来的时延与不确定性。质量问题优先在规划契约和 generator 容量上修，而不是依赖后置审查。

## 来源任务

- `PPT-FLOW-001`
- `PPT-FLOW-002`
- `PPT-RUNTIME-001`
- `PPT-RUNTIME-002`
- `PPT-RUNTIME-003`
- `PPT-STATUS-001`
- `PPT-UX-004`
- `PPT-OPS-002`

