//  shared utilities for parsing json responses

package base

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"

	"github.com/xerohard/ai/v2/sdk"
)

// parses a streaming JSON response and calls onChunk for each content chunk
func ParseJsonStream(body io.Reader, onChunk func(string) error) error {
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

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}

			if err := json.Unmarshal(line, &chunk); err == nil {
				for _, c := range chunk.Choices {
					if c.Delta.Content != "" {
						if err := onChunk(c.Delta.Content); err != nil {
							return err
						}
					}
				}
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

// extracts a CompletionResponse from a JSON response body
func ExtractJsonResponse(body []byte) (*sdk.CompletionResponse, error) {

	var compResp sdk.CompletionResponse

	if err := json.Unmarshal(body, &compResp); err == nil {
		if compResp.Role != "" || len(compResp.ToolCalls) > 0 {
			return &compResp, nil
		}
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Role      string `json:"role"`
				Content   string `json:"content,omitempty"`
				ToolCalls []struct {
					ID        string `json:"id"`
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"tool_calls,omitempty"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	if len(parsed.Choices) == 0 {
		return &sdk.CompletionResponse{}, nil
	}

	msg := parsed.Choices[0].Message

	// Convert tool calls to SDK format
	toolCalls := make([]sdk.ToolCallRequest, 0, len(msg.ToolCalls))
	for _, tc := range msg.ToolCalls {
		toolCalls = append(toolCalls, sdk.ToolCallRequest{
			ID:        tc.ID,
			Name:      tc.Name,
			Arguments: json.RawMessage(tc.Arguments),
		})
	}

	return &sdk.CompletionResponse{
		Content:   msg.Content,
		ToolCalls: toolCalls,
		Role:      msg.Role,
	}, nil
}
