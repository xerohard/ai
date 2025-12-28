// Mistral provider

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

type MistralProvider struct {
	*base.Provider
	APIKey string
}

func NewMistralProvider(apiKey string) *MistralProvider {
	p := &MistralProvider{
		APIKey: apiKey,
	}
	p.Provider = &base.Provider{APICaller: p}
	return p
}

func (p *MistralProvider) CallAPI(ctx context.Context, messages []sdk.Message, streamMode bool, opts *sdk.Options) (io.ReadCloser, error) {
	url := "https://api.mistral.ai/v1/chat/completions"

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

func (p *MistralProvider) ParseResponse(body io.Reader, onChunk func(string) error) error {
	return base.ParseJsonStream(body, onChunk)
}
