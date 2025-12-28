// Anthropic provider

package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/xerohard/ai/v2/base"
	"github.com/xerohard/ai/v2/sdk"
)

type AnthropicProvider struct {
	*base.Provider
	APIKey string
}

func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	p := &AnthropicProvider{
		APIKey: apiKey,
	}
	p.Provider = &base.Provider{APICaller: p}
	return p
}

func (p *AnthropicProvider) CallAPI(
	ctx context.Context,
	messages []sdk.Message,
	streamMode bool,
	opts *sdk.Options,
) (io.ReadCloser, error) {
	url := "https://api.anthropic.com/v1/messages"

	chatMessages := []map[string]string{}
	for _, m := range messages {
		chatMessages = append(chatMessages, map[string]string{
			"role":    m.Role,
			"content": m.Content,
		})
	}

	var systemPrompt string
	if len(chatMessages) > 0 && chatMessages[0]["role"] == "system" {
		systemPrompt = chatMessages[0]["content"]
		chatMessages = chatMessages[1:]
	}

	body := map[string]interface{}{
		"system":     systemPrompt,
		"messages":   chatMessages,
		"stream":     streamMode,
		"max_tokens": 1024,
	}
	if opts != nil {
		if opts.Model != "" {
			body["model"] = opts.Model
		}
		if opts.MaxCompletionTokens != 0 {
			body["max_tokens"] = opts.MaxCompletionTokens
		}
		if opts.ReasoningEffort != "" {
			body["reasoning_effort"] = opts.ReasoningEffort
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
	req.Header.Set("x-api-key", p.APIKey)
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

func (p *AnthropicProvider) ParseResponse(body io.Reader, onChunk func(string) error) error {
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
			var evt struct {
				Type  string `json:"type"`
				Delta struct {
					Text string `json:"text"`
				} `json:"delta"`
			}

			if err := json.Unmarshal(line, &evt); err == nil {
				if evt.Type == "content_block_delta" && evt.Delta.Text != "" {
					if err := onChunk(evt.Delta.Text); err != nil {
						return err
					}
				}
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
