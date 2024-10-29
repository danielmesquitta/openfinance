package mockgpt

import (
	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/gpt"
)

type MockGPT struct {
	CompletionsByMessage map[string]string
}

func NewMockGPT(completionsByMessage map[string]string) *MockGPT {
	return &MockGPT{
		CompletionsByMessage: completionsByMessage,
	}
}

func (m MockGPT) CreateChatCompletion(message string) (string, error) {
	completion, ok := m.CompletionsByMessage[message]
	if !ok {
		return "", entity.NewErr("message not found")
	}

	return completion, nil
}

var _ gpt.Provider = (*MockGPT)(nil)
