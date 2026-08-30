# 2026-08-30 独立 Deck Planner 视觉物化闭环

## 目标与职责边界

- 将图片策略、素材语义契约、下载顺序和独立验证从后端 Planner 提示词收敛到 `ppt-deck-planner` skill，使通用 Agent 即使不具备 `search_images` 工具也能按 `unsplash fetch` 物化素材。
- Planner 只写 `visual_policy`、页面 `visual_intent` 和前景 `image` 组件的 `asset_purpose`、`asset_subject`、`asset_query`、`composition`、`orientation`；不得虚构本地路径、来源或署名。
- 后端保留其确定性素材适配层，Reviewer 后统一搜索、下载、去重并写回来源信息；不再要求 Planner 直接调用下载工具。

## 实现

- skill 默认 `visual_policy.mode="required"`，纯文字 deck 必须显式使用 `mode="none"` 并说明原因；组件契约同步这一要求。
- 通用 `hydrate_unsplash_assets.ts` 同时扫描页面级视觉意图和 `components[].type="image"`，已有有效本地素材会跳过，全部成功后才写回 manifest。
- `validate_deck.py` 与 `render_deck.py` 均接入视觉素材验证；required 模式下只有查询但没有已物化本地文件/来源的 deck 会被拒绝。
- Go 契约保留 `orientation`，Prompt/类型回归覆盖视觉语义字段；Planner 不再注册图片搜索工具。评审器仅对尚未有本地素材的搜索查询施加简洁性约束。
- 修正并扩充 3 份 gold DeckSpec，使它们符合当前内容密度和组件容量门槛。

## 本地与 benchmark 证据

- `go test ./...` 通过。
- `go test ./test/plan_benchmark -v -run TestGoldDeckSpecsPassReviewer -count 1`：5/5 通过。
- Python `compileall` 通过，skill 32 个单元测试通过；minimal 示例的 `validate_deck.py` 与 `render_deck.py` 均通过。
- 真实 Planner 首稿 benchmark 已按 `PPT_BENCH_RUN_PLANNER=true`、1 个样例、20 分钟上限执行；模型初始化成功后首个请求由上游返回 HTTP 402，未产生 DeckSpec。该外部计费/额度问题不影响确定性测试或运行服务，但实际模型首稿分数需要在上游恢复后重跑。

## 部署与线上冒烟

- 目标：`remote-dev:/ppt/ppt-agent`；2026-08-30 21:15（Asia/Shanghai）启动新进程 PID `746358`，cwd `/ppt/ppt-agent/backend`，监听 `:8080`。
- 二进制：`/ppt/ppt-agent-linux`，SHA-256 `164bcf50a9ca71dc94ac456b1caea1737cdaef299031ac9a090464c40502854f`；旧二进制和两个 skill 根目录已按时间戳备份以支持回退。
- 同步了 `/ppt/ppt-agent/skills/ppt-deck-planner` 与 `/ppt/skills/ppt-deck-planner`，且不包含认证文件或用户交付物。
- `GET /api/health`、`/health/ready`、`/api/templates/layouts`、`/`、`/dashboard` 均为 HTTP 200。
- 远端 `validate_deck.py` 和 `render_deck.py` 对 minimal 示例成功生成 29,467 字节 PPTX，随后已删除该冒烟文件。
- 已清理远端上传暂存包；桌面安全策略拒绝删除本轮本地生成 PPTX、传输二进制/包和 benchmark 目录，保留其精确路径待允许的本地清理流程处理。保留远端回退备份。

## 后续风险

- 上游模型 HTTP 402 解除后，应重跑同一真实 Planner benchmark；无需改动 skill 或重新部署。
