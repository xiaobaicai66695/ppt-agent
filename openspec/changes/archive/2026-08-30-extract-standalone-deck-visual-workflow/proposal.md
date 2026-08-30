## Why

`ppt-deck-planner` 已具备独立渲染与 Unsplash 下载脚本，但其“图片可选”文案与后端 Planner 的默认背景策略相冲突。通用 Agent 因而可能既不规划图片也不下载前景图片，且渲染入口不会阻断未物化的视觉资产。

## What Changes

- 将通用的视觉策略、视觉意图字段、素材物化顺序与渲染前置校验收敛为 `ppt-deck-planner` 的唯一契约。
- 使独立 Unsplash 素材解析器同时处理页级视觉意图和 `image` 组件，并保留可追溯元数据。
- 让独立预检/渲染入口在声明图片却未物化时失败；纯文字 deck 必须显式选择无图策略。
- 收窄后端 Planner 指令：Planner 只产出图片语义，后端渲染工作流确定性下载素材，避免重复工具下载。
- 为 Planner 首稿质量门补充回归断言，并运行现有 benchmark 的低成本和真实模型路径（配置可用时）。

## Capabilities

### New Capabilities

- `standalone-deck-visual-workflow`: 定义独立 DeckSpec 的视觉策略、素材物化和渲染前校验闭环。
- `planner-first-draft-visual-quality`: 定义 Planner 首稿的视觉意图质量门和 benchmark 回归证据。

### Modified Capabilities

- 无。

## Impact

- `ppt-agent/skills/ppt-deck-planner` 的 skill、组件契约、独立脚本、预检、渲染入口和测试。
- `ppt-agent/backend/pkg/prompts/planner`、Planner Prompt 测试及可选的图片工具注册策略。
- `ppt-agent/backend/test/plan_benchmark` 及其运行记录；真实模型 benchmark 需要现有模型凭据。
