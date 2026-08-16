package openai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	openaisdk "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/danielmesquitta/openfinance-to-sheets/internal/provider/gpt"
)

type chatCompletionRequest struct {
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	ResponseFormat *struct {
		Type string `json:"type"`
	} `json:"response_format"`
}

func newTestClient(serverURL string) *OpenAIClient {
	return &OpenAIClient{client: openaisdk.NewClient(
		option.WithAPIKey("test"),
		option.WithBaseURL(serverURL+"/"),
		option.WithMaxRetries(0),
	)}
}

func writeCompletion(writer http.ResponseWriter, content string) {
	writer.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(writer, `{
		"id":"completion",
		"object":"chat.completion",
		"created":0,
		"model":"gpt-5-mini",
		"choices":[{
			"index":0,
			"message":{"role":"assistant","content":%q},
			"finish_reason":"stop"
		}]
	}`, content)
}

func TestCreateChatCompletionBuildsGenericRequest(t *testing.T) {
	tests := []struct {
		name               string
		message            string
		options            []gpt.ChatCompletionOption
		wantRoles          []string
		wantContents       []string
		wantResponseFormat string
	}{
		{
			name:         "default user message",
			message:      "hello",
			wantRoles:    []string{"user"},
			wantContents: []string{"hello"},
		},
		{
			name:    "system message and JSON response",
			message: `{"input":"value"}`,
			options: []gpt.ChatCompletionOption{
				gpt.WithSystemMessage("return JSON"),
				gpt.WithJSONResponse(),
			},
			wantRoles:          []string{"system", "user"},
			wantContents:       []string{"return JSON", `{"input":"value"}`},
			wantResponseFormat: "json_object",
		},
		{
			name:    "last system message wins",
			message: "hello",
			options: []gpt.ChatCompletionOption{
				gpt.WithSystemMessage("first"),
				gpt.WithSystemMessage("last"),
			},
			wantRoles:    []string{"system", "user"},
			wantContents: []string{"last", "hello"},
		},
		{
			name:         "empty system message is omitted",
			message:      "hello",
			options:      []gpt.ChatCompletionOption{gpt.WithSystemMessage("")},
			wantRoles:    []string{"user"},
			wantContents: []string{"hello"},
		},
		{
			name:         "nil option is ignored",
			message:      "hello",
			options:      []gpt.ChatCompletionOption{nil},
			wantRoles:    []string{"user"},
			wantContents: []string{"hello"},
		},
		{
			name:         "system-only completion",
			options:      []gpt.ChatCompletionOption{gpt.WithSystemMessage("hello")},
			wantRoles:    []string{"system"},
			wantContents: []string{"hello"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requestChannel := make(chan chatCompletionRequest, 1)
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				var completionRequest chatCompletionRequest
				if err := json.NewDecoder(request.Body).Decode(&completionRequest); err != nil {
					t.Errorf("decode request: %v", err)
				}
				requestChannel <- completionRequest
				writeCompletion(writer, "completed")
			}))
			t.Cleanup(server.Close)

			content, err := newTestClient(server.URL).CreateChatCompletion(
				t.Context(),
				test.message,
				test.options...,
			)
			if err != nil {
				t.Fatalf("CreateChatCompletion() error = %v", err)
			}
			if content != "completed" {
				t.Fatalf("content = %q", content)
			}

			completionRequest := <-requestChannel
			if len(completionRequest.Messages) != len(test.wantRoles) {
				t.Fatalf("messages = %#v", completionRequest.Messages)
			}
			for index, message := range completionRequest.Messages {
				if message.Role != test.wantRoles[index] || message.Content != test.wantContents[index] {
					t.Fatalf("message[%d] = %#v", index, message)
				}
			}

			if test.wantResponseFormat == "" {
				if completionRequest.ResponseFormat != nil {
					t.Fatalf("response format = %#v", completionRequest.ResponseFormat)
				}
			} else if completionRequest.ResponseFormat == nil ||
				completionRequest.ResponseFormat.Type != test.wantResponseFormat {
				t.Fatalf("response format = %#v", completionRequest.ResponseFormat)
			}
		})
	}
}

func TestCreateChatCompletionRejectsEmptyMessages(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		writeCompletion(writer, "unexpected")
	}))
	t.Cleanup(server.Close)

	_, err := newTestClient(server.URL).CreateChatCompletion(
		t.Context(),
		"",
		gpt.WithSystemMessage(""),
	)
	if err == nil {
		t.Fatal("CreateChatCompletion() error = nil")
	}
	if requests.Load() != 0 {
		t.Fatalf("requests = %d", requests.Load())
	}
}

func TestCreateChatCompletionRejectsInvalidResponses(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
	}{
		{
			name:       "API error",
			statusCode: http.StatusInternalServerError,
			response:   `{"error":{"message":"unavailable","type":"server_error"}}`,
		},
		{
			name:       "no choices",
			statusCode: http.StatusOK,
			response: `{
				"id":"completion",
				"object":"chat.completion",
				"created":0,
				"model":"gpt-5-mini",
				"choices":[]
			}`,
		},
		{
			name:       "empty content",
			statusCode: http.StatusOK,
			response: `{
				"id":"completion",
				"object":"chat.completion",
				"created":0,
				"model":"gpt-5-mini",
				"choices":[{
					"index":0,
					"message":{"role":"assistant","content":""},
					"finish_reason":"stop"
				}]
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(test.statusCode)
				_, _ = fmt.Fprint(writer, test.response)
			}))
			t.Cleanup(server.Close)

			_, err := newTestClient(server.URL).CreateChatCompletion(t.Context(), "hello")
			if err == nil {
				t.Fatal("CreateChatCompletion() error = nil")
			}
		})
	}
}
