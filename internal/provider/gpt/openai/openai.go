package openai

import (
	"context"
	"fmt"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/provider/gpt"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type OpenAIClient struct {
	client openai.Client
}

func NewOpenAIClient(env *config.Env) *OpenAIClient {
	client := openai.NewClient(
		option.WithAPIKey(env.OpenAIToken),
	)

	return &OpenAIClient{
		client: client,
	}
}

func (o *OpenAIClient) CreateChatCompletion(ctx context.Context, message string) (string, error) {
	resp, err := o.client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model: openai.ChatModelGPT5Mini,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(message),
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}

var _ gpt.Provider = (*OpenAIClient)(nil)
