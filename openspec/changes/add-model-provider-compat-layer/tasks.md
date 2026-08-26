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
