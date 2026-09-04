# 003 生成编排与交付闭环决策

## 当前决策

首次生成以受控意图路由、PPTPlanner 完整草稿、TaskPlanReviewer 质量门、Go 原子提交和并发渲染为准。`tasks.json` 是跨 Agent、生成器、TaskManager 与前端的唯一正式页面契约；交付终态由代码维护的元数据和文件对账决定，不能由模型自然语言宣称完成。

Planner 必须尽力填写完整的 DeckSpec 内容字段，包括页面结构、稳定英文 `content_type`、`layout_variant`、`content_plan.components`、视觉意图与来源；运行态、系统派生字段和文件路径由 Go 管理。草稿只写入 `tasks.draft.json`，Reviewer 根据确定性问题报告定向修补，最多三轮后才允许原子提交正式 `tasks.json`。

## 执行准则

- Planner 只通过受控 manifest 工具初始化草稿；Reviewer 只可批量修补草稿；任何 Agent 都不得以普通文件编辑绕过校验或提交边界。
- 规划阶段一次补齐可确定的内容字段，避免将资料检索、语义补全、容量判断或版式猜测下放给 Renderer/Skill 脚本。
- benchmark 和开发调试评价的是完整草稿质量，Reviewer 不是常规内容补全器。
- SlideExecutor 只负责单页生成，不修改任务完成状态；后端根据输出文件和交付元数据收敛 N/N。
- 不做自动的整套 QA/Fix 循环；PPTFixer 只响应用户对已交付页面的定点修改，且由 Go 限制可改页面与字段。
- 流式等待、慢模型、畸形 manifest 与暂时性模型错误必须有恢复策略，不能使整套任务永久卡死。
- 会话续写和终态 conversation 不能阻塞任务完成；历史恢复必须有锁边界、事件游标和轻量返回。

## 取舍

这个决策限制了模型自由修改文件和全自动返工，但显著降低 LLM 循环、重复读写 `tasks.json` 和高成本后置修复带来的不确定性。质量优先在规划契约和 generator 容量上解决。

## 来源任务

- `PPT-FLOW-001`
- `PPT-FLOW-002`
- `PPT-RUNTIME-001`
- `PPT-RUNTIME-002`
- `PPT-RUNTIME-003`
- `PPT-STATUS-001`
- `PPT-UX-004`
- `PPT-OPS-002`
