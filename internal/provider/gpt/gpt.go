package gpt

import "context"

type Provider interface {
	CreateChatCompletion(ctx context.Context, message string) (string, error)
}
