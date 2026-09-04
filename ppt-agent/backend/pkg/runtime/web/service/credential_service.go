package service

import (
	"os"
	"strings"

	"github.com/cloudwego/ppt-agent/pkg/agent/modelcompat"
	"github.com/cloudwego/ppt-agent/pkg/db"
	webmodel "github.com/cloudwego/ppt-agent/pkg/runtime/web/model"
	"github.com/cloudwego/ppt-agent/pkg/utils/logger"
)

// ResolveUserModelCredential selects the account override when present and
// otherwise resolves the configured system provider. The API key never leaves
// this service boundary except as an input to the model factory.
func ResolveUserModelCredential(userID int) webmodel.ModelCredential {
	provider := DefaultModelProvider()
	accountKey := ""
	if userID > 0 && db.DB != nil {
		key, err := db.GetUserAPIKey(uint(userID))
		if err != nil {
			logger.Warn("user_api_key_lookup_failed", "user_id", userID, "error", err.Error())
		} else if key != nil {
			provider = modelcompat.NormalizeProvider(key.Provider)
			accountKey = key.APIKey
		}
	}
	return webmodel.ModelCredential{
		Provider: string(provider),
		APIKey:   modelcompat.ResolveProviderAPIKey(provider, accountKey),
	}
}

// DefaultModelProvider resolves the primary provider using the same chain
// precedence as the legacy Web handler.
func DefaultModelProvider() modelcompat.Provider {
	for _, entry := range strings.Split(os.Getenv("MODEL_CHAIN"), ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		entryKey := providerEnvKey(entry)
		if provider := strings.TrimSpace(os.Getenv("MODEL_" + entryKey + "_PROVIDER")); provider != "" {
			return modelcompat.NormalizeProvider(provider)
		}
	}
	if provider := strings.TrimSpace(os.Getenv("MODEL_PRIMARY_PROVIDER")); provider != "" {
		return modelcompat.NormalizeProvider(provider)
	}
	return modelcompat.NormalizeProvider(os.Getenv("MODEL_PROVIDER"))
}

// SystemProviderKeyConfigured reports whether a system key is available for
// provider without exposing the key itself.
func SystemProviderKeyConfigured(provider modelcompat.Provider) bool {
	provider = modelcompat.NormalizeProvider(string(provider))
	if modelcompat.ResolveProviderAPIKey(provider, "") != "" {
		return true
	}
	for _, entry := range append(strings.Split(os.Getenv("MODEL_CHAIN"), ","), "primary", "text", "qa") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		entryKey := providerEnvKey(entry)
		entryProvider := modelcompat.NormalizeProvider(FirstNonEmpty(
			os.Getenv("MODEL_"+entryKey+"_PROVIDER"),
			os.Getenv("MODEL_PROVIDER"),
		))
		if entryProvider != provider {
			continue
		}
		if strings.TrimSpace(os.Getenv("MODEL_"+entryKey+"_API_KEY")) != "" {
			return true
		}
		if keyEnv := strings.TrimSpace(os.Getenv("MODEL_" + entryKey + "_API_KEY_ENV")); keyEnv != "" && strings.TrimSpace(os.Getenv(keyEnv)) != "" {
			return true
		}
	}
	return false
}

// IsSupportedAccountProvider reports whether an account may select provider.
func IsSupportedAccountProvider(provider modelcompat.Provider) bool {
	switch modelcompat.NormalizeProvider(string(provider)) {
	case modelcompat.ProviderArk, modelcompat.ProviderOpenAI, modelcompat.ProviderOpenAICompat,
		modelcompat.ProviderSiliconFlow, modelcompat.ProviderDeepSeek, modelcompat.ProviderQwen:
		return true
	default:
		return false
	}
}

// MaskAPIKey returns a non-reversible display form suitable for JSON output.
func MaskAPIKey(apiKey string) string {
	apiKey = strings.TrimSpace(apiKey)
	runes := []rune(apiKey)
	if len(runes) <= 8 {
		return strings.Repeat("*", len(runes))
	}
	return string(runes[:4]) + strings.Repeat("*", 8) + string(runes[len(runes)-4:])
}

func providerEnvKey(value string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(value, "-", "_"), " ", "_"))
}

// FirstNonEmpty returns the first non-blank value from a provider config.
func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
