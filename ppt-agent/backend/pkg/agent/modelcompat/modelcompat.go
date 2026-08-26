package modelcompat

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	arkext "github.com/cloudwego/eino-ext/components/model/ark"
	openaiext "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	arkmodel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

type Provider string

const (
	ProviderArk          Provider = "ark"
	ProviderOpenAI       Provider = "openai"
	ProviderOpenAICompat Provider = "openai_compatible"
	ProviderSiliconFlow  Provider = "siliconflow"
	ProviderDeepSeek     Provider = "deepseek"
	ProviderQwen         Provider = "qwen"

	SiliconFlowBaseURL = "https://api.siliconflow.cn/v1"
	DefaultTimeout     = 10 * time.Minute
)

type ModelSpec struct {
	Provider            Provider
	Model               string
	APIKey              string
	BaseURL             string
	Region              string
	Timeout             time.Duration
	MaxTokens           *int
	MaxCompletionTokens *int
	Temperature         *float32
	TopP                *float32
	DisableThinking     *bool
	JSONSchema          *openaiext.ChatCompletionResponseFormatJSONSchema
	ExtraFields         map[string]any
}

type Factory interface {
	NewToolCallingChatModel(ctx context.Context, spec ModelSpec) (model.ToolCallingChatModel, error)
}

type DefaultFactory struct{}

func (DefaultFactory) NewToolCallingChatModel(ctx context.Context, spec ModelSpec) (model.ToolCallingChatModel, error) {
	spec = NormalizeSpec(spec)
	switch spec.Provider {
	case ProviderArk:
		return arkext.NewChatModel(ctx, BuildArkConfig(spec))
	case ProviderOpenAI, ProviderOpenAICompat, ProviderSiliconFlow:
		return openaiext.NewChatModel(ctx, BuildOpenAIConfig(spec))
	default:
		return nil, fmt.Errorf("unsupported model provider %q", spec.Provider)
	}
}

func NormalizeProvider(provider string) Provider {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	switch normalized {
	case "", string(ProviderArk):
		return ProviderArk
	case "openai":
		return ProviderOpenAI
	case "openai_compatible", "openai_compat", "compatible":
		return ProviderOpenAICompat
	case "siliconflow", "silicon_flow", "硅基流动":
		return ProviderSiliconFlow
	case "deepseek", "deep_seek":
		return ProviderDeepSeek
	case "qwen", "dashscope", "aliyun":
		return ProviderQwen
	default:
		return Provider(normalized)
	}
}

func NormalizeSpec(spec ModelSpec) ModelSpec {
	spec.Provider = NormalizeProvider(string(spec.Provider))
	spec.Model = strings.TrimSpace(spec.Model)
	spec.APIKey = strings.TrimSpace(spec.APIKey)
	spec.BaseURL = strings.TrimSpace(spec.BaseURL)
	spec.Region = strings.TrimSpace(spec.Region)
	if spec.Timeout <= 0 {
		spec.Timeout = DefaultTimeout
	}
	if spec.Provider == ProviderSiliconFlow && spec.BaseURL == "" {
		spec.BaseURL = SiliconFlowBaseURL
	}
	if spec.ExtraFields != nil {
		copied := make(map[string]any, len(spec.ExtraFields))
		for k, v := range spec.ExtraFields {
			copied[k] = v
		}
		spec.ExtraFields = copied
	}
	return spec
}

func DisplayName(spec ModelSpec, fallbackIndex int) string {
	spec = NormalizeSpec(spec)
	name := spec.Model
	if name == "" {
		name = "unknown"
	}
	if fallbackIndex >= 0 {
		return fmt.Sprintf("%s/%s(backup-%d)", spec.Provider, name, fallbackIndex)
	}
	return fmt.Sprintf("%s/%s", spec.Provider, name)
}

func TrackerKey(spec ModelSpec) string {
	spec = NormalizeSpec(spec)
	if spec.Model == "" {
		return string(spec.Provider)
	}
	return fmt.Sprintf("%s:%s", spec.Provider, spec.Model)
}

