# PPT Agent 模型与结构体文档索引

本文档集覆盖 `ppt-agent/backend` 的生产代码中定义的模型和结构体。阅读顺序以数据进入系统后的使用链路为准：请求与会话 → 路由与规划 → 任务执行与交付 → 持久化、观测与运营。

| 文档 | 覆盖范围 |
| --- | --- |
| [持久化模型](./ppt-agent-persistence-models.md) | MySQL GORM 模型、管理端聚合结构体，以及全部字段、约束和关联。 |
| [生产结构体总索引](./ppt-agent-production-struct-catalog.md) | 所有生产 struct 的包、源码位置、使用场景与字段职责分组。 |
| [生产 struct 逐字段参考](./ppt-agent-production-struct-fields.md) | 所有生产 struct 的字段名、Go 类型、tag/约束和逐字段说明。 |

## 范围口径

- **纳入**：`backend/main.go`、`backend/cmd/` 和 `backend/pkg/` 下的生产 Go 文件中的具名 struct；包含导出 DTO、内部运行时状态、第三方 API 载荷和配置结构。
- **不纳入正文**：`*_test.go` 中的 fake、fixture、断言和基准测试临时结构；它们不属于运行时数据契约。测试文件仍可通过源码检索定位。
- **字段说明方式**：持久化和对外 API/Agent 契约逐字段说明；仅用于包内实现的私有结构按“字段职责分组”说明，并链接到唯一源码定义，避免文档复述会频繁变动的实现细节。

## 结构与链路

```text
HTTP 请求 / 外部工具载荷
        ↓
Web 路由 DTO ──→ Router Agent ──→ DeckSpec / Plan
        ↓                              ↓
TaskInfo / TaskState ──→ 渲染、修复、交付、SSE
        ↓
TaskRecord / Conversation / RuntimeEvent / Feedback
        ↓
管理统计、失败分析与基准评估
```

新增或修改 struct 时，应同步更新总索引；若字段跨越 API、Agent、生成器或数据库边界，还应在对应专题文档补充逐字段契约说明。
