package gpt

type GPTProvider interface {
	CreateChatCompletion(message string) (string, error)
}
