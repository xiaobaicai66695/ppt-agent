# Workbench Chat Benchmark

该 benchmark 覆盖会话路由已经进入 `chat` 之后的最终回复质量，而不是只验证 intent。默认运行使用固定的检索/图片 fixture，调用生产 `BuildChatReplyForBenchmark`，不会访问网络或模型；它只承担低成本的确定性契约回归，不冒充模型评测。

```powershell
cd ppt-agent/backend
go test ./test/chat_benchmark -v -count 1
```

`testdata/cases.json` 与 `validation_cases.json` 各至少 10 条。两套数据不得复用相同 ID 或规范化用户请求；测试集用于修复回归，validation 用作 holdout。

默认质量门包括：检索证据优先于泛化 PPT 引导、上下文主题补全、无图片能力时如实降级、来源/图片 URL 仅允许有效 HTTP(S) 地址且去重、Markdown 链接在 SSE 分段拼接后仍完整。

## 真实模型评测

真实模型评测使用生产 `MODEL_TEXT_*` / `ARK_TEXT_MODEL` fallback 链，并对 test 与 validation 的 20 条 fixture 逐条调用模型。检索和图片结果仍由 fixture 提供，因此不会把外部搜索的不稳定性混进模型质量结果。每条回复会写到 `results/live-*`，该目录已忽略，不会提交。

```powershell
cd D:\environment\codeGo\llm-examples\projects\ppt-agent\backend
$env:PPT_CHAT_BENCH_RUN_LIVE="true"
$env:PPT_CHAT_BENCH_LIMIT="0" # 0 表示全部 20 条
go test ./test/chat_benchmark -v -run TestChatBenchmarkLiveModelResponses -count 1 -timeout 20m
```

先做低成本真实模型冒烟时，将 `PPT_CHAT_BENCH_LIMIT` 设为 `1`。模型初始化和每条用例的默认超时分别为 45 秒，可通过 `PPT_CHAT_BENCH_MODEL_INIT_TIMEOUT` 与 `PPT_CHAT_BENCH_CASE_TIMEOUT` 调整。