func ConcurrencyKey(spec ModelSpec) string {
	spec = NormalizeSpec(spec)
	modelName := spec.Model
	if modelName == "" {
		modelName = "unknown"
	}
	keyHash := "default"
	if spec.APIKey != "" {
		sum := sha256.Sum256([]byte(spec.APIKey))
		keyHash = hex.EncodeToString(sum[:])[:12]
	}
	return fmt.Sprintf("%s:%s:key:%s", spec.Provider, modelName, keyHash)
}

func BuildArkConfig(spec ModelSpec) *arkext.ChatModelConfig {
	spec = NormalizeSpec(spec)
	conf := &arkext.ChatModelConfig{
		APIKey:              spec.APIKey,
		BaseURL:             spec.BaseURL,
		Region:              spec.Region,
		Model:               spec.Model,
		Timeout:             &spec.Timeout,
		MaxTokens:           spec.MaxTokens,
		MaxCompletionTokens: spec.MaxCompletionTokens,
		Temperature:         spec.Temperature,
		TopP:                spec.TopP,
	}
	if shouldDisableThinking(spec) {
		conf.Thinking = &arkmodel.Thinking{Type: arkmodel.ThinkingTypeDisabled}
	}
	if spec.JSONSchema != nil {
		conf.ResponseFormat = &arkext.ResponseFormat{
			Type: arkmodel.ResponseFormatJSONSchema,
			JSONSchema: &arkmodel.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:        spec.JSONSchema.Name,
				Description: spec.JSONSchema.Description,
				Schema:      spec.JSONSchema.JSONSchema,
				Strict:      spec.JSONSchema.Strict,
			},
		}
	}
	return conf
}

func BuildOpenAIConfig(spec ModelSpec) *openaiext.ChatModelConfig {
	spec = NormalizeSpec(spec)
	conf := &openaiext.ChatModelConfig{
		APIKey:              spec.APIKey,
		BaseURL:             spec.BaseURL,
		Model:               spec.Model,
		Timeout:             spec.Timeout,
		MaxTokens:           spec.MaxTokens,
		MaxCompletionTokens: spec.MaxCompletionTokens,
		Temperature:         spec.Temperature,
		TopP:                spec.TopP,
		ExtraFields:         spec.ExtraFields,
	}
	if spec.JSONSchema != nil {
		conf.ResponseFormat = &openaiext.ChatCompletionResponseFormat{
			Type:       openaiext.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: spec.JSONSchema,
		}
	}
	if shouldDisableThinking(spec) {
		if conf.ExtraFields == nil {
			conf.ExtraFields = map[string]any{}
		}
		if _, ok := conf.ExtraFields["enable_thinking"]; !ok {
			conf.ExtraFields["enable_thinking"] = false
		}
	}
	return conf
}

func ResolveProviderAPIKey(provider Provider, explicitKey string) string {
	if key := strings.TrimSpace(explicitKey); key != "" {
		return key
	}
	switch NormalizeProvider(string(provider)) {
	case ProviderArk:
		return strings.TrimSpace(os.Getenv("ARK_API_KEY"))
	case ProviderOpenAI:
		return strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	case ProviderSiliconFlow:
		return strings.TrimSpace(os.Getenv("SILICONFLOW_API_KEY"))
	case ProviderDeepSeek:
		return strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	case ProviderQwen:
		return strings.TrimSpace(os.Getenv("DASHSCOPE_API_KEY"))
	default:
		return strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	}
}

func shouldDisableThinking(spec ModelSpec) bool {
	if spec.DisableThinking != nil {
		return *spec.DisableThinking
	}
	modelName := strings.ToLower(spec.Model)
	return strings.Contains(modelName, "deepseek") ||
		strings.Contains(modelName, "qwen3.5") ||
		strings.Contains(modelName, "qwen3.6") ||
		strings.Contains(modelName, "kimi-k2")
}
