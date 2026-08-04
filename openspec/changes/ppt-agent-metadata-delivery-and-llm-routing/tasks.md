## 1. LLM-first Routing

- [x] 1.1 Extend structured intent output with Agent route, pipeline, concurrency, and routing source fields
- [x] 1.2 Make the classifier call the LLM first and use a deterministic non-semantic fallback when model routing fails
- [x] 1.3 Reuse the single classification result in the router and validate it against executable initial-generation routes
- [x] 1.4 Add focused tests for LLM-first routing, invalid model output fallback, and absence of duplicate classification

## 2. QA-free Initial Generation

- [x] 2.1 Remove Reviewer and automatic Fixer registration from the initial DeepAgent
- [x] 2.2 Remove QA/Fix and model-driven filesystem verification from the main Agent prompt and tool list
- [x] 2.3 Normalize all initial generation defaults and observable pipeline output to `plan -> generate` with QA disabled
- [x] 2.4 Add prompt and Agent-construction regression tests for the reduced generation loop

## 3. Metadata-driven Delivery

- [x] 3.1 Add a code-owned delivery snapshot to TaskState and update it from backend manifest reconciliation
- [x] 3.2 Stop the main Agent with an internal success cause as soon as metadata reaches nonzero N/N
- [x] 3.3 Replace terminal disk validation with delivery snapshot validation and deterministic completion messaging
- [x] 3.4 Add regression tests for N/N early completion, incomplete metadata failure, empty metadata failure, and user cancellation

## 4. Status and Verification

- [x] 4.1 Update frontend/backend phase wording so completed metadata never remains in `checking generated files`
- [x] 4.2 Run focused Go tests, full backend build, frontend tests/build, and strict OpenSpec validation
- [x] 4.3 Deploy the Linux backend and frontend, then verify health, LLM route output, QA-disabled status, and prompt completion at N/N
- [x] 4.4 Write the iteration record and mark `PPT-FLOW-001` done with validation links
