## Context

当前独立 skill 将图片描述为可选，Unsplash CLI 只解析页级 `visual_intent`，`render_deck.py` 不会调用视觉资产校验。后端则要求默认背景并在渲染前物化资产，Planner Prompt 又要求工具可用时主动下载，形成重复且不一致的职责边界。

## Goals / Non-Goals

**Goals:**

- 将平台无关的视觉规划、下载顺序与渲染前置条件放入通用 skill。
- 保持 Planner 只产出语义计划，使用确定性素材解析代替 LLM 自主下载。
- 让独立路径和后端路径都能在素材缺失时明确失败或显式无图。
- 用静态 prompt 测试与现有 benchmark 证明首稿质量门未回退。

**Non-Goals:**

- 不引入新的图片 provider 或在 generator 内发起网络请求。
- 不改变图片 provider 的认证保存方式，也不将凭据写入 DeckSpec。
- 不要求纯文字、数据密集或用户明确要求无图的 deck 下载图片。

## Decisions

### Core skill owns visual policy

skill 声明默认 `visual_policy.mode="required"`，纯文字任务必须显式使用 `mode="none"` 或 `clean_text_only`。Planner Prompt 只引用该规则，不再复制每页背景、检索词和下载细节。这样独立 Agent、后端 Planner 和后续调用者共享同一判断标准。

### Plan and materialize are separate phases

Planner 必须写可解析的背景 `visual_intent` 和前景 `image` 组件查询。独立 CLI 在规划完成后扫描两类声明并原子回写本地路径、来源和署名；后端继续由 `MaterializePlannedDeckAssets` 在渲染前完成同一职责。相比让 Planner 使用 `search_images(download=true)`，该方式避免重复下载、迭代耗尽和模型遗漏工具调用。

### Validation is the enforcement point

`validate_deck.py` 和 `render_deck.py` 都调用视觉资产校验。只有声明了图像但尚未物化、或 `required` 策略不满足最低图片页数时阻断；无图模式保持可渲染。渲染入口再做一次校验，避免调用方绕过显式预检。

### Benchmark verifies contracts before expensive model runs

静态测试覆盖 Prompt 不再指挥下载、skill 包含完整流程，以及视觉预检的拒绝行为。先运行离线 gold DeckSpec benchmark；如果环境已有模型配置，再运行一条真实 Planner benchmark 并记录首稿确定性审查结果，缺少凭据时如实标记而非伪造结果。

## Risks / Trade-offs

- [旧的 query-only 示例在渲染时失败] → 将示例明确设为 `none`，并在迁移文档提供下载步骤。
- [独立 CLI 重写 manifest 时遗漏组件字段] → 将视觉项收集抽为共享逻辑，并补充前景组件下载测试。
- [默认 required 增加图片 API 成本] → 保留显式无图策略；后端按 content type 去重背景，独立 CLI 复用已解析路径。
- [真实模型 benchmark 不可运行] → 离线 gold benchmark 和 Prompt 契约测试仍是必跑门，线上模型 benchmark 在可用凭据下执行。

## Migration Plan

1. 更新 skill、契约和示例，明确默认/豁免策略。
2. 更新独立解析器与预检，补齐前景图片处理。
3. 收窄后端 Planner Prompt 并添加回归测试。
4. 执行本地测试、gold benchmark 和可用的真实 Planner benchmark。
5. 因改动会影响服务器运行链路，构建 Linux 交付物、部署、最小冒烟并回填记录；失败时保留证据并停止归档。

## Open Questions

- 无；真实模型 benchmark 是否可执行由现有本地模型环境决定，不改变实现范围。
