## Why

PPT pages can already reach 18/18 while the task remains stuck in a model-driven "file verification" step, adding latency and creating a false impression that delivery is incomplete. The current rule-first intent classifier can also emit `unknown / 0 pages / 0%` and still advertise a QA pipeline that the product has decided to abandon.

## What Changes

- Make code-maintained delivery metadata the authority for generation completion and terminate the main Agent as soon as every planned slide is marked delivered.
- Remove model instructions and tools that ask the main Agent to inspect generated PPTX files after page generation.
- Permanently remove Reviewer and automatic Fixer stages from initial PPT generation; the explicit post-delivery page-edit flow remains available.
- Replace rule-first intent classification with one LLM routing call that returns intent, domain, complexity, page estimate, Agent type, pipeline, concurrency, and reasoning.
- Validate LLM routing output in code and fall back to a conservative generation route only when the model call or schema fails.
- Update observable step output so it reports the actual LLM routing source and always shows QA as disabled.

## Capabilities

### New Capabilities
- `ppt-agent-llm-routing`: LLM-first structured routing for PPT requests with validated, non-rule-based Agent selection and a deterministic failure fallback.

### Modified Capabilities
- `ppt-agent-runtime-harness`: Make runtime delivery metadata authoritative for terminal completion and remove LLM-driven final filesystem verification.
- `ppt-agent-developer-status-ui`: Report metadata-driven delivery completion and the real QA-disabled pipeline in user-visible task status.

## Impact

- Backend intent classifier, router, learning engine, DeepAgent construction, task manager lifecycle, runtime metadata, prompts, and focused tests.
- Frontend task phase labels and workbench tests where completion wording depends on the delivery phase.
- Existing task API and SSE shapes remain backward compatible; semantics of completion and routing status become stricter.
- Reviewer/Fixer implementation remains in the repository for explicit user-requested corrections, but is no longer registered in initial generation.
