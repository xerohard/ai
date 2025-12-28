// Gemini provider

package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/xerohard/ai/v2/base"
	"github.com/xerohard/ai/v2/sdk"
)

type GeminiProvider struct {
	*base.Provider
	APIKey string
}

func NewGeminiProvider(apiKey string) *GeminiProvider {
	p := &GeminiProvider{
		APIKey: apiKey,
	}
	p.Provider = &base.Provider{APICaller: p}
	return p
}

type GeminiFunctionDeclaration struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  *GeminiParameters `json:"parameters,omitempty"`
}

type GeminiParameters struct {
	Type       string                       `json:"type"`
	Properties map[string]GeminiPropertyDef `json:"properties"`
	Required   []string                     `json:"required,omitempty"`
}

type GeminiPropertyDef struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type GeminiToolConfig struct {
	FunctionDeclarations []GeminiFunctionDeclaration `json:"functionDeclarations,omitempty"`
}

type GeminiPart struct {
	Text             string                  `json:"text,omitempty"`
	FunctionCall     *GeminiFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *GeminiFunctionResponse `json:"functionResponse,omitempty"`
}

type GeminiFunctionResponse struct {
	Name     string         `json:"name"`
	Response map[string]any `json:"response"`
}

type GeminiFunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type GenerationConfig struct {
	Temperature     float32 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type GeminiRequest struct {
	Contents          []GeminiContent   `json:"contents"`
	SystemInstruction *GeminiContent    `json:"system_instruction,omitempty"`
	GenerationConfig  *GenerationConfig `json:"generation_config,omitempty"`
	Tools             *GeminiToolConfig `json:"tools,omitempty"`
}

type Candidate struct {
	Content      GeminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type GeminiResponseChunk struct {
	Candidates []Candidate `json:"candidates"`
}

type PromptFeedback struct {
	BlockReason string `json:"blockReason"`
}

type GeminiResponse struct {
	Candidates     []Candidate     `json:"candidates"`
	PromptFeedback *PromptFeedback `json:"promptFeedback,omitempty"`
}

type ContentBlockedError struct {
	Reason string
	Body   []byte
}

func (e *ContentBlockedError) Error() string {
	return fmt.Sprintf("content blocked by safety filters. Finish Reason: %s. Response Body: %s", e.Reason, string(e.Body))
}

func (p *GeminiProvider) CallAPI(ctx context.Context, messages []sdk.Message, streamMode bool, opts *sdk.Options) (io.ReadCloser, error) {

	var model string
	if opts != nil && opts.Model != "" {
		model = opts.Model
	}

	var url string
	if streamMode {
		url = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", model, p.APIKey)
	} else {
		url = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, p.APIKey)
	}

	var systemInstruction *GeminiContent
	var geminiContents []GeminiContent

	for _, msg := range messages {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}

		if role == "system" {
			systemInstruction = &GeminiContent{
				Parts: []GeminiPart{{Text: msg.Content}},
			}
			continue
		}

		if role == "tool" {
			var response map[string]any
			if err := json.Unmarshal([]byte(msg.Content), &response); err != nil {
				response = map[string]any{"result": msg.Content}
			}

			functionName := ""
			if len(geminiContents) > 0 {
				lastContent := geminiContents[len(geminiContents)-1]
				for _, part := range lastContent.Parts {
					if part.FunctionCall != nil {
						functionName = part.FunctionCall.Name
						break
					}
				}
			}

			geminiContents = append(geminiContents, GeminiContent{
				Role: "function",
				Parts: []GeminiPart{{
					FunctionResponse: &GeminiFunctionResponse{
						Name:     functionName,
						Response: response,
					},
				}},
			})
			continue
		}
		var parts []GeminiPart

		if msg.Content != "" {
			parts = append(parts, GeminiPart{Text: msg.Content})
		}

		if len(msg.ToolCalls) > 0 {
			for _, toolCall := range msg.ToolCalls {
				var args map[string]any
				if err := json.Unmarshal(toolCall.Arguments, &args); err != nil {
					args = make(map[string]any)
				}

				parts = append(parts, GeminiPart{
					FunctionCall: &GeminiFunctionCall{
						Name: toolCall.Name,
						Args: args,
					},
				})
			}
		}

		if len(parts) > 0 {
			geminiContents = append(geminiContents, GeminiContent{
				Role:  role,
				Parts: parts,
			})
		}
	}

	reqBody := GeminiRequest{
		Contents:          geminiContents,
		SystemInstruction: systemInstruction,
	}

	if opts != nil {
		cfg := &GenerationConfig{}
		cfg.Temperature = 0.7

		if opts.Temperature > 0 {
			cfg.Temperature = opts.Temperature
		}

		if opts.MaxCompletionTokens > 0 {
			cfg.MaxOutputTokens = opts.MaxCompletionTokens
		}

		if len(opts.Tools) > 0 {
			toolConfig := convertSDKToolsToProviderTools(opts.Tools)
			reqBody.Tools = toolConfig
		}

		reqBody.GenerationConfig = cfg
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))

	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &sdk.APIError{
			StatusCode: resp.StatusCode,
			Message:    string(b),
			Body:       b,
		}
	}

	if !streamMode {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var response GeminiResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse non-streaming JSON response: %w. Body: %s", err, string(body))
		}

		if response.PromptFeedback != nil && response.PromptFeedback.BlockReason != "" {
			return nil, &ContentBlockedError{
				Reason: response.PromptFeedback.BlockReason,
				Body:   body,
			}
		}

		if len(response.Candidates) > 0 {
			candidate := response.Candidates[0]

			if candidate.FinishReason == "SAFETY" || candidate.FinishReason == "RECITATION" {
				return nil, &ContentBlockedError{
					Reason: candidate.FinishReason,
					Body:   body,
				}
			}

			var fullText string
			var toolCalls []sdk.ToolCallRequest
			for i, part := range candidate.Content.Parts {
				if part.Text != "" {
					fullText += part.Text
				}

				if part.FunctionCall != nil {
					argsJSON, err := json.Marshal(part.FunctionCall.Args)
					if err != nil {
						return nil, fmt.Errorf("failed to marshal function call args: %w", err)
					}

					toolCalls = append(toolCalls, sdk.ToolCallRequest{
						ID:        fmt.Sprintf("call_%d", i),
						Name:      part.FunctionCall.Name,
						Arguments: json.RawMessage(argsJSON),
					})
				}
			}

			if len(toolCalls) > 0 {
				response := &sdk.CompletionResponse{
					Content:   fullText,
					ToolCalls: toolCalls,
					Role:      "assistant",
				}
				responseJSON, err := json.Marshal(response)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal completion response: %w", err)
				}
				return io.NopCloser(bytes.NewReader(responseJSON)), nil
			}

			if fullText == "" {

				return nil, fmt.Errorf("non-streaming response body was successfully parsed but contained empty text (FinishReason: %s). Raw body: %s", candidate.FinishReason, string(body))
			}

			return io.NopCloser(bytes.NewReader([]byte(fullText))), nil
		}

		return nil, fmt.Errorf("non-streaming response body was successfully parsed but contained no candidates. Raw body: %s", string(body))
	}

	return resp.Body, nil
}

