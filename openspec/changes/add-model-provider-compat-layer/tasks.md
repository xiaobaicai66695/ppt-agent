## 1. Compatibility Layer

- [x] 1.1 Create `pkg/agent/modelcompat` with provider constants, `ModelSpec`, profile defaults, and provider-aware display/identity helpers.
- [x] 1.2 Implement provider adapters for Ark and OpenAI-compatible model creation, including SiliconFlow profile defaults.
- [x] 1.3 Add focused tests for provider defaults, API key/base URL resolution, display names, and unsupported provider errors.

## 2. Existing Model Factory Integration

- [x] 2.1 Refactor `pkg/agent/utils/model.go` to resolve fallback chains as `modelcompat.ModelSpec` while preserving old `ARK_*` behavior.
- [x] 2.2 Preserve existing wrappers: fallback call/stream retry, stream tool-call delta sanitation, compression, and runtime status injection.
- [x] 2.3 Update rate-limit pause keys and logs to include provider identity.

## 3. Validation and Documentation

- [x] 3.1 Extend existing model utility tests for legacy Ark fallback and provider-aware chain parsing.
- [x] 3.2 Run focused backend tests for `pkg/agent/modelcompat`, `pkg/agent/utils`, and impacted deck model construction paths.
- [x] 3.3 Update implementation notes or iteration archive with provider compatibility behavior and remaining DeepSeek/Qwen follow-up.

## 4. Provider-aware Runtime Concurrency

- [x] 4.1 Remove the task-level single-running-task rejection from `TaskManager.CreateTask`.
- [x] 4.2 Add provider/model/API-key resource identities for model-call concurrency control without exposing raw keys.
- [x] 4.3 Gate `FallbackChatModel.Generate` and `Stream` by resource-local concurrency slots while preserving fallback behavior.
- [x] 4.4 Add focused tests for resource key isolation and per-resource serialization.
- [x] 4.5 Re-run backend validation and update deployment notes.

## 5. Account Key and Observable Timeline UI

- [x] 5.1 Extend account API key UI so users must choose the upstream provider before saving their own key.
- [x] 5.2 Carry account key provider through backend task config and apply the key only to matching provider model specs.
- [x] 5.3 Support DeepSeek and Qwen as OpenAI-compatible provider profiles with safe default base URLs and env fallback.
- [x] 5.4 Record visible assistant answer chunks as sanitized runtime timeline events and render assistant text/tool cards in event order.
- [x] 5.5 Add focused frontend/backend tests and update iteration/deployment notes.
