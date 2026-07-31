# CWK Integration 人机协作迭代工作流

>本工作流定义用户与 Agent 在 `kb-agent-integration` 项目中的协作节奏、交付物标准和需求转轨规则。

**版本**: 2.0  
**日期**: 2026-05-25  
**状态**: 已实施

---

## 1. 核心理念

- **按规模分流**: 小需求允许当场闭环；中大型需求必须先登记到 `docs/issues/todo.md`。
- **按需预研**: 只有在边界不清、依赖外部系统或存在方案权衡时，才编写 `docs/research/` 预研文档。
- **双路线推进**: 已登记事项可以直接开始做（`direct`），也可以进入 OPSX/OpenSpec 工作流（`opsx`）。
- **透明可追踪**: 每个需求的当前状态、产出物位置必须在登记表中一目了然。

---

## 2. 事项登记入口

### 2.1 文件位置

```
docs/issues/todo.md
```

### 2.2 由谁更新

- **Agent**: 对于中大型需求，或虽然是小需求但明显无法当场闭环时，第一个动作是记录或更新此文件。
- **User**: 可随时直接编辑此文件，修正或补充需求描述。

---

## 3. 需求分类与判定标准

| 类型 | 特征 | 默认动作 | 举例 |
|------|------|----------|------|
| **small** | 单点修改、已有模式复用、通常可在当前会话闭环 | 可直接开始；若中途失控则补登记 | 文档 typo、配置调整、小修复、测试补充 |
| **medium** | 涉及多个文件或模块，需要跨会话跟踪 | 必须先登记到 `todo.md` | 一组相关 CLI 改动、配置机制扩展、较大文档更新 |
| **large** | 跨层级、涉及设计决策、影响公共接口或协议 | 必须先登记到 `todo.md`，通常优先考虑 `opsx` | 新系统接入、新增 MCP 工具族、接口变更、架构调整 |

---

## 4. 流转规则

### 4.1 小需求 —— 当场闭环或升级登记

```
用户提出需求
    │
    ▼
Agent 判断是否可在当前会话内闭环
    │
    ▼
← 可闭环：直接实现并完成验证
← 不可闭环/复杂度超出预期：登记到 todo.md，再按 medium/large 规则继续
```

### 4.2 中大型需求 —— 先登记再分流

```
用户提出需求
    │
    ▼
Agent 记录到 todo.md（状态: pending）
    │
    ▼
按需补充 research（状态: researching）
    │
    ▼
选择路线：direct 或 opsx
    │
    ▼
← direct：进入实现（状态: in_progress）
← opsx：进入 openspec/changes/<change>/（状态: in_opsx）
    │
    ▼
实现完成后回填 todo.md（状态: done）
```

### 4.3 何时写 `docs/research/`

满足任一条件时，建议先补预研：

- 需求边界不清，用户口径尚未稳定
- 依赖外部系统、外部 API 或新三方库
- 有两个及以上可行方案需要比较
- 预计会影响项目架构、接口契约或长期维护成本

不满足以上条件时，不强制产出预研文档。

---

## 5. 后续工作流产出物规范

已登记事项的产出物需要在 `todo.md` 中持续回填。推荐位置如下：

| 阶段 | 标准产出物 | 路径示例 |
|------|-----------|---------|
| Research | 预研文档（可选） | `docs/research/YYYY-MM-DD-short-title.md` |
| Direct | 直接实现相关链接 | 代码路径、PR、补充文档 |
| OPSX | 正式变更目录 | `openspec/changes/<change>/` |
| Final | 最终产出 | 代码路径、测试文件、补充文档 |

> 对于 `opsx` 路线，不在本工作流中硬编码 proposal/spec/design/tasks 的固定顺序，统一以当前 OPSX schema 和 `openspec status --change <name>` 输出为准。

---

## 6. Agent 行为准则

1. **先判断规模再行动**: Agent 必须先判断事项是 `small`、`medium` 还是 `large`。
2. **中大型事项先登记**: 对 `medium` / `large` 事项，Agent 在动手前必须更新 `todo.md`。
3. **小需求失控就升级**: 原本按 `small` 处理的事项，只要出现跨会话、跨层级、边界不清等迹象，必须补登记。
4. **预研按需而不是滥写**: 只有在不确定性真实存在时才创建 `docs/research/` 文档。
5. **路线必须显式**: 已登记事项必须标明 `direct` 或 `opsx`，并在过程中持续回填链接。
6. **闭环要检查**: 涉及代码改动的事项完成后，运行 `uv run ruff format . && uv run ruff check . --fix && uv run pyright && uv run pytest -q`。
7. **每日汇报**: 在日常汇报中，列出 `todo.md` 中状态为 `researching`、`planned`、`in_progress`、`in_opsx` 和 `blocked` 的事项。

---

## 7. 示例场景

### 场景 A：简单需求（文档更新）

> 用户：“README 里架构图的一处文字写错了。”

- Agent 判断该事项可在当前会话内闭环，因此不强制登记到 `todo.md`。
- 当场修改 README 并验证。
- 若过程中发现影响多个文档或需要跨会话跟踪，再补登记。

### 场景 B：复杂需求（新系统接入）

> 用户：“接入 GitLab 代码审查功能。”

- Agent 在 `todo.md` 新增记录，`size=large`，`route=opsx`，`status=pending`。
- 用户确认边界（只做 MR 查询还是 CI 状态也要？）。
- 如需对接 API 能力，先补 `docs/research/` 预研文档。
- Agent 创建 `openspec/changes/<change>/` 并按当前 schema 推进。
- 实现完成后回填 `todo.md` 为 `done`。

### 场景 C：中等需求（直接实现）

> 用户：“给现有 ITR CLI 补一个 comment 子命令，沿用当前命令组织方式。”

- Agent 判断该事项涉及多个文件，属于 `medium`。
- 先在 `todo.md` 登记，标记 `route=direct`。
- 若无需方案权衡，则不写 `docs/research/`。
- 直接进入实现，完成后回填产出物路径和状态。

---

## 8. 关联文档

- [需求总池](../issues/todo.md)
- [预研文档规范](../research/README.md)
