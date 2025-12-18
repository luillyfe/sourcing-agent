package agent

import (
	"testing"

	"github.com/luillyfe/sourcing-agent/pkg/llm"
)

func TestAnalyzeRequirementsUnclear(t *testing.T) {
	// Mock response for unclear query
	mockResp := &llm.Response{
		Content: []llm.ContentBlock{
			{
				Type: "text",
				Text: `
` + "```json" + `
{
  "required_skills": [],
  "experience_level": "",
  "locations": [],
  "keywords": [],
  "nice_to_have": [],
  "unclear_request": true,
  "clarification_question": "Which programming language are you looking for?"
}
` + "```",
			},
		},
	}

	client := &MockLLMClient{
		CallAPIFunc: func(messages []llm.Message, tools []llm.Tool) (*llm.Response, error) {
			return mockResp, nil
		},
	}

	reqs, _, err := analyzeRequirements(client, "Find senior Go devs in Lima")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !reqs.UnclearRequest {
		t.Error("Expected UnclearRequest to be true")
	}
	if reqs.ClarificationQuestion != "Which programming language are you looking for?" {
		t.Errorf("Expected specific clarification question, got %s", reqs.ClarificationQuestion)
	}
}

func TestRunStage2Unclear(t *testing.T) {
	// Mock response that returns unclear=true
	mockResp := &llm.Response{
		Content: []llm.ContentBlock{
			{
				Type: "text",
				Text: `
` + "```json" + `
{
  "unclear_request": true,
  "clarification_question": "What is the role?"
}
` + "```",
			},
		},
	}

	client := &MockLLMClient{
		CallAPIFunc: func(messages []llm.Message, tools []llm.Tool) (*llm.Response, error) {
			return mockResp, nil
		},
	}

	// githubClient can be nil because it shouldn't be reached
	_, err := RunStage2(client, nil, "bad query")

	if err == nil {
		t.Fatal("Expected error for unclear request, got nil")
	}

	expectedErr := "request unclear: What is the role?"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}
