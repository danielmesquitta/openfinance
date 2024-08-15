package gpt

type Provider interface {
	CreateChatCompletion(message string) (string, error)
}
