## Context

当前基线已移除历史 `pptbench` CLI；Planner 基准通过 Go test harness 运行。Chat 由持久会话 handler 在路由后调用，可选地叠加 web/image 资料和模型 fallback，因此最终回答层面的回归尚不可见。当前搜索提供方可能返回低质量或重定向 URL，且 SSE 必须保持 Markdown 链接跨分段完整。

## Goals / Non-Goals

**Goals:**

- Evaluate the same production chat reply assembly used by the workbench.
- Keep deterministic fixture tests independent from model availability, then use a bounded model/Judge suite for semantic quality.
- Gate direct answers, contextual material requests, safe capability degradation and source/image attribution.

**Non-Goals:**

- Replace the external search provider or promise factual correctness without a source.
- Turn chat into a PPT-generation workflow or alter task ownership semantics.
- Store benchmark outputs in production task data.

## Decisions

- Add `chat` as a first-class Go test benchmark suite, rather than embedding checks in router. Router stops at `intent=chat`; chat evaluates final response behavior and therefore needs its own runner.
- Use fixture-provided augmentations for baseline cases, rather than hitting search/image APIs during every benchmark. This makes coverage, source validation and fallback behavior reproducible; selected model cases use the production response prompt with the same fixture evidence.
- Validate source URLs before rendering supplement Markdown. Only HTTP(S) absolute URLs with a host are retained; malformed, unsafe and duplicate URLs are omitted. This prevents model/search text from creating broken source links.
- Give retrieval-backed responses precedence over generic route fallback text. When usable retrieval evidence exists, the reply must summarize that evidence before the generic PPT invitation can appear.
- Hold test and validation apart by travel/topic, intent, evidence and expected constraints. Both are at least ten samples and protected by the existing offline dataset check.

## Risks / Trade-offs

- [Judge variation] → use deterministic assertions for safety, source and fallback contracts; Judge scores narrative usefulness only.
- [Search sources are uneven] → filter obvious malformed URLs and benchmark source presentation, while retaining provider-independent diagnostics for source authority.
- [Model unavailable] → fixture fallback remains testable and benchmark output reports model errors rather than concealing them.
- [Existing dirty worktree] → add only new benchmark/adapter files and narrow edits to chat helpers; do not reformat or revert adjacent user changes.

## Migration Plan

1. Add production reply adapter and fixtures behind the current Go test harness.
2. Run local tests and test/validation chat suites.
3. Build Linux artifacts, deploy with a timestamped backup, and smoke the unauthenticated boundary plus a guest/non-production conversation if safely available.
4. Roll back by restoring the prior binary/frontend/skill backup if health or conversation smoke fails.

## Open Questions

- None for the first release; source authority scoring can be expanded after collecting benchmark outcomes.
