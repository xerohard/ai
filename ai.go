// SDK entry point and provider initializers

package ai

import (
	"github.com/xerohard/ai/v2/providers"
	"github.com/xerohard/ai/v2/sdk"
)

type (
	Message           = sdk.Message
	SDK               = sdk.SDK
	CompletionRequest = sdk.CompletionRequest
	Tool              = sdk.Tool
	InputSchema       = sdk.InputSchema
)

func Anannas(apiKey string) *SDK {
	return sdk.NewSDK(providers.NewAnannasProvider(apiKey))
}

func Anthropic(apiKey string) *SDK {
	return sdk.NewSDK(providers.NewAnthropicProvider(apiKey))
}

func Gemini(apiKey string) *SDK {
	return sdk.NewSDK(providers.NewGeminiProvider(apiKey))
}

func GroqCloud(apiKey string) *SDK {
	return sdk.NewSDK(providers.NewGroqCloudProvider(apiKey))
}

func Mistral(apiKey string) *SDK {
	return sdk.NewSDK(providers.NewMistralProvider(apiKey))
}

func OpenAi(apiKey string) *SDK {
	return sdk.NewSDK(providers.NewOpenAiProvider(apiKey))
}

func OpenRouter(apiKey string) *SDK {
	return sdk.NewSDK(providers.NewOpenRouterProvider(apiKey))
}

func Perplexity(apiKey string) *SDK {
	return sdk.NewSDK(providers.NewPerplexityProvider(apiKey))
}

func Xai(apiKey string) *SDK {
	return sdk.NewSDK(providers.NewXaiProvider(apiKey))
}
