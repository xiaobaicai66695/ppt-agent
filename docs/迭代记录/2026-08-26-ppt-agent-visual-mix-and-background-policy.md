# 2026-08-26 PPT Agent 视觉混排与背景策略上线记录

- ID: 20260826-visual-mix-background-policy
- Type: fix/deployment
- Scope: Planner/Reviewer prompt, deterministic plan review, Deck Planner contract, background and foreground image planning
- Completed: 2026-08-26 11:18 Asia/Shanghai
- Changes:
  - 将背景策略收敛为整套 PPT 只使用 2 个浅色主题背景查询，并按页面轮换；更多真实图片必须进入 `image` 组件并使用 `scene/evidence` 用途，作为图文混排素材。
  - 强化 Planner 首稿说明：普通叙事页优先安排 `image_text`、`case_study`、`example_detail`，`image_text` 在 `image_left`、`image_right`、`image_top_band` 三种版式间轮换。
  - 在 Go 审查中新增确定性质量门：拦截整套背景查询超过 2 个、叙事页图文混排不足、`image_text` 缺少页内示例图片组件、连续图文页 `layout_variant` 单一。
  - 修正 `component_contracts.json`：`image_text` 明确暴露三种已实现 variants，`title_slide/content_slide/agenda` 不再误标图文 variants。
- Verification:
  - Local: `go test ./pkg/agent/deck ./pkg/prompts` passed.
  - Local: `go build ./...` passed.
  - Local: `python -m json.tool skills/ppt-deck-planner/templates/component_contracts.json` passed.
  - Online: `/api/health` returned `{"status":"ok"}`.
  - Online: `/api/templates/layouts` returned HTTP 200; response contains `image_text` and `image_top_band`.
  - Online: deployed contract reports `title_slide_variants=[]`, `content_slide_variants=[]`, `image_text_variants=["image_left","image_right","image_top_band"]`.
  - Online: deployed binary contains the new review probes `背景查询超过`、`图文混排不足`、`layout_variant 过于单一`.
- Deployment:
  - Target: `remote-dev:/ppt/ppt-agent`.
  - Backend binary SHA-256: `f544a98b2038df5874e25cdbe74af5c0f70bf776019dd6a2d7403259f0fffd5a`.
  - Backend restarted from PID `3226031` to PID `3409972`, working directory `/ppt/ppt-agent/backend`, listening on `:8080`.
  - Rollback copy retained at `/ppt/ppt-agent/deploy-backup-20260826111653-visual-mix`.
- Cleanup:
  - Removed remote staged files `/tmp/ppt-agent-linux-visual-mix`, `/tmp/ppt-deck-planner-SKILL-visual-mix.md`, `/tmp/component_contracts-visual-mix.json`, `/tmp/generators-visual-mix.md`, and `/tmp/layouts-visual-mix.json`.
  - No model-generation smoke task was created; validation used health, layouts contract and binary/contract probes to avoid unnecessary model/image-search cost.
- Residual Risk:
  - This validates the deterministic gate and deployed contract, not full LLM behavior on a fresh user deck. A future low-cost 2-4 page generation smoke can confirm Planner compliance end to end when model spend is acceptable.

---

# 2026-08-26 PPT Agent 模型优先级与会话流式输出上线记录

- ID: 20260826-model-priority-and-streaming
- Type: fix/deployment
- Scope: backend model fallback order, ADK streaming, SSE assistant conversation
- Completed: 2026-08-26 13:36 Asia/Shanghai
- Changes:
  - 线上模型顺序调整为 `deepseek-ai/DeepSeek-V4-Flash` -> `Qwen/Qwen3.6-27B` -> `Qwen/Qwen3.5-35B-A3B` -> `Qwen/Qwen3.5-397B-A17B` -> `Qwen/Qwen3.5-122B-A10B`。
  - `ADK_ENABLE_STREAMING` 改为默认开启；显式设置 `false/0/no/off` 时可回退到阻塞输出。
  - 在 fallback ChatModel 的 `Stream()` 出口合并 tool call delta：普通 assistant 文本 chunk 继续实时进入 SSE，工具调用参数延迟到 EOF 后合并成完整消息，避免 ToolNode 收到半截 JSON。
  - 保留本轮同时进入 binary 的 Planner manifest header 兜底：模型未写 `theme/template` 时默认使用 `ocean_soft/generic`，降低 `update_tasks_manifest` 因缺少 header 反复失败的概率。
- Verification:
  - Local: `go test ./pkg/agent/deck ./pkg/agent/utils ./pkg/task ./pkg/web` passed.
  - Local: `go build ./...` passed.
  - Online: `/api/health` returned `{"status":"ok"}`.
  - Online: `/health/ready` returned `status=ok` with MySQL, Python and LibreOffice ready.
  - Online: `/api/templates/layouts` returned HTTP 200.
  - Online: `model_init_success` confirmed DeepSeek as `backup=0`, followed by the four Qwen backups as `backup=1..4`.
  - Online: low-cost 1-page smoke task `917193f8-f2d1-48dd-b39d-e3639c3af1b8` emitted incremental SSE `answer` chunks (`I`, `'ve`, ` read`, ...), confirming frontend conversation streaming is supported by the current backend route and model path.
