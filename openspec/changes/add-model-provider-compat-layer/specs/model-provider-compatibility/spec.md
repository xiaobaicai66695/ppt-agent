## ADDED Requirements

### Requirement: Provider-aware model chain resolution
The system SHALL resolve model fallback chains as provider-aware model specifications containing at least provider, model name, API key source, base URL, timeout, and provider-specific options.

#### Scenario: New provider-aware configuration is present
- **WHEN** `MODEL_CHAIN` and matching `MODEL_<ENTRY>_*` variables are configured
- **THEN** the system resolves each entry into an ordered provider-aware model specification
- **AND** fallback order follows `MODEL_CHAIN`

#### Scenario: Legacy Ark configuration is used
- **WHEN** provider-aware model configuration is absent
- **THEN** the system resolves models from `ARK_MODEL` and `ARK_MODEL_BACKUP*`
- **AND** existing Ark-only deployments continue to initialize without configuration changes

### Requirement: Provider adapter creation
The system SHALL create Eino `model.ToolCallingChatModel` instances through provider adapters instead of hardcoding a single upstream package in the fallback model factory.

#### Scenario: Ark model creation
- **WHEN** a resolved model specification has provider `ark`
- **THEN** the system creates the model with `ark.NewChatModel`
- **AND** Ark-specific base URL, region, thinking, and JSON schema options are applied

#### Scenario: OpenAI-compatible model creation
- **WHEN** a resolved model specification has provider `openai`, `openai_compatible`, or `siliconflow`
- **THEN** the system creates the model with `openai.NewChatModel`
- **AND** OpenAI-compatible API key, base URL, timeout, token, temperature, top-p, and response format options are applied

### Requirement: SiliconFlow profile
The system SHALL support SiliconFlow as a named provider profile without requiring a dedicated Eino SiliconFlow package.

#### Scenario: SiliconFlow default base URL
- **WHEN** a model specification uses provider `siliconflow` and no explicit base URL is configured
- **THEN** the system uses `https://api.siliconflow.cn/v1` as the OpenAI-compatible base URL

#### Scenario: SiliconFlow API key resolution
- **WHEN** a model specification uses provider `siliconflow` and no account-specific key is supplied
- **THEN** the system resolves the API key from `SILICONFLOW_API_KEY`

### Requirement: Provider-safe fallback identity
The system SHALL identify fallback pause and logging targets with provider and model identity rather than model name alone.

#### Scenario: Same model name across providers
- **WHEN** two fallback entries share the same model name but use different providers
- **THEN** rate-limit pause tracking treats them as distinct targets
- **AND** logs include enough provider information to identify the active upstream

### Requirement: Provider-aware model call concurrency
The system SHALL limit concurrent model calls by provider/model/API-key resource rather than rejecting all concurrent tasks globally.

#### Scenario: Tasks use different upstream resources
- **WHEN** multiple PPT tasks are created for different provider/model/API-key resources
- **THEN** the system allows those tasks to run concurrently
- **AND** each model call is constrained only by its own upstream resource slot

#### Scenario: Tasks share the same upstream resource
- **WHEN** multiple PPT tasks call the same provider/model/API-key resource
- **THEN** model calls for that resource are serialized by default
- **AND** unrelated upstream resources are not blocked by that serialization
