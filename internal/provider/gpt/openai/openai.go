package openai

import (
	"context"
	"fmt"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/provider/gpt"
	"github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	client *openai.Client
}

func NewOpenAIClient(env *config.Env) *OpenAIClient {
	client := openai.NewClient(env.OpenAIAPIToken)

	return &OpenAIClient{
		client: client,
	}
}

func (o *OpenAIClient) CreateChatCompletion(message string) (string, error) {
	resp, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: message,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("chat completion err: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}

var _ gpt.GPTProvider = (*OpenAIClient)(nil)
