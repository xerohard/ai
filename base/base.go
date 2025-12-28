// shared utilities and base structures for providers

package base

import (
	"context"
	"fmt"
	"io"

	"github.com/xerohard/ai/v2/sdk"
)

type Provider struct {
	APICaller
}

type APICaller interface {
	CallAPI(ctx context.Context, messages []sdk.Message, streamMode bool, opts *sdk.Options) (io.ReadCloser, error)
}

type StreamParser interface {
	ParseResponse(body io.Reader, onChunk func(string) error) error
}

// adds a system prompt to the beginning of the messages
func (p *Provider) AddSystemPrompt(messages []sdk.Message, opts *sdk.Options) []sdk.Message {
	if opts != nil && opts.SystemPrompt != "" {
		if len(messages) == 0 || messages[0].Role != "system" {
			return append([]sdk.Message{{Role: "system", Content: opts.SystemPrompt}}, messages...)
		}
	}
	return messages
}

// creates a completion by calling the API and processing the response
func (p *Provider) CreateCompletion(
	ctx context.Context,
	messages []sdk.Message,
	opts *sdk.Options,
) (*sdk.CompletionResponse, error) {
	messages = p.AddSystemPrompt(messages, opts)

	body, err := p.CallAPI(ctx, messages, false, opts)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	respBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	content, err := ExtractJsonResponse(respBytes)
	if err != nil {
		return &sdk.CompletionResponse{
			Content: string(respBytes),
		}, nil
	}

	return content, nil
}

// creates a streaming completion by calling the API and returning a ReadCloser for the streaming of response
func (p *Provider) CreateCompletionStream(
	ctx context.Context,
	messages []sdk.Message,
	opts *sdk.Options,
) (io.ReadCloser, error) {
	messages = p.AddSystemPrompt(messages, opts)

	body, err := p.CallAPI(ctx, messages, true, opts)
	if err != nil {
		return nil, err
	}

	parser, ok := p.APICaller.(StreamParser)
	if !ok {
		body.Close()
		return nil, fmt.Errorf("streaming not supported by this provider")
	}

	r, w := io.Pipe()

	go func() {
		defer body.Close()
		err := parser.ParseResponse(body, func(chunk string) error {
			_, writeErr := w.Write([]byte(chunk))
			return writeErr
		})
		if err != nil {
			w.CloseWithError(err)
		} else {
			w.Close()
		}
	}()

	return r, nil
}
