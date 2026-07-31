# 预研文档规范

`docs/research/` 用于存放**按需编写**的事项预研文档。

## 什么时候写

满足任一条件时，建议先写预研：

- 需求边界不清，仍需澄清范围
- 依赖外部系统、外部 API 或新三方库
- 有两个及以上可行方案需要比较
- 预计会影响架构、接口契约或长期维护成本

以下情况通常不需要单独预研：

- 只是沿用现有模式的小改动
- 路线明确，可以直接实现
- 已经存在等价的 proposal / design / spec 文档

## 命名规范

建议使用：

```text
docs/research/YYYY-MM-DD-short-title.md
```

例如：

```text
docs/research/2026-05-25-gitlab-review-integration.md
```

## 推荐结构

```md
# <事项标题> 预研

## 背景
## 目标与非目标
## 现状
## 方案选项
## 推荐方案
## 风险与待确认问题
## 结论
## 关联事项
- TODO: docs/issues/todo.md#<ID>
```

## 与 `todo.md` 的关系

- 预研文档不是事项主表，事项仍以 `docs/issues/todo.md` 为准
- 若某事项写了预研文档，应在 `todo.md` 的 `Research` 列填入路径
- 预研结论如果进入正式变更流程，应继续链接到 `openspec/changes/<change>/` 或最终代码产出