func (p *GeminiProvider) ParseResponse(body io.Reader, onChunk func(string) error) error {
	reader := bufio.NewReader(body)

	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			if bytes.HasPrefix(line, []byte("data: ")) {
				line = line[len("data: "):]
			}
			if bytes.Equal(line, []byte("[DONE]")) {
				return nil
			}

			var chunk GeminiResponseChunk

			if jsonErr := json.Unmarshal(line, &chunk); jsonErr == nil {
				if len(chunk.Candidates) > 0 {
					candidate := chunk.Candidates[0]

					if candidate.FinishReason == "SAFETY" || candidate.FinishReason == "RECITATION" {
						return &ContentBlockedError{Reason: candidate.FinishReason, Body: line}
					}

					if len(candidate.Content.Parts) > 0 {
						text := candidate.Content.Parts[0].Text
						if text != "" {
							if chunkErr := onChunk(text); chunkErr != nil {
								return chunkErr
							}
						}
					}

					if candidate.FinishReason == "STOP" || candidate.FinishReason == "MAX_TOKENS" {
						return nil
					}
				}
			} else {
				continue
			}
		}

		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func convertSDKToolsToProviderTools(sdkTools map[string]sdk.Tool) *GeminiToolConfig {
	if len(sdkTools) == 0 {
		return nil
	}

	declarations := make([]GeminiFunctionDeclaration, 0, len(sdkTools))

	for name, tool := range sdkTools {
		var required []string
		properties := make(map[string]GeminiPropertyDef)

		for propName, prop := range tool.InputSchema {
			// Convert SDK property to Gemini property (without Required field)
			properties[propName] = GeminiPropertyDef{
				Type:        prop.Type,
				Description: prop.Description,
			}

			if prop.Required {
				required = append(required, propName)
			}
		}

		declarations = append(declarations, GeminiFunctionDeclaration{
			Name:        name,
			Description: tool.Description,
			Parameters: &GeminiParameters{
				Type:       "object",
				Properties: properties,
				Required:   required,
			},
		})
	}

	return &GeminiToolConfig{
		FunctionDeclarations: declarations,
	}
}
