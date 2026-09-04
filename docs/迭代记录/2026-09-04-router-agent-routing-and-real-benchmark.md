# RouterAgent 路由闭环与真实模型 Benchmark

## 交付内容

- 恢复并独立实现 `RouterAgent`：默认会话由它区分闲聊、规划、新建和已有 PPT 修复；它只输出结构化路由，不拥有任务写入、渲染或工具权限。
- 明确的新建 PPT 请求进入 `PPTPlanner`；已有 Deck 的指定页调整进入 `PPTFixer`；未选择已有任务的页修改请求只提示选择任务，不会猜测目标任务。
- 工作台的“PPT 生成”选择是显式指令，后端直接返回 `create / prepare_create`，绕过 RouterAgent；前端也把该选择作为最高优先级，不因分类结果降级为聊天。
- 统一 Router 与确定性 fallback：主题明确的新建请求即使没有受众或风格也交给 Planner 补齐；仅纯空泛的“帮我做个 PPT”保留澄清。
- 增补 Router、Planner、Reviewer、Fixer 的 test/validation fixture，使每个分类至少 10 条；恢复 `cmd/pptbench` 可执行 runner。

## 本地验证

```text
cd ppt-agent/backend
go test ./pkg/agent/router ./pkg/runtime/web ./cmd/pptbench -count=1
go build ./...

cd ppt-agent/frontend
npm test -- --run
npm run build
python .../frontend-design-premium/scripts/audit_project.py frontend --mode strict
```

上述命令通过；前端静态审计为 0 finding。

## 真实模型 Benchmark（DeepSeek）

运行时通过 `deepseek/deepseek-chat(backup-0)`，凭据只注入 benchmark 进程，未写入仓库或本记录。

| 套件 | 数据集 | 结果 | 运行产物 |
| --- | --- | --- | --- |
| Router | test | 10/10，均分 5.0 | `benchmark/runs/20260904-171100-test-router-rerun` |
| Router | validation | 10/10，均分 5.0 | `benchmark/runs/20260904-171000-validation-router-rerun` |
| Planner | test | 5/11，均分 3.0 | `benchmark/runs/20260904-171200-test-planner` |
| Reviewer | test | 7/10，均分 4.0 | `benchmark/runs/20260904-172000-test-reviewer` |
| Fixer | test | 6/10，均分 3.5 | `benchmark/runs/20260904-172100-test-fixer` |

Planner、Reviewer、Fixer 的失败均保留为真实质量信号：Planner 主要是未产出 `tasks.draft.json` 后的恢复路径将长输入回填为占位内容；Reviewer/Fixer 分别仍有视觉策略、图表标签、结论页与局部样式/数值/时间线修订不足。本次未篡改 validation 样本或用 mock 掩盖这些问题。

## 上线证据

- 部署目标：`remote-dev:/ppt/ppt-agent`，2026-09-04。
- 服务进程：`/ppt/ppt-agent-linux -mode web -addr :8080`，最终 PID `2474967`，监听 `:8080`；启动日志确认 MySQL 已连接且无立即失败。
- 前端静态产物部署至服务实际读取的 `/ppt/frontend/dist`；`GET /`、`GET /api/health`、`GET /api/templates/layouts` 均成功。
- 线上 API 冒烟结果：
  - 默认“为新员工制作 3 页安全培训 PPT” → `create / pptagent / prepare_create`；
  - 手动“PPT 生成” → `create / pptagent / prepare_create`；
  - 未绑定任务“把第 2 页标题改短一点” → `fix / ask_clarification`。
- 冒烟仅创建会话任务，未启动 PPT 渲染；任务已逐一删除。