- Deployment:
  - Target: `remote-dev:/ppt/ppt-agent`.
  - Backend restarted from PID `3409972` to PID `3442257`, working directory `/ppt/ppt-agent/backend`, listening on `:8080`.
  - Binary rollback copy retained at `/ppt/ppt-agent/ppt-agent-linux.bak-20260826-133108`.
  - Env rollback copy retained at `/ppt/ppt-agent/backend/.env.bak-20260826-120152`.
- Cleanup:
  - Smoke task was deleted after streaming evidence was captured; workdir `/ppt/ppt-agent/weboutput/1-917193f8-f2d1-48dd-b39d-e3639c3af1b8` no longer exists.
  - Removed remote staged transfer files `/tmp/ppt-agent-linux.zip` and `/tmp/ppt-agent-unpack`.
- Residual Risk:
  - Smoke intentionally stopped after verifying streaming chunks, so the task cancellation produced expected `context canceled` log lines. This validates model init order and streaming delivery, not full PPT completion quality for DeepSeek on a fresh deck.

---

# 2026-08-26 PPT Agent 首轮 DeckSpec 质量增强上线记录

- ID: 20260826-first-draft-quality-gate
- Type: fix/deployment
- Scope: Planner first-draft prompt, update_tasks_manifest preflight, Deck Planner contract
- Completed: 2026-08-26 14:08 Asia/Shanghai
- Changes:
  - Planner 首轮提示前置 `agenda`、`stat_slide`、`argument_block` 和顶层 `theme/template` 的硬约束，减少工具返回问题后多轮自修。
  - `update_tasks_manifest(mode=initialize)` 在预检前做有限确定性归一：缺失 `theme/template` 使用 `ocean_soft/generic`；`agenda` 超量时压缩到 6 个以内并补阅读路径观点；`stat_slide` 缺观点锚点时自动补 `insight`；非长论述页的短 `argument_block` 自动降级为 `insight`。
  - `argument_block` 字数问题的审查消息增加当前估算字数，避免 Planner 不知道还差多少导致二次扩写。
  - `component_contracts.json` 对齐：`agenda.max_components=6` 且推荐 `insight/key_point/toc_item`；`stat_slide` 推荐组件加入 `insight`。
- Verification:
  - Local: `python -m json.tool ppt-agent/skills/ppt-deck-planner/templates/component_contracts.json` passed.
  - Local: `go test ./pkg/agent/deck ./pkg/prompts` passed.
  - Local: `go build ./...` passed.
  - Online: `/api/health` returned `{"status":"ok"}`.
  - Online: `/health/ready` returned `status=ok` with MySQL, Python and LibreOffice ready.
  - Online: deployed contract reports `agenda_max=6`, `agenda_components=insight,key_point,toc_item`, and `stat_components=stat,number_callout,insight,source_note`.
  - Online: deployed binary contains the new first-draft probes `agenda_insight`、`insight_auto`、`argument_block 当前约`.
- Deployment:
  - Target: `remote-dev:/ppt/ppt-agent`.
  - Backend restarted from PID `3442257` to PID `3450855`, working directory `/ppt/ppt-agent/backend`, listening on `:8080`.
  - Rollback copy retained at `/ppt/ppt-agent/deploy-backup-20260826-140628-first-draft-quality`.
  - A follow-up contract-only sync corrected `agenda.max_components` from 8 to 6 without another restart.
- Cleanup:
  - Removed local transfer zip `ppt-agent/ppt-agent-linux.zip`.
  - Removed remote staged files `/tmp/component_contracts.json`, `/tmp/ppt-agent-linux-first-draft.zip`, and `/tmp/ppt-agent-first-draft-unpack`.
- Residual Risk:
  - This validates deterministic normalization and deployed contract. It does not run a full-cost fresh 10+ page international affairs deck; that should be a targeted smoke only when model/image-search spend is acceptable.

---

# 2026-08-26 PPT Agent 图片物化降级与内部摘要过滤上线记录

- ID: 20260826-asset-materialize-degrade-and-summary-filter
- Type: bugfix/deployment
- Scope: backend asset materialization, SSE assistant conversation visibility
- Completed: 2026-08-26 14:50 Asia/Shanghai
- Changes:
  - 修复 Unsplash 搜索或下载返回可恢复错误时中断整套 PPT 的问题；`HTTP 410/404/429/5xx`、内容移除、无搜索结果和短时网络调用失败现在只跳过对应背景槽位或图片组件，页面回退到浅色无图背景/无外部素材。
  - 保留配置和认证类硬失败：缺少 `UNSPLASH_ACCESS_KEY`、client nil、context canceled/deadline、`HTTP 401` 仍返回错误，避免隐藏系统配置问题。
  - 过滤内部上下文压缩摘要 JSON，包含 `user_intent_summary` 且包含 `progress_summary` 或 `conversation_summary` 的消息不再作为 `answer` SSE 事件进入前端会话。
  - 流式输出遇到疑似 JSON/code fence 开头时先缓存到消息结束后判断；普通文本仍按 chunk 实时输出，普通 Planner JSON 输出保留可见。
