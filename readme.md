# AI SDK

<p align="left">
    <a href="https://github.com/xerohard/ai/releases/tag/v2.0.0">
        <img src="https://img.shields.io/badge/v2.0.0-blue.svg" alt="v2.0.0">
    </a>
    <img src="https://img.shields.io/badge/Go-00ADD8?logo=go&labelColor=white" alt="Go">
    <img src="https://img.shields.io/badge/License-MIT-green" alt="License">
    <a href="https://github.com/xerohard/ai">
        <img src="https://img.shields.io/github/stars/xerohard/ai?style=social" alt="GitHub stars">
        <img src="https://img.shields.io/github/forks/xerohard/ai?style=social" alt="GitHub forks">
        <img src="https://img.shields.io/github/issues/xerohard/ai" alt="GitHub issues">
        <img src="https://img.shields.io/github/last-commit/xerohard/ai" alt="Last commit">
    </a>
    <br/>
</p>

A simple Go SDK for interacting with LLM providers. Supports streaming completions, custom instructions, and easy provider integration.

## Features

- Easy to use Go SDK
- Chat completions (non-streaming and streaming)
- Easily switch between providers and models
- Options for customizing requests (model, system prompt, max tokens, temperature, reasoning effort)

## Upcoming Features

- Support for tools calling

## Providers

   <div align="left">
    <img src="https://img.shields.io/badge/GroqCloud-FF6F00" alt="GroqCloud">
    <img src="https://img.shields.io/badge/Mistral-1976D2" alt="Mistral">
    <img src="https://img.shields.io/badge/OpenRouter-43A047" alt="OpenRouter">
    <img src="https://img.shields.io/badge/OpenAI-6E4AFF" alt="OpenAI">
    <img src="https://img.shields.io/badge/Perplexity-00B8D4" alt="Perplexity">
    <img src="https://img.shields.io/badge/Anthropic-FF4081" alt="Anthropic">
    <img src="https://img.shields.io/badge/Gemini-7C4DFF" alt="Gemini">
    <img src="https://img.shields.io/badge/Xai-FFFFFF" alt="Xai">
    <img src="https://img.shields.io/badge/Anannas-FF6F00" alt="Anannas">
    <br/>
    </div>

---

- Anannas (`Anannas`)
- Anthropic (`Anthropic`)
- Gemini (`Gemini`)
- GroqCloud (`GroqCloud`)
- Mistral (`Mistral`)
- OpenAI (`OpenAi`)
- OpenRouter (`OpenRouter`)
- Perplexity (`Perplexity`)
- Xai (`Xai`)

## Project Structure

```text
go.mod                   # Go module file
LICENSE                  # License file
readme.md                # Project documentation
ai.go                    # Main package entrypoint

base/
│  └── base.go           # Base provider
│  └── shared.go         # Shared logic
sdk/                     # Core SDK interfaces and types
│  ├── errors.go         # API errors handling
│  ├── message.go        # Message type and roles
│  ├── options.go        # Options type for request customization
│  └── provider.go       # Provider interface and SDK wrapper
providers/               # Provider implementations
│  ├── anannas.go        # Anannas provider
│  ├── anthropic.go      # Anthropic provider
│  ├── gemini.go         # Gemini provider
│  ├── groqcloud.go      # GroqCloud provider
│  ├── mistral.go        # Mistral provider
│  ├── openai.go         # OpenAI provider
│  ├── openrouter.go     # OpenRouter provider
│  └── perplexity.go     # Perplexity provider
│  └── xai.go            # Xai provider
example/                 # Example usage of the SDK
│  └── readme.md
```

## Getting Started

Import the SDK in your Go project:

```go
import "github.com/xerohard/ai/v2"
```

## Declaring Providers

To use a provider, initialize it with your API key using the provided constructor functions:

```go
// Anannas
client := ai.Anannas("YOUR_ANANNAS_API_KEY")

// Anthropic
client := ai.Anthropic("YOUR_ANTHROPIC_API_KEY")

// Gemini
client := ai.Gemini("YOUR_GEMINI_API_KEY")

// GroqCloud
client := ai.GroqCloud("YOUR_GROQ_API_KEY")

// Mistral
client := ai.Mistral("YOUR_MISTRAL_API_KEY")

// OpenAI
client := ai.OpenAi("YOUR_OPENAI_API_KEY")

// OpenRouter
client := ai.OpenRouter("YOUR_OPEN_ROUTER_API_KEY")

// Perplexity
client := ai.Perplexity("YOUR_PERPLEXITY_API_KEY")

// Xai
client := ai.Xai("YOUR_XAI_API_KEY")
```

## Usage

Create a `CompletionRequest` to specify messages, model, and other options, then call `ChatCompletion()`:

```go
resp := client.ChatCompletion(ctx, &ai.CompletionRequest{
	Messages: []ai.Message{
		{Role: "user", Content: "Your message here"},
	},
	Model:        "llama3-8b-8192",
	SystemPrompt: "You are a helpful assistant.",
	MaxTokens:    150,
	Temperature:  0.7,
	Stream:       true,
})

if resp.Error != nil {
	log.Fatalf("Chat completion failed: %v", resp.Error)
}

if resp.Stream != nil {
	defer resp.Stream.Close()
	fmt.Println("Response:")
	if _, err := io.Copy(os.Stdout, resp.Stream); err != nil {
		log.Fatalf("Failed to read stream: %v", err)
	}
	fmt.Println()
} else {
	fmt.Println("Response:", resp.Content)
}
```

### CompletionRequest Options

- `Model` (string): The model to use (e.g., "gpt-4o", "llama3-8b-8192").
- `SystemPrompt` (string): Custom system prompt to guide the AI's behavior.
- `MaxTokens` (int): The maximum number of tokens to generate.
- `ReasoningEffort` (string): Custom reasoning effort (e.g., "low", "medium", "high").
- `Temperature` (float32): Controls randomness of the output (0.0 to 1.0).
- `Stream` (bool): Set to `true` for a streaming response, `false` for a single response.

## Examples

All code examples for this SDK latest version can be found in the [ai-sdk-examples](https://github.com/xerohard/ai-sdk-examples) repository.

## Contributing

Contributions are welcome!

### Pull Requests

1.  Fork the repository.
2.  Create a new branch.
3.  Commit your changes.
4.  Push to the branch.
5.  Open a pull request.

### Issues

If you find a bug or have a feature request, please open an issue on GitHub.

## Contributors

<a href="https://github.com/xerohard/ai/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=xerohard/ai" />
</a>

---

**Note:** This project is in early development. Features, and structure may change frequently.
