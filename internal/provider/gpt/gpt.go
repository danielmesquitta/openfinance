package gpt

import "context"

type ChatCompletionOptions struct {
	SystemMessage string
	JSONResponse  bool
}

type ChatCompletionOption func(*ChatCompletionOptions)

func WithSystemMessage(message string) ChatCompletionOption {
	return func(options *ChatCompletionOptions) {
		options.SystemMessage = message
	}
}

func WithJSONResponse() ChatCompletionOption {
	return func(options *ChatCompletionOptions) {
		options.JSONResponse = true
	}
}

type Provider interface {
	CreateChatCompletion(
		ctx context.Context,
		message string,
		options ...ChatCompletionOption,
	) (string, error)
}
