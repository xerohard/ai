// OpenRouter provider

package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/xerohard/ai/v2/base"
	"github.com/xerohard/ai/v2/sdk"
)

type OpenRouterProvider struct {
	*base.Provider
	APIKey string
}

func NewOpenRouterProvider(apiKey string) *OpenRouterProvider {
	p := &OpenRouterProvider{
		APIKey: apiKey,
	}
	p.Provider = &base.Provider{APICaller: p}
	return p
}

func (p *OpenRouterProvider) CallAPI(ctx context.Context, messages []sdk.Message, streamMode bool, opts *sdk.Options) (io.ReadCloser, error) {
	url := "https://openrouter.ai/api/v1/chat/completions"

	chatMessages := []map[string]string{}
	for _, m := range messages {
		chatMessages = append(chatMessages, map[string]string{
			"role":    m.Role,
			"content": m.Content,
		})
	}

	body := map[string]interface{}{
		"messages": chatMessages,
		"stream":   streamMode,
	}
	if opts != nil {
		if opts.Model != "" {
			body["model"] = opts.Model
		}
		if opts.MaxCompletionTokens != 0 {
			body["max_tokens"] = opts.MaxCompletionTokens
		}
		if opts.ReasoningEffort != "" {
			body["reasoning"] = map[string]interface{}{
				"effort": opts.ReasoningEffort,
			}
		}
		if opts.Temperature != 0 {
			body["temperature"] = opts.Temperature
		}
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/xerohard/ai/v2")
	req.Header.Set("X-Title", "unsafe0x0/ai")

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

	return resp.Body, nil
}

func (p *OpenRouterProvider) ParseResponse(body io.Reader, onChunk func(string) error) error {
	return base.ParseJsonStream(body, onChunk)
}
