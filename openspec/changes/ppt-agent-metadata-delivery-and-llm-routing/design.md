## Context

The initial generation path always constructs `PPTTaskDeepAgent`, but its observable routing is produced by a rule-first classifier and a second rule-based router pass. The main Agent prompt then asks the model to inspect PPTX files after every generation window and at the end. Meanwhile the task manager already reconciles files into `tasks.json` and RuntimeMeta, so two different actors compete to decide whether delivery is complete. This causes 18/18 tasks to remain running while the model performs redundant final checks.

QA is already intended to be disabled for cost and latency, but route defaults and logs still advertise `plan -> generate -> qa -> fix`; the initial DeepAgent also always registers Fixer and can conditionally register Reviewer.

## Goals / Non-Goals

**Goals:**

- Use one LLM call as the source of intent and Agent routing metadata.
- Keep routing output inside supported values through deterministic schema validation.
- Make code-maintained delivery metadata the only terminal completion gate.
- Stop the main Agent promptly after every planned slide is delivered.
- Remove QA/automatic fixing from initial generation and its visible pipeline.
- Preserve explicit post-delivery user correction through the existing Fixer path.

**Non-Goals:**

- Introduce a new physical QuickAgent implementation in this change.
- Remove Reviewer/Fixer source code that remains useful for manual workflows or future experiments.
- Eliminate filesystem access from backend code; backend reconciliation still converts observed output files into metadata, while the LLM no longer performs filesystem verification.
- Change public task or SSE endpoint shapes.

## Decisions

### One LLM result owns classification and routing

Extend the structured classification response with `agent_type`, `pipeline`, and `concurrency`. The learning engine calls the classifier once, then passes that result into the router. The router validates the model recommendation and applies cost/priority calculations without reclassifying or keyword matching.

Only currently executable initial-generation Agent types are accepted. Unsupported values fall back to `deep`; pipeline stages are normalized to `plan` and `generate`, and QA/Fix flags are always skipped. This gives the model semantic control without allowing it to select nonexistent runtime implementations.

Alternative considered: keep rule classification as a fallback. Rejected because the requirement explicitly removes rule matching and rules were producing misleading high-level observability. Failure fallback is a fixed operational route, not semantic keyword matching.

### Delivery tracker is the terminal authority

The task manager owns a per-task delivery snapshot containing planned count, delivered count, declared output names, pending task IDs, and completion state. The existing backend reconciler may observe output files and update `tasks.json`, but terminal code reads the snapshot rather than asking the Agent to check files or parsing its prose.

When the snapshot reaches nonzero `done == total`, the progress monitor cancels the Agent run with a dedicated internal completion cause. The task manager recognizes that cause as successful metadata completion, persists the snapshot, emits `complete`, and does not run a second terminal filesystem validation pass.

Alternative considered: wait for the Agent to return naturally and only validate metadata afterward. Rejected because it leaves the current 18/18 stall intact.

### Initial generation has no QA or automatic Fixer

DeepAgent registers only SlideExecutor for initial generation. Reviewer and Fixer are removed from its sub-agent list, and the main prompt contains only planning and windowed generation. The explicit continuation handler may still instantiate Fixer for user-requested page edits.

Route defaults and visible status always use `plan -> generate` and `QA: disabled`, independent of environment variables. This avoids configuration drift where an old `ENABLE_QA=true` silently reactivates cost-heavy behavior.

### Prompt tools match the reduced control loop

The main Agent keeps manifest, read, and search tools. Bash and batch conversion are removed from the main Agent tool list because final file inspection and thumbnail preparation belong to backend code. SlideExecutor retains Python/search/read tools required to generate each page.

## Risks / Trade-offs

- [Risk] A generator process reports success but produces a corrupt PPTX. -> Metadata completion only establishes delivery, not visual correctness; preview conversion remains a separate observable failure and manual regeneration path.
- [Risk] The Agent is cancelled immediately after the last page and loses its closing prose. -> The task manager emits a deterministic delivery-complete message from metadata.
- [Risk] LLM routing returns malformed JSON. -> Use one deterministic deep/plan/generate fallback and clearly label the routing source as fallback.
- [Risk] Polling delay adds up to two seconds after the last page. -> Trigger metadata synchronization from existing file-ready handling and retain polling as a fallback.
- [Risk] Existing tests expect rule-only classification. -> Replace production classifier tests with LLM-first and fallback-contract tests; legacy offline rule evaluation is no longer a release gate.

## Migration Plan

1. Deploy backend with metadata completion and LLM-first routing while keeping API contracts unchanged.
2. Restart the service so no old task goroutine continues with the former prompt.
3. Verify a generated task reaches complete immediately at N/N, with no subsequent `bash` verification tool event and no QA stage.
4. Roll back by restoring the previous binary; persisted task records and `tasks.json` remain compatible.

## Open Questions

- A distinct QuickAgent can be introduced later; until then the router exposes only executable initial-generation choices and records model reasoning for future split decisions.
