package openai

import (
	"context"
	"errors"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/provider/gpt"
)

type OpenAIClient struct {
	client openai.Client
}

func NewOpenAIClient(env *config.Env) *OpenAIClient {
	return &OpenAIClient{
		client: openai.NewClient(option.WithAPIKey(env.OpenAIToken)),
	}
}

func (o *OpenAIClient) CreateChatCompletion(
	ctx context.Context,
	message string,
	options ...gpt.ChatCompletionOption,
) (string, error) {
	completionOptions := gpt.ChatCompletionOptions{}
	for _, completionOption := range options {
		if completionOption != nil {
			completionOption(&completionOptions)
		}
	}

	messages := make([]openai.ChatCompletionMessageParamUnion, 0)
	if completionOptions.SystemMessage != "" {
		messages = append(messages, openai.SystemMessage(completionOptions.SystemMessage))
	}
	if message != "" {
		messages = append(messages, openai.UserMessage(message))
	}
	if len(messages) == 0 {
		return "", errors.New("chat completion requires at least one message")
	}

	params := openai.ChatCompletionNewParams{
		Model:           openai.ChatModelGPT5_6Luna,
		ReasoningEffort: openai.ReasoningEffortMedium,
		Messages:        messages,
	}
	if completionOptions.JSONResponse {
		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: new(shared.NewResponseFormatJSONObjectParam()),
		}
	}

	response, err := o.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("create chat completion: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", errors.New("chat completion returned no choices")
	}

	content := response.Choices[0].Message.Content
	if content == "" {
		return "", errors.New("chat completion returned empty content")
	}

	return content, nil
}

var _ gpt.Provider = (*OpenAIClient)(nil)
