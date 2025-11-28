package agent

import (
	"testing"

	"github.com/luillyfe/sourcing-agent/pkg/llm"
)

func TestAnalyzeRequirements(t *testing.T) {
	// Mock response
	mockResp := &llm.Response{
		Content: []llm.ContentBlock{
			{
				Type: "text",
				Text: `
Here are the requirements:
` + "```json" + `
{
  "required_skills": ["Go", "microservices"],
  "experience_level": "senior",
  "locations": ["Lima"],
  "keywords": ["backend"],
  "nice_to_have": ["Docker"]
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

	reqs, err := analyzeRequirements(client, "Find senior Go devs in Lima")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(reqs.RequiredSkills) != 2 {
		t.Errorf("Expected 2 required skills, got %d", len(reqs.RequiredSkills))
	}
	if reqs.ExperienceLevel != "senior" {
		t.Errorf("Expected senior experience, got %s", reqs.ExperienceLevel)
	}
}

func TestGenerateSearchStrategy(t *testing.T) {
	// Mock response
	mockResp := &llm.Response{
		Content: []llm.ContentBlock{
			{
				Type: "text",
				Text: `
` + "```json" + `
{
  "primary_search": {
    "language": "go",
    "location": "lima",
    "min_repos": 10
  },
  "fallback_searches": [],
  "repository_keywords": ["microservices"],
  "profile_filters": {
    "min_followers": 5,
    "bio_keywords": []
  }
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
	reqs := &Requirements{RequiredSkills: []string{"Go"}}

	strategy, err := generateSearchStrategy(client, reqs)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if strategy.PrimarySearch.Language != "go" {
		t.Errorf("Expected language go, got %s", strategy.PrimarySearch.Language)
	}
}
