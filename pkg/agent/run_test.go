package agent

import (
	"testing"

	"github.com/luillyfe/sourcing-agent/pkg/github"
	"github.com/luillyfe/sourcing-agent/pkg/llm"
)

type MockLLMClient struct {
	CallAPIFunc func(messages []llm.Message, tools []llm.Tool) (*llm.Response, error)
}

func (m *MockLLMClient) CallAPI(messages []llm.Message, tools []llm.Tool) (*llm.Response, error) {
	return m.CallAPIFunc(messages, tools)
}

func TestRun(t *testing.T) {
	mockLLM := &MockLLMClient{
		CallAPIFunc: func(messages []llm.Message, tools []llm.Tool) (*llm.Response, error) {
			return &llm.Response{
				Content: []llm.ContentBlock{
					{
						Type: "text",
						Text: "Here is a candidate.",
					},
				},
				StopReason: "end_turn",
			}, nil
		},
	}

	mockGithub := &github.Client{} // We don't need a real client for this test as we won't call it

	result, err := Run(mockLLM, mockGithub, "Find Go devs")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	expected := "Here is a candidate."
	if result != expected {
		t.Errorf("Expected result %q, got %q", expected, result)
	}
}
