package utils

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	arkModel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

type ChatModelConfig struct {
	MaxTokens   int
	Temperature float32
	TopP        float32
	JSONSchema  *openai.ChatCompletionResponseFormatJSONSchema
}

type Option func(*ChatModelConfig)

func WithMaxTokens(max int) Option {
	return func(c *ChatModelConfig) {
		c.MaxTokens = max
	}
}

func WithTemperature(temp float32) Option {
	return func(c *ChatModelConfig) {
		c.Temperature = temp
	}
}

func WithTopP(topP float32) Option {
	return func(c *ChatModelConfig) {
		c.TopP = topP
	}
}

func WithJSONSchema(schema *openai.ChatCompletionResponseFormatJSONSchema) Option {
	return func(c *ChatModelConfig) {
		c.JSONSchema = schema
	}
}

func LoadEnvConfig(cfg *ChatModelConfig) {
	cfg.MaxTokens = 4096
	cfg.Temperature = 0
	cfg.TopP = 0
}

func NewChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	modelType := strings.ToLower(os.Getenv("MODEL_TYPE"))

	// Create Ark ChatModel when MODEL_TYPE is "ark"
	if modelType == "ark" {
		cm, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
			APIKey:  os.Getenv("ARK_API_KEY"),
			Model:   os.Getenv("ARK_MODEL"),
			BaseURL: os.Getenv("ARK_BASE_URL"),
			Thinking: &arkModel.Thinking{
				Type: arkModel.ThinkingTypeDisabled,
			},
		})
		if err != nil {
			log.Fatalf("ark.NewChatModel failed: %v", err)
			return nil, err
		}
		return cm, nil
	}

	// Create OpenAI ChatModel (default)
	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		Model:   os.Getenv("OPENAI_MODEL"),
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
	})
	if err != nil {
		log.Fatalf("openai.NewChatModel failed: %v", err)
		return nil, err
	}
	return cm, nil
}