- Verification:
  - Local: `go test ./pkg/agent/deck` passed.
  - Local: `go test ./pkg/agent/utils ./pkg/task ./pkg/web` passed.
  - Local: `go build ./...` passed.
  - Online: `/api/health` returned `{"status":"ok"}`.
  - Online: `/health/ready` returned `status=ok` with MySQL, Python and LibreOffice ready.
  - Online: `/api/templates/layouts` and `/api/themes` returned HTTP 200.
  - Online: deployed binary contains `background_asset_skipped` and `image_asset_skipped` probes.
- Deployment:
  - Target: `remote-dev:/ppt/ppt-agent`.
  - Backend restarted from PID `3450855` to PID `3460351`, working directory `/ppt/ppt-agent/backend`, listening on `:8080`.
  - Binary rollback copies retained at `/ppt/ppt-agent/ppt-agent-linux.bak-20260826-133108` and `/ppt/ppt-agent/ppt-agent-linux.bak.`.
- Cleanup:
  - Removed remote staged file `/tmp/ppt-agent-linux-assetfix`.
  - No model-generation smoke task was created; validation used health, readiness, template/theme API and binary probes to avoid unnecessary model/image-search spend.
- Residual Risk:
  - This fixes the observed `unsplash API HTTP 410: Content removed` hard failure path and visible compression-summary leak. It does not yet correct the upstream summarizer's bad inferred values such as `5325页`; those are now hidden from users and should be handled separately if they affect model context quality.

---

# 2026-08-26 PPT Agent 工具调用去重与流式观察顺序修复上线记录

- ID: 20260826-tool-call-dedupe-and-stream-order
- Type: bugfix/deployment
- Scope: backend streaming tool-call sanitizer, image search tool cache, frontend inline runtime conversation
- Completed: 2026-08-26 16:02 Asia/Shanghai
- Changes:
  - 修复 `search_images` 同一 query 被 Planner 先查候选、后下载时重复请求 Unsplash 的问题；工具实例新增任务级搜索缓存和下载缓存，后续相同搜索参数直接复用候选结果。
  - 调整流式 tool call 清洗：继续阻止半截 JSON 参数进入 ToolNode，但当单个 tool call 参数已拼成合法 JSON 时立即释放，不再无条件等到整个模型流 EOF。
  - 前端工具卡按 `tool name + query/args` 签名配对 start/end，避免同名并发工具按队列错配。
  - 前端对重复完成的同 query 工具卡做折叠，并在搜索/搜图工具摘要和调用参数中展示 `search_reason`，让工具调用目的可直接观察。
- Verification:
  - Local: `go test ./pkg/tools/image ./pkg/agent/utils` passed.
  - Local: `npm test -- --run src/utils/workbench.test.ts` passed.
  - Local: `go test ./pkg/agent/deck ./pkg/agent/utils ./pkg/tools/image ./pkg/task ./pkg/web` passed.
  - Local: `go build ./...` passed.
  - Local: `npm run build` passed.
  - Online: `/api/health` returned `{"status":"ok"}`.
  - Online: `/health/ready` returned `status=ok` with MySQL, Python and LibreOffice ready.
  - Online: `/` returned frontend HTML successfully.
  - Online: deployed binary contains `image_search_cache_hit`, `image_download_cache_hit`, and `sanitizeStreamingToolCallDeltas` probes.
- Deployment:
  - Target: `remote-dev:/ppt/ppt-agent`.
  - Backend restarted from PID `3460351` to PID `3478414`, working directory `/ppt/ppt-agent/backend`, listening on `:8080`.
  - Frontend `dist` replaced under `/ppt/ppt-agent/frontend/dist`.
  - Rollback copies retained at `/ppt/ppt-agent/ppt-agent-linux.bak-20260826160113-tooltrace-v2` and `/ppt/ppt-agent/frontend/dist.bak-20260826160113-tooltrace-v2`.
- Cleanup:
  - Removed remote staged files `/tmp/ppt-agent-linux-tooltrace`, `/tmp/ppt-agent-frontend-dist-tooltrace.tar`, `/tmp/ppt-agent-linux-tooltrace-v2`, and `/tmp/ppt-agent-frontend-dist-tooltrace-v2.tar`.
  - Removed local transfer archive `ppt-agent/ppt-agent-frontend-dist-tooltrace.tar`.
- Residual Risk:
  - This reduces duplicate external image requests and fixes the inline display pairing/deduplication. If the selected model intentionally batches many different tool calls into one assistant turn, the UI can now show each tool's purpose, but true step-by-step ReAct alternation still depends on model/tool-call behavior.
