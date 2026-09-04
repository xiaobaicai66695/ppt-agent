## ADDED Requirements

### Requirement: Model creation SHALL be separated from fallback execution
Provider configuration parsing, model factory construction, fallback chain selection, and request execution SHALL be independently readable units behind stable internal interfaces.

#### Scenario: Provider-specific model chain is preserved
- **WHEN** a role requests a configured provider/model chain
- **THEN** the same provider order, account credential selection, timeout, and fallback behavior are used after the refactor

### Requirement: Model concurrency and transient failure behavior SHALL remain stable
The model runtime SHALL preserve per-resource concurrency limits, rate-limit pauses, retry/fallback decisions, and error classification while moving limiter logic out of provider configuration code.

#### Scenario: Rate-limited primary falls back
- **WHEN** the primary model is rate-limited or returns a classified transient failure
- **THEN** the existing pause and fallback policy is applied without exposing credentials or changing the caller's error contract

### Requirement: Streaming sanitation and request observability SHALL be isolated
Streaming tool-call normalization, answer chunk handling, request metadata projection, and sensitive-value redaction SHALL be isolated from model construction and SHALL retain their current output contracts.

#### Scenario: Streamed tool-call chunks are normalized
- **WHEN** an upstream stream emits partial tool-call JSON or answer chunks
- **THEN** the consumer receives the same normalized stream semantics, and request diagnostics contain only the existing safe role/name summaries
