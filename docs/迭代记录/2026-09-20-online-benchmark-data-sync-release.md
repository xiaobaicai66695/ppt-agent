# 2026-09-20 线上评测证据与 Benchmark 数据同步发布

## 交付

- 生产任务在 Planner、Reviewer、PlannerRefiner、Fixer 边界写入私有、内容寻址的输入/输出证据；任务终态、反馈和删除撤回均回写评测会话。
- 管理端支持候选项审核、显式 case revision、冻结 dataset version 与校验和导出；未审核或已撤回证据不能进入发布数据集。
- `pptbench dataset pull` 将已授权 bundle 校验后写入本地私有导入根；Planner、Reviewer、Fixer 可选择导入版本，Router 保持仓库 fixture-only。

## 本地验证

- `go test ./...`：通过。
- `go build ./...`：通过。
- `openspec validate add-online-benchmark-data-sync --type change --strict`：通过。
- 覆盖证据存储、校验和 bundle、导入安装、数据集路径解析、Router fixture-only 以及评测会话唯一索引。

## 部署与冒烟

- 部署目标：`remote-dev:/ppt/ppt-agent`；按发布规则停止旧进程，完整替换 `backend` 源码与 `skills`，保留 `backend/.env`、`weboutput`。
- 交付提交：`73aea37`、`1299339`；Linux AMD64 二进制 SHA-256：`b4d6eedccc444f4d84953c36a2b5075a19235a281e09d899a7cfba80a7f4ec5d`。
- 最终进程：PID `3660166`，工作目录 `/ppt/ppt-agent/backend`，命令 `/ppt/ppt-agent-linux -mode web -addr :8080`，监听 `:8080`。
- 公网 `GET /api/health` 返回 `ok`，`GET /api/templates/layouts` 返回 200；未认证 `GET /api/admin/evaluations/candidates` 返回 401。
- 使用临时访客创建 1 页任务后，生产 MySQL 记录到 1 个评测会话、2 个阶段快照和 4 个工件；证明迁移与真实阶段捕获已生效。
- 冒烟任务、会话消息、运行事件、工作目录及临时访客账号已删除；对应评测会话和 2 个候选项已标记 withdrawn。4 个内容寻址工件按设计保留为私有孤儿，交由后续保留期清理，避免误删可被其他证据复用的内容。

## 数据治理说明

- 冒烟候选项未经过脱敏和人工审核，故没有将其提升为 case 或导出 bundle；这正是发布门禁的预期行为。
- 实际轮换由管理员以已审核 revision 冻结 validation，完成评估窗口后另建 active-dev 版本；后续 validation 会拒绝复用已暴露的 core / active-dev split group。
